package httpengine

import (
	"context"
	"log"
	"sync"
	"time"
	"phforge/internal/config"
)

// ServerManager handles multiple Server instances running on different ports
type ServerManager struct {
	servers map[int]*Server
	mu      sync.Mutex
}

// NewManager creates a new ServerManager
func NewManager() *ServerManager {
	return &ServerManager{
		servers: make(map[int]*Server),
	}
}

// RegisterSite registers a site to the appropriate server based on its port.
// If the server for that port doesn't exist, it creates and starts it.
func (m *ServerManager) RegisterSite(siteConfig config.SiteConfig, fastcgiPorts []int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	port := siteConfig.Port
	if port == 0 {
		port = 8080 // Default port fallback
	}

	srv, exists := m.servers[port]
	if !exists {
		srv = NewServer(port)
		m.servers[port] = srv
		
		// Start this new server in the background
		go func(p int, server *Server) {
			if err := server.Start(); err != nil {
				log.Printf("Server on port %d stopped or failed: %v", p, err)
			}
		}(port, srv)
	}

	srv.RegisterSite(siteConfig, fastcgiPorts)
}

// ClearSites clears all registered sites from all servers
func (m *ServerManager) ClearSites() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, srv := range m.servers {
		srv.ClearSites()
	}
}

// ShutdownAll stops all running HTTP servers gracefully
func (m *ServerManager) ShutdownAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for port, srv := range m.servers {
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down server on port %d: %v", port, err)
		} else {
			log.Printf("Successfully shut down HTTP server on port %d", port)
		}
	}
}
