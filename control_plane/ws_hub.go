package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const maxWSConnections = 200

// MetricsHub manages WebSocket connections and broadcasts metrics.
// Single broadcaster pattern prevents N duplicate tickers.
type MetricsHub struct {
	// clients maps connection to TenantID
	clients    map[*websocket.Conn]string
	register   chan registration
	unregister chan *websocket.Conn
	broadcast  chan struct{} // Signal to force broadcast
	mu         sync.RWMutex
	api        *API
}

type registration struct {
	conn     *websocket.Conn
	tenantID string
}

// NewMetricsHub creates a new WebSocket hub.
func NewMetricsHub(api *API) *MetricsHub {
	return &MetricsHub{
		clients:    make(map[*websocket.Conn]string),
		register:   make(chan registration),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan struct{}),
		api:        api,
	}
}

// Run starts the hub's main loop.
func (h *MetricsHub) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case reg := <-h.register:
			h.mu.Lock()
			// Connection cap to prevent overload
			if len(h.clients) >= maxWSConnections {
				h.mu.Unlock()
				reg.conn.Close()
				log.Printf("WebSocket connection rejected: max connections (%d) reached", maxWSConnections)
				continue
			}
			h.clients[reg.conn] = reg.tenantID
			h.mu.Unlock()
			log.Printf("WebSocket client registered for tenant %s. Total: %d", reg.tenantID, len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered. Total: %d", len(h.clients))

		case <-ticker.C:
			h.broadcastAll(ctx)
		}
	}
}

// broadcastAll fetches metrics for each tenant and sends to respective clients.
func (h *MetricsHub) broadcastAll(ctx context.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Identify active tenants
	tenants := make(map[string]bool)
	for _, tenantID := range h.clients {
		tenants[tenantID] = true
	}

	for tenantID := range tenants {
		metrics, err := h.api.dashboardService.GetDashboardMetrics(ctx, tenantID)
		if err != nil {
			log.Printf("Failed to collect metrics for tenant %s: %v", tenantID, err)
			continue
		}

		// Send to clients of this tenant
		for conn, tid := range h.clients {
			if tid == tenantID {
				// Set write deadline to prevent blocking on dead connections
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteJSON(metrics); err != nil {
					log.Printf("WebSocket write error: %v", err)
					// Unregister will be handled by read pump or next ping
					go h.Unregister(conn)
				}
			}
		}
	}
}

// shutdown gracefully closes all client connections.
func (h *MetricsHub) shutdown() {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Printf("Shutting down WebSocket hub with %d clients", len(h.clients))

	for conn := range h.clients {
		conn.Close()
	}
	h.clients = make(map[*websocket.Conn]string)
}

// Register adds a new client connection.
func (h *MetricsHub) Register(conn *websocket.Conn, tenantID string) {
	h.register <- registration{conn: conn, tenantID: tenantID}
}

// Unregister removes a client connection.
func (h *MetricsHub) Unregister(conn *websocket.Conn) {
	h.unregister <- conn
}

// ClientCount returns the number of connected clients.
func (h *MetricsHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
