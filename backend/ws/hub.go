// Package ws implements the WebSocket hub and connection handler for the
// ML-KEM visualisation backend.
package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

// client wraps a single WebSocket connection and its send channel.
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active WebSocket clients and coordinates broadcasts.
// It is safe for concurrent use.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

// NewHub returns an initialised, empty Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
	}
}

// register adds c to the set of active clients.
func (h *Hub) register(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

// unregister removes c from the set of active clients and closes its send channel.
func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
}

// Broadcast sends msg to every currently registered client.
// Clients whose send buffer is full are silently skipped to avoid blocking.
func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			// Drop the message for this client rather than blocking the broadcaster.
		}
	}
}

// ClientCount returns the number of currently registered clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
