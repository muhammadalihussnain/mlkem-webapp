// Package server wires up the HTTP server that serves the frontend static files
// and mounts the WebSocket endpoint at /ws.
package server

import (
	"log"
	"net/http"

	"ntt_verification/ws"
)

// defaultAddr is the listen address used when no override is provided.
const defaultAddr = ":8080"

// wsPath is the URL path for the WebSocket endpoint.
const wsPath = "/ws"

// Config holds tuneable server parameters.
type Config struct {
	// Addr is the TCP address to listen on (e.g. ":8080"). Defaults to defaultAddr.
	Addr string
	// StaticDir is the path to the frontend build output directory.
	// When empty, no static files are served.
	StaticDir string
}

// New constructs and returns an *http.Server wired with the WebSocket hub and
// an optional static file handler.
func New(cfg Config) *http.Server {
	if cfg.Addr == "" {
		cfg.Addr = defaultAddr
	}

	hub := ws.NewHub()
	mux := http.NewServeMux()

	// WebSocket endpoint.
	mux.HandleFunc(wsPath, ws.Handler(hub))

	// Static frontend files — served only when a directory is configured.
	if cfg.StaticDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))
	}

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}
}

// ListenAndServe starts the server and blocks until it exits.
func ListenAndServe(cfg Config) {
	srv := New(cfg)
	log.Printf("server: listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
