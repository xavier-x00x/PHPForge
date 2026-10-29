package audit

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	MaxLogSize = 10 * 1024 * 1024 // 10 MB
	MaxBackups = 5
)

var (
	auditFile *os.File
	logger    *log.Logger
	mu        sync.Mutex
	currentDir string
)

// InitLogger initializes the audit logger to write to the specified directory
func InitLogger(logDir string) error {
	mu.Lock()
	defer mu.Unlock()

	currentDir = logDir

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return openLogFile()
}

// openLogFile opens the log file. Caller must hold the mutex.
func openLogFile() error {
	logPath := filepath.Join(currentDir, "audit.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open audit log file: %w", err)
	}

	auditFile = file
	logger = log.New(auditFile, "", log.Ldate|log.Ltime)
	return nil
}

// checkRotation checks if the file exceeds MaxLogSize and rotates if necessary. Caller must hold mutex.
func checkRotation() {
	if auditFile == nil {
		return
	}
	info, err := auditFile.Stat()
	if err != nil {
		return
	}
	if info.Size() >= MaxLogSize {
		auditFile.Close()
		basePath := filepath.Join(currentDir, "audit.log")

		// Shift old backups: .4 -> .5, .3 -> .4, etc.
		for i := MaxBackups - 1; i >= 1; i-- {
			oldPath := fmt.Sprintf("%s.%d", basePath, i)
			newPath := fmt.Sprintf("%s.%d", basePath, i+1)
			if _, err := os.Stat(oldPath); err == nil {
				os.Rename(oldPath, newPath)
			}
		}

		// Rename current to .1
		os.Rename(basePath, basePath+".1")

		openLogFile()
	}
}

// LogEvent writes an event to the audit log
func LogEvent(eventType, message string) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		// Fallback to standard log if not initialized
		log.Printf("[AUDIT] [%s] %s\n", eventType, message)
		return
	}

	checkRotation()
	// Example output: 2026/07/23 10:00:00 [SYSTEM] Service started
	logger.Printf("[%s] %s\n", eventType, message)
}

// LogAccess writes an HTTP access log entry
func LogAccess(ip, method, path string, status int, duration string) {
	mu.Lock()
	defer mu.Unlock()

	if logger == nil {
		return
	}

	checkRotation()
	logger.Printf("[ACCESS] %s | %s %s | %d | %s\n", ip, method, path, status, duration)
}

// Close gracefully closes the audit log file
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if auditFile != nil {
		auditFile.Close()
		auditFile = nil
		logger = nil
	}
}
