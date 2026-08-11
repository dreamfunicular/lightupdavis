package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
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

func makeHandler(cancel context.CancelFunc) (ch chan queue.UpdateBrightnessEvent, handler func(w http.ResponseWriter, r *http.Request)) {
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
			e := arr[i]

			if e.BulbNo == -1 {
				cancel()
				fmt.Println("Bulb was -1, exiting")
				return
			}

			ch <- e
		}
	}

	return ch, handler
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	ch, handler := makeHandler(cancel)

	var wg sync.WaitGroup
	wg.Go(func() { startServer(ctx, handler) })
	cancel()
	wg.Wait()

	select {
	case e := <-ch:
		fmt.Println(e)
	default:
	}
	// Spin up proccess queue that reads from channel

	// queue.StartProcessQueue(ctx)
	// Create handler function that writes to channel
	// Spin up server with handler function
}
