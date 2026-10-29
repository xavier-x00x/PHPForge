package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"phforge/internal/config"
	"phforge/internal/httpengine"
	"phforge/internal/procmanager"
)

var startTime = time.Now()

// APIServer represents the internal local REST API
type APIServer struct {
	port       int
	manager    *httpengine.ServerManager
	baseDir    string
	configPath string
	apiKey     string
	server     *http.Server
}

// NewAPIServer creates a new API Server instance
func NewAPIServer(port int, manager *httpengine.ServerManager, baseDir string, configPath string, apiKey string) *APIServer {
	return &APIServer{
		port:       port,
		manager:    manager,
		baseDir:    baseDir,
		configPath: configPath,
		apiKey:     apiKey,
	}
}

// Start begins listening for internal API requests
func (s *APIServer) Start() error {
	mux := http.NewServeMux()
	
	// API routes
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/sites", s.handleSites)
	mux.HandleFunc("/api/v1/reload", s.handleReload)
	mux.HandleFunc("/phforge-health", s.handleHealth)

	// API route only listens on localhost for security
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (s *APIServer) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Dummy status
	json.NewEncoder(w).Encode(map[string]string{
		"status": "running",
		"version": "1.0.0-MVP",
	})
}

func (s *APIServer) handleSites(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method == http.MethodGet {
		sites, err := config.LoadSitesConfig(s.configPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(sites)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *APIServer) checkAuth(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	apiKeyHeader := r.Header.Get("X-API-Key")
	
	if authHeader == "Bearer "+s.apiKey || apiKeyHeader == s.apiKey {
		return true
	}
	return false
}

func (s *APIServer) handleReload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Security: Verify API Key
	if !s.checkAuth(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if s.manager != nil {
		srvCfg, _ := config.LoadServerConfig(filepath.Join(s.baseDir, "config", "server.json"))
		phpWorkers := 16
		if srvCfg != nil && srvCfg.PHPWorkers > 0 {
			phpWorkers = srvCfg.PHPWorkers
		}
		
		sitesCfg, err := config.LoadSitesConfig(s.configPath)
		if err == nil {
			// Track which PHP versions are needed in the new config
			neededPHPVersions := make(map[string]bool)
			
			// Clear all existing routing first
			s.manager.ClearSites()
			
			for _, site := range sitesCfg.Sites {
				phpBinPath := filepath.Join(s.baseDir, "php-versions", "php-"+site.PHPVersion, "php-cgi.exe")
				
				// Validate if PHP binary exists before trying to start pool
				if _, err := os.Stat(phpBinPath); os.IsNotExist(err) {
					log.Printf("WARNING: PHP binary not found at %s. Skipping site %s", phpBinPath, site.Domain)
					continue
				}
				
				neededPHPVersions[site.PHPVersion] = true
				
				// This will reuse existing ports if the pool is already running!
				ports, err := procmanager.StartWorkerPool(site.PHPVersion, phpBinPath, phpWorkers)
				if err != nil {
					log.Printf("WARNING: Failed to start worker pool for %s: %v", site.PHPVersion, err)
					continue
				}
				s.manager.RegisterSite(site, ports) // Will overwrite existing routing without downtime
			}
			
			// Clean up unused PHP versions (Zero Downtime!)
			activeVersions := procmanager.GetActivePHPVersions()
			for _, activeVer := range activeVersions {
				if !neededPHPVersions[activeVer] {
					log.Printf("PHP version %s is no longer needed. Stopping worker pool...", activeVer)
					procmanager.StopWorkerPool(activeVer)
				}
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "OK",
		"message": "Core service reloaded successfully",
	})
}

func (s *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Security: Verify API Key
	if !s.checkAuth(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	duration := time.Since(startTime).Truncate(time.Second)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	var uptime string
	if days > 0 {
		uptime = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		uptime = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		uptime = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		uptime = fmt.Sprintf("%ds", seconds)
	}

	healthData := map[string]interface{}{
		"status":              "OK",
		"uptime":              uptime,
		"active_php_workers":  procmanager.GetActiveWorkerCount(),
		"memory_allocated_mb": m.Alloc / 1024 / 1024,
		"memory_sys_mb":       m.Sys / 1024 / 1024,
	}

	json.NewEncoder(w).Encode(healthData)
}
