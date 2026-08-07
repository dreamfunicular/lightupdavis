package main

import (
	"context"
	"fmt"
	"net/http"
)

func startServer(ctx context.Context, handler func(w http.ResponseWriter, r *http.Request)) {
	// http.HandleFunc("/", handler)
	// log.Fatal(http.ListenAndServe(":8080", nil))
	srv := &http.Server{Addr: ":8080"}
	http.HandleFunc("/", handler)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println("Error caused server to fail:", err)
		}
	}()

	<-ctx.Done()
	err := srv.Shutdown(ctx)

	if err != nil {
		fmt.Println("error ocurred")
	}
}

// func main() {
// 	ctx, cancel := context.WithCancel(context.Background())

// 	go startServer(ctx, func(w http.ResponseWriter, r *http.Request) { fmt.Println("Received incoming request") })

// 	time.Sleep(100 * time.Millisecond)
// 	cancel()
// 	//		// MVP Project
// 	//		// Create channel
// 	//		// Spin up proccess queue that reads from channel
// 	//		// Create handler function that writes to channel
// 	//		// Spin up server with handler function
// 	//	}
// }
