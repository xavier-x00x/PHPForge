package fastcgi

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/yookoala/gofast"
)

// ProxyHandler represents an HTTP handler that forwards requests to a FastCGI worker
type ProxyHandler struct {
	fastcgiHandler http.Handler
	sem            chan struct{}
}

// roundRobinClientFactory creates a ClientFactory that distributes connections evenly across multiple addresses
func roundRobinClientFactory(addresses []string) gofast.ClientFactory {
	if len(addresses) == 1 {
		return gofast.SimpleClientFactory(gofast.SimpleConnFactory("tcp", addresses[0]))
	}

	factories := make([]gofast.ClientFactory, len(addresses))
	for i, addr := range addresses {
		factories[i] = gofast.SimpleClientFactory(gofast.SimpleConnFactory("tcp", addr))
	}

	var counter uint32
	return func() (gofast.Client, error) {
		// Round-robin selection
		idx := atomic.AddUint32(&counter, 1) % uint32(len(factories))
		return factories[idx]()
	}
}

// NewProxyHandler creates a new handler that forwards to the given FastCGI ports
func NewProxyHandler(ports []int, documentRoot string) *ProxyHandler {
	var addresses []string
	for _, port := range ports {
		addresses = append(addresses, fmt.Sprintf("127.0.0.1:%d", port))
	}
	
	clientFactory := roundRobinClientFactory(addresses)
	
	poolSize := len(ports)
	if poolSize <= 0 {
		poolSize = 1 // Safe fallback
	}
	
	// NewPHPFS correctly routes requests to PHP files in documentRoot
	// It also automatically sets REDIRECT_STATUS=200 which is required by PHP-CGI on Windows
	fastcgiHandler := gofast.NewHandler(
		gofast.NewPHPFS(documentRoot)(gofast.BasicSession),
		clientFactory,
	)

	return &ProxyHandler{
		fastcgiHandler: fastcgiHandler,
		sem:            make(chan struct{}, poolSize),
	}
}

// ServeHTTP implements the http.Handler interface
func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire semaphore token with a timeout to prevent infinite blocking during burst traffic
	select {
	case p.sem <- struct{}{}:
		// Token acquired, proceed to FastCGI handler
		defer func() { <-p.sem }() // Release token when done
		p.fastcgiHandler.ServeHTTP(w, r)
	case <-time.After(30 * time.Second):
		// Timeout reached, return 503 Service Unavailable
		http.Error(w, "Server Too Busy", http.StatusServiceUnavailable)
	}
}
