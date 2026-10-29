package httpengine

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"phforge/internal/audit"
	"phforge/internal/config"
	"phforge/internal/fastcgi"
)

// Server represents the main HTTP web server
type Server struct {
	port      int
	server    *http.Server
	sites     map[string]*SiteRouter
	sitesMu   sync.RWMutex
}

// SiteRouter holds the routing information for a specific site
type SiteRouter struct {
	Config       config.SiteConfig
	FastCGIPorts []int
	ProxyHandler *fastcgi.ProxyHandler
	FileServer   http.Handler
}

// NewServer creates a new web server instance
func NewServer(port int) *Server {
	return &Server{
		port:  port,
		sites: make(map[string]*SiteRouter),
	}
}

// RegisterSite adds or updates a site in the router
func (s *Server) RegisterSite(siteConfig config.SiteConfig, fastcgiPorts []int) {
	s.sitesMu.Lock()
	defer s.sitesMu.Unlock()

	if _, exists := s.sites[siteConfig.Domain]; exists {
		log.Printf("WARNING: Domain %s is already registered on port %d. Overwriting previous configuration!", siteConfig.Domain, s.port)
	}

	s.sites[siteConfig.Domain] = &SiteRouter{
		Config:       siteConfig,
		FastCGIPorts: fastcgiPorts,
		ProxyHandler: fastcgi.NewProxyHandler(fastcgiPorts, siteConfig.DocumentRoot),
		FileServer:   http.FileServer(http.Dir(siteConfig.DocumentRoot)),
	}
	log.Printf("Registered site %s -> %d FastCGI ports\n", siteConfig.Domain, len(fastcgiPorts))
}

// ClearSites removes all registered sites from the router
func (s *Server) ClearSites() {
	s.sitesMu.Lock()
	defer s.sitesMu.Unlock()
	s.sites = make(map[string]*SiteRouter)
}

// loggingResponseWriter captures the HTTP status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *loggingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	// Wrap handler with logging middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		s.ServeHTTP(lw, r)
		
		duration := time.Since(start).String()
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}
		
		audit.LogAccess(ip, r.Method, r.URL.Path, lw.statusCode, duration)
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", s.port),
		Handler: handler,
	}
	
	log.Printf("HTTP Server listening on port %d (0.0.0.0)\n", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server without interrupting any active connections
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// isSensitivePath checks if the requested URL path targets sensitive files or directories
func isSensitivePath(urlPath string) bool {
	// Block Windows NTFS alternate data streams (::$DATA, :stream)
	if strings.Contains(urlPath, ":") {
		return true
	}

	clean := strings.ToLower(filepath.ToSlash(filepath.Clean(urlPath)))
	parts := strings.Split(clean, "/")

	// Check any dotfile or hidden directory (e.g. .env, .git, .editorconfig, .htaccess)
	for _, part := range parts {
		if strings.HasPrefix(part, ".") && part != "." && part != ".." {
			return true
		}
	}

	base := filepath.Base(clean)

	// Block sensitive framework and dependency files
	sensitiveFiles := map[string]bool{
		"composer.json":     true,
		"composer.lock":     true,
		"package.json":      true,
		"package-lock.json": true,
		"artisan":           true,
		"web.config":        true,
		"phpunit.xml":       true,
		"server.php":        true,
		"webpack.mix.js":    true,
		"vite.config.js":    true,
	}
	if sensitiveFiles[base] {
		return true
	}

	// Block dangerous file extensions that should never be served as static files
	ext := filepath.Ext(base)
	dangerousExts := map[string]bool{
		".env":    true,
		".log":    true,
		".sql":    true,
		".sqlite": true,
		".db":     true,
		".bak":    true,
		".ini":    true,
		".conf":   true,
		".yml":    true,
		".yaml":   true,
		".sh":     true,
		".bat":    true,
	}
	if dangerousExts[ext] {
		return true
	}

	return false
}

// ServeHTTP implements the http.Handler interface (VHost routing)
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	// Remove port from host if present
	if idx := strings.IndexByte(host, ':'); idx >= 0 {
		host = host[:idx]
	}

	s.sitesMu.RLock()
	siteRouter, exists := s.sites[host]
	
	// Fallback 1: Catch-all domain wildcard
	if !exists {
		siteRouter, exists = s.sites["*"]
	}
	
	// Fallback 2: If accessed via LAN IP and there is only 1 site on this port, serve it
	if !exists && len(s.sites) == 1 {
		for _, v := range s.sites {
			siteRouter = v
			exists = true
			break
		}
	}
	s.sitesMu.RUnlock()

	if !exists {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Security: Block sensitive files (.env, .git, config files, NTFS streams)
	if isSensitivePath(r.URL.Path) {
		audit.LogEvent("SECURITY", fmt.Sprintf("Blocked access to sensitive path %s (Host: %s, Remote: %s)", r.URL.Path, r.Host, r.RemoteAddr))
		http.Error(w, "403 Forbidden: Access to sensitive file is prohibited", http.StatusForbidden)
		return
	}

	// Try to serve static file
	targetPath := filepath.Join(siteRouter.Config.DocumentRoot, filepath.Clean(r.URL.Path))
	
	// Security: Prevent path traversal outside document root
	cleanDocRoot := filepath.Clean(siteRouter.Config.DocumentRoot)
	cleanTargetPath := filepath.Clean(targetPath)
	if !strings.HasPrefix(cleanTargetPath, cleanDocRoot) {
		audit.LogEvent("SECURITY", fmt.Sprintf("Directory traversal attempt blocked: %s", r.URL.Path))
		http.Error(w, "403 Forbidden: Invalid file path", http.StatusForbidden)
		return
	}

	// Check if the exact file or directory exists on disk
	info, err := os.Stat(targetPath)
	if err == nil {
		if info.IsDir() {
			// Directory requested, check for index.php
			indexPHP := filepath.Join(targetPath, "index.php")
			if _, err := os.Stat(indexPHP); err == nil {
				r.URL.Path = strings.TrimRight(r.URL.Path, "/") + "/index.php"
				siteRouter.ProxyHandler.ServeHTTP(w, r)
				return
			}
			// Directory without index.php, serve via FileServer (might show index.html or 403)
			siteRouter.FileServer.ServeHTTP(w, r)
			return
		}

		// It's a file. If it's a PHP file (case-insensitive check), forward to FastCGI.
		if strings.EqualFold(filepath.Ext(targetPath), ".php") {
			siteRouter.ProxyHandler.ServeHTTP(w, r)
			return
		}

		// Static file, serve it
		siteRouter.FileServer.ServeHTTP(w, r)
		return
	}

	// File does NOT exist on disk.
	// Laravel/Modern PHP Front Controller pattern: Rewrite everything to index.php
	indexPHP := filepath.Join(siteRouter.Config.DocumentRoot, "index.php")
	if _, err := os.Stat(indexPHP); err == nil {
		r.URL.Path = "/index.php"
		siteRouter.ProxyHandler.ServeHTTP(w, r)
		return
	}

	// File not found and no index.php fallback available
	http.NotFound(w, r)
}
