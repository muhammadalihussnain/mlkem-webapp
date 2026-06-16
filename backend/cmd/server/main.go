package main

import "ntt_verification/server"

func main() {
	server.ListenAndServe(server.Config{
		Addr:      ":8080",
		StaticDir: "../frontend/dist",
	})
}
