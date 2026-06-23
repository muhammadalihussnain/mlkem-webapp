package main

import (
	"os"

	"ntt_verification/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server.ListenAndServe(server.Config{
		Addr: ":" + port,
	})
}
