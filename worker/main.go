package main

import (
	"context"
	"net/http"
)

func runServer(ctx context.Context, handler func(w http.ResponseWriter, r *http.Request)) {
	http.HandleFunc("/", handler)
	http.ListenAndServe("localhost:8080", nil)
}

func main() {
	// MVP Project
	// Create channel
	// Spin up proccess queue that reads from channel
	// Create handler function that writes to channel
	// Spin up server with handler function
}
