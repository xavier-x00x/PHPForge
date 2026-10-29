package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	"phforge/internal/api"
	"phforge/internal/audit"
	"phforge/internal/config"
	"phforge/internal/httpengine"
	"phforge/internal/procmanager"
)

var (
	elog         debug.Log
	wg           sync.WaitGroup
	coreStopChan chan struct{}
)

type phforgeService struct{}

func (m *phforgeService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	
	coreStopChan = make(chan struct{})
	
	// Start core components
	wg.Add(1)
	go func() {
		defer wg.Done()
		startCore()
	}()
	
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				audit.LogEvent("SYSTEM", "Service Stop/Shutdown requested by Windows Service Manager")
				if coreStopChan != nil {
					close(coreStopChan)
				}
				break loop
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	
	// Wait for startCore to finish cleanup
	wg.Wait()
	
	return
}

func startCore() {
	// Use executable path as base directory + dist
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	baseDir := filepath.Join(filepath.Dir(exePath), "dist")
	sitesConfigPath := filepath.Join(baseDir, `config`, `sites.json`)
	logDir := filepath.Join(baseDir, `logs`)
	
	// Ensure directories exist
	os.MkdirAll(filepath.Join(baseDir, `config`), 0755)
	
	// Initialize Audit Logger
	if err := audit.InitLogger(logDir); err != nil {
		log.Printf("Warning: Failed to initialize audit logger: %v\n", err)
	}
	defer audit.Close()

	audit.LogEvent("SYSTEM", "PHForge Core Service starting up...")

	// Clean up any orphaned PHP workers from previous ungraceful exits
	procmanager.CleanupOrphans(baseDir)

	// Read Server config for ports (optional, hardcoded for MVP if missing)
	apiPort := 8081
	
	// Load Server config
	serverConfigPath := filepath.Join(baseDir, `config`, `server.json`)
	phpWorkers := 16
	apiKey := "phforge-secret-key"
	if srvCfg, err := config.LoadServerConfig(serverConfigPath); err == nil {
		if srvCfg.APIPort > 0 { apiPort = srvCfg.APIPort }
		if srvCfg.PHPWorkers > 0 { phpWorkers = srvCfg.PHPWorkers }
		if srvCfg.APIKey != "" { apiKey = srvCfg.APIKey }
	}

	manager := httpengine.NewManager()
	apiServer := api.NewAPIServer(apiPort, manager, baseDir, sitesConfigPath, apiKey)

	// Load Sites Config dynamically
	sitesCfg, err := config.LoadSitesConfig(sitesConfigPath)
	if err != nil {
		log.Printf("Warning: Failed to load sites.json: %v", err)
		audit.LogEvent("ERROR", fmt.Sprintf("Failed to load sites.json: %v", err))
	} else {
		for _, site := range sitesCfg.Sites {
			// Find PHP executable
			phpBinPath := filepath.Join(baseDir, "php-versions", "php-"+site.PHPVersion, "php-cgi.exe")
			
			// Start PHP worker pool for this site
			ports, err := procmanager.StartWorkerPool(site.PHPVersion, phpBinPath, phpWorkers)
			if err != nil {
				log.Printf("Warning: Failed to start PHP worker pool for %s (PHP %s): %v", site.Domain, site.PHPVersion, err)
				audit.LogEvent("ERROR", fmt.Sprintf("Failed to start PHP worker pool for %s: %v", site.Domain, err))
				continue
			}
			
			// Register site to http manager
			manager.RegisterSite(site, ports)
			audit.LogEvent("CONFIG", fmt.Sprintf("Registered Site %s on %d FastCGI ports (Web Port %d) with PHP %s", site.Domain, len(ports), site.Port, site.PHPVersion))
		}
	}

	// Start API server
	go func() {
		log.Printf("Starting Local REST API on port %d", apiPort)
		if err := apiServer.Start(); err != nil {
			audit.LogEvent("FATAL", fmt.Sprintf("API Server failed: %v", err))
			log.Fatalf("API Server failed: %v", err)
		}
	}()

	// Block main thread to keep background servers running
	log.Println("Core Service started. Servers are running in the background.")
	
	// Graceful shutdown on Ctrl+C, kill, or Windows Service Stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	if coreStopChan != nil {
		select {
		case <-sigChan:
		case <-coreStopChan:
		}
	} else {
		<-sigChan
	}
	
	log.Println("Shutting down Core Service...")
	audit.LogEvent("SYSTEM", "Core Service shutting down...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	manager.ShutdownAll()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down API server: %v", err)
	}

	procmanager.StopAllWorkers()
	log.Println("All PHP workers stopped.")
}

func main() {
	installFlag := flag.Bool("install", false, "Install as Windows Service")
	uninstallFlag := flag.Bool("uninstall", false, "Uninstall Windows Service")
	flag.Parse()

	if *installFlag {
		if err := installService("PHForgeCoreSvc", "PHPForge Core Service"); err != nil {
			log.Fatalf("Failed to install service: %v", err)
		}
		fmt.Println("Service installed successfully.")
		return
	}
	if *uninstallFlag {
		if err := uninstallService("PHForgeCoreSvc"); err != nil {
			log.Fatalf("Failed to uninstall service: %v", err)
		}
		fmt.Println("Service uninstalled successfully.")
		return
	}

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	
	if inService {
		runService("PHForgeCoreSvc", false)
		return
	}
	
	// Console mode
	elog = debug.New("PHForgeCoreSvc")
	log.Println("Running in console mode (not as a service)")
	startCore()
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &phforgeService{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}

func installService(name, desc string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	s, err = m.CreateService(name, exePath, mgr.Config{
		StartType:   mgr.StartAutomatic,
		DisplayName: name,
		Description: desc,
	})
	if err != nil {
		return err
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %v", err)
	}
	return nil
}

func uninstallService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %v", err)
	}
	return nil
}
