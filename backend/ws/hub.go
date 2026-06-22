// Package ws implements the WebSocket hub and connection handler for the
// ML-KEM visualisation backend.
package ws

import (
	"github.com/gorilla/websocket"
)

// client wraps a single WebSocket connection and its send channel.
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// hubCmd is an internal command sent to the hub's run loop.
type hubCmd struct {
	kind     hubCmdKind
	c        *client
	msg      []byte
	replyCh  chan int // used by clientCount
}

type hubCmdKind int

const (
	cmdRegister hubCmdKind = iota
	cmdUnregister
	cmdBroadcast
	cmdClientCount
)

// Hub maintains the set of active WebSocket clients and coordinates broadcasts.
// All mutations happen in a single goroutine (run) to eliminate data races.
type Hub struct {
	cmds chan hubCmd
	done chan struct{}
}

// NewHub returns an initialised Hub and starts its internal run loop.
func NewHub() *Hub {
	h := &Hub{
		cmds: make(chan hubCmd, 256),
		done: make(chan struct{}),
	}
	go h.run()
	return h
}

// run is the Hub's single-writer goroutine. All client map mutations happen here.
func (h *Hub) run() {
	clients := make(map[*client]struct{})
	for cmd := range h.cmds {
		switch cmd.kind {
		case cmdRegister:
			clients[cmd.c] = struct{}{}

		case cmdUnregister:
			if _, ok := clients[cmd.c]; ok {
				delete(clients, cmd.c)
				close(cmd.c.send)
			}

		case cmdBroadcast:
			for c := range clients {
				select {
				case c.send <- cmd.msg:
				default:
					// Client is too slow — drop rather than block the broadcaster.
				}
			}

		case cmdClientCount:
			cmd.replyCh <- len(clients)
		}
	}
	close(h.done)
}

// register adds c to the set of active clients.
func (h *Hub) register(c *client) {
	h.cmds <- hubCmd{kind: cmdRegister, c: c}
}

// unregister removes c from the set of active clients and closes its send channel.
func (h *Hub) unregister(c *client) {
	h.cmds <- hubCmd{kind: cmdUnregister, c: c}
}

// Broadcast sends msg to every currently registered client.
func (h *Hub) Broadcast(msg []byte) {
	h.cmds <- hubCmd{kind: cmdBroadcast, msg: msg}
}

// ClientCount returns the number of currently registered clients.
func (h *Hub) ClientCount() int {
	replyCh := make(chan int, 1)
	h.cmds <- hubCmd{kind: cmdClientCount, replyCh: replyCh}
	return <-replyCh
}
