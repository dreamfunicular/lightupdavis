package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func startServer(ctx context.Context, handler func(w http.ResponseWriter, r *http.Request)) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("Error caused server to fail:", err)
		}
	}()

	<-ctx.Done()
	err := srv.Shutdown(ctx)

	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	// ctx, cancel := context.WithCancel(context.Background())

	// go startServer(ctx, func(w http.ResponseWriter, r *http.Request) { fmt.Println("Received incoming request") })

	// time.Sleep(100 * time.Millisecond)
	// cancel()

	// ch := make(chan queue.UpdateBrightnessEvent, 100)
	// Spin up proccess queue that reads from channel

	// queue.StartProcessQueue(ctx)
	// Create handler function that writes to channel
	// Spin up server with handler function
}
