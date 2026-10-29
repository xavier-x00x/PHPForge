package procmanager

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Worker represents a single managed php-cgi process
type Worker struct {
	PHPVersion string
	Port       int
	Cmd        *exec.Cmd
	mu         sync.Mutex
	isRunning  bool
	stopChan   chan struct{}
	pidFile    string
}

// WorkerGroup represents a pool of php-cgi processes for a specific PHP version
type WorkerGroup struct {
	PHPVersion string
	Ports      []int
	Workers    []*Worker
}

// Global active worker groups mapped by PHPVersion
var (
	activeGroups = make(map[string]*WorkerGroup)
	workerMutex  sync.Mutex
)

var (
	lastAllocatedPort = 9000
	portAllocMutex    sync.Mutex
)

// getFreePort asks the kernel for a free open port that is ready to use,
// and ensures we don't hand out the same port twice during rapid concurrent startups.
func getFreePort() (int, error) {
	portAllocMutex.Lock()
	defer portAllocMutex.Unlock()

	for i := 0; i < 10000; i++ {
		lastAllocatedPort++
		if lastAllocatedPort > 65000 {
			lastAllocatedPort = 9000
		}

		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", lastAllocatedPort))
		if err != nil {
			continue
		}

		// Check if OS actually considers it free
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			continue
		}
		
		l.Close()
		return lastAllocatedPort, nil
	}
	
	return 0, fmt.Errorf("no free ports available")
}

// StartWorkerPool launches multiple php-cgi processes for a specific PHP version and returns their ports.
// If a pool for the same PHP version is already active, it reuses the existing ports.
func StartWorkerPool(phpVersion, phpBinPath string, count int) ([]int, error) {
	workerMutex.Lock()
	if group, exists := activeGroups[phpVersion]; exists {
		workerMutex.Unlock()
		return group.Ports, nil
	}
	workerMutex.Unlock()

	group := &WorkerGroup{
		PHPVersion: phpVersion,
	}

	for i := 0; i < count; i++ {
		port, err := getFreePort()
		if err != nil {
			log.Printf("Failed to get free port for worker pool %s: %v\n", phpVersion, err)
			continue
		}

		worker := &Worker{
			PHPVersion: phpVersion,
			Port:       port,
			stopChan:   make(chan struct{}),
			pidFile:    filepath.Join(filepath.Dir(filepath.Dir(phpBinPath)), "..", "pids", fmt.Sprintf("worker-%d.pid", port)),
		}
		
		// ensure pids directory exists
		os.MkdirAll(filepath.Dir(worker.pidFile), 0755)

		group.Ports = append(group.Ports, port)
		group.Workers = append(group.Workers, worker)

		go worker.run(phpBinPath)
	}

	if len(group.Ports) == 0 {
		return nil, fmt.Errorf("failed to start any workers for PHP %s", phpVersion)
	}

	workerMutex.Lock()
	activeGroups[phpVersion] = group
	workerMutex.Unlock()

	// Wait briefly to ensure they started
	time.Sleep(500 * time.Millisecond)

	return group.Ports, nil
}

// StopWorkerPool stops all workers for a given PHP version
func StopWorkerPool(phpVersion string) {
	workerMutex.Lock()
	group, exists := activeGroups[phpVersion]
	if exists {
		delete(activeGroups, phpVersion)
	}
	workerMutex.Unlock()

	if exists {
		for _, w := range group.Workers {
			w.Stop()
		}
	}
}

// StopAllWorkers stops all running php-cgi workers
func StopAllWorkers() {
	workerMutex.Lock()
	var list []*WorkerGroup
	for _, g := range activeGroups {
		list = append(list, g)
	}
	activeGroups = make(map[string]*WorkerGroup)
	workerMutex.Unlock()

	for _, g := range list {
		for _, w := range g.Workers {
			w.Stop()
		}
	}
}

// run starts the process and restarts it if it crashes
func (w *Worker) run(phpBinPath string) {
	crashCount := 0
	var lastCrash time.Time

	for {
		w.mu.Lock()
		// Start php-cgi on the specific port using the -b flag
		bindAddr := fmt.Sprintf("127.0.0.1:%d", w.Port)
		w.Cmd = exec.Command(phpBinPath, "-b", bindAddr)
		
		// CRITICAL for Windows FastCGI: we disable children because it doesn't work.
		// Instead we rely on the Go side process manager.
		w.Cmd.Env = append(os.Environ(),
			"PHP_FCGI_MAX_REQUESTS=10000",
			"PHP_FCGI_CHILDREN=0",
		)
		
		err := w.Cmd.Start()
		if err != nil {
			log.Printf("Failed to start PHP worker (%s): %v\n", w.PHPVersion, err)
			w.mu.Unlock()
			time.Sleep(5 * time.Second) // backoff before retry
			continue
		}
		w.isRunning = true
		
		// Write PID file
		if w.Cmd.Process != nil {
			os.WriteFile(w.pidFile, []byte(strconv.Itoa(w.Cmd.Process.Pid)), 0644)
		}
		w.mu.Unlock()

		log.Printf("PHP Worker (%s) started on %s\n", w.PHPVersion, bindAddr)

		// Wait for the process to exit
		err = w.Cmd.Wait()
		
		w.mu.Lock()
		w.isRunning = false
		os.Remove(w.pidFile)
		w.mu.Unlock()

		// Check if we were told to stop
		select {
		case <-w.stopChan:
			log.Printf("PHP Worker (%s) on %s stopped intentionally.\n", w.PHPVersion, bindAddr)
			return
		default:
			// Process exited, check if it was graceful (e.g. reached max requests)
			if err == nil {
				log.Printf("PHP Worker (%s) on %s reached max requests and exited gracefully. Restarting...\n", w.PHPVersion, bindAddr)
				time.Sleep(1 * time.Second) // Small delay before restart
				continue
			}

			// Crashed, implement Circuit Breaker
			now := time.Now()
			if now.Sub(lastCrash) < 10*time.Second {
				crashCount++
			} else {
				crashCount = 1
			}
			lastCrash = now

			if crashCount >= 3 {
				log.Printf("CIRCUIT BREAKER TRIGGERED: PHP Worker (%s) on %s crashed 3 times rapidly. Suspending for 60s...\n", w.PHPVersion, bindAddr)
				time.Sleep(60 * time.Second)
				crashCount = 0 // Reset after cooldown
			} else {
				log.Printf("PHP Worker (%s) on %s exited (err: %v). Restarting in 1s...\n", w.PHPVersion, bindAddr, err)
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// Stop signals the worker loop to stop and kills the process
func (w *Worker) Stop() {
	close(w.stopChan)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.isRunning && w.Cmd != nil && w.Cmd.Process != nil {
		w.Cmd.Process.Kill()
	}
}

// GetActiveWorkerCount returns the total number of currently managed php-cgi processes
func GetActiveWorkerCount() int {
	workerMutex.Lock()
	defer workerMutex.Unlock()
	count := 0
	for _, g := range activeGroups {
		count += len(g.Workers)
	}
	return count
}

// GetActivePHPVersions returns a list of all PHP versions currently running
func GetActivePHPVersions() []string {
	workerMutex.Lock()
	defer workerMutex.Unlock()
	var versions []string
	for v := range activeGroups {
		versions = append(versions, v)
	}
	return versions
}

// CleanupOrphans reads all pid files in dist/pids and kills the processes if they exist
func CleanupOrphans(baseDir string) {
	pidsDir := filepath.Join(baseDir, "pids")
	files, err := os.ReadDir(pidsDir)
	if err != nil {
		return // directory doesn't exist or can't be read, nothing to clean up
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".pid") {
			pidPath := filepath.Join(pidsDir, file.Name())
			data, err := os.ReadFile(pidPath)
			if err == nil {
				pidStr := strings.TrimSpace(string(data))
				_, err := strconv.Atoi(pidStr)
				if err == nil {
					// Use taskkill with an ImageName filter to ENSURE we only kill php-cgi.exe
					// This prevents accidentally killing a recycled PID that now belongs to another app
					cmd := exec.Command("taskkill", "/F", "/PID", pidStr, "/FI", "IMAGENAME eq php-cgi.exe")
					cmd.Run() // Ignore errors, it might already be dead
				}
			}
			// Clean up the stale pid file
			os.Remove(pidPath)
		}
	}
}

