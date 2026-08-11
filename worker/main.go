package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dreamfunicular/lightupdavis/worker/queue"
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

func channelHandler() (ch chan queue.UpdateBrightnessEvent, handler func(w http.ResponseWriter, r *http.Request)) {
	ch = make(chan queue.UpdateBrightnessEvent, 100)

	handler = func(w http.ResponseWriter, r *http.Request) {
		var arr []queue.UpdateBrightnessEvent
		var b []byte

		b, err := io.ReadAll(r.Body)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = json.Unmarshal(b, &arr)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for i := range arr {
			ch <- arr[i]
		}
	}

	return ch, handler
}

// func main() {
// ctx, cancel := context.WithCancel(context.Background())

// ch, handler := channelHandler()

// go startServer(ctx, func(w http.ResponseWriter, r *http.Request) { fmt.Println("Received incoming request") })
// Spin up proccess queue that reads from channel

// queue.StartProcessQueue(ctx)
// Create handler function that writes to channel
// Spin up server with handler function
// }
