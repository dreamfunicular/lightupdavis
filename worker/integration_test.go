package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/dreamfunicular/lightupdavis/worker/queue"
)

func TestIntegrationShutdown(t *testing.T) {
	var wg sync.WaitGroup

	wg.Go(func() { main() })

	time.Sleep(20 * time.Millisecond)

	curr := time.Now()

	shutoffRequest := []queue.UpdateBrightnessEvent{
		{
			Time:   curr.Add(20 * time.Millisecond),
			BulbNo: -1,
			Power:  1,
		},
	}

	b, err := json.Marshal(shutoffRequest)

	if err != nil {
		t.Errorf("JSON marshal failure")
	}

	http.Post("http://localhost:8080", "application/json", bytes.NewBuffer(b))

	wg.Wait()
}

func TestIntegrationOneRequest(t *testing.T) {
	var wg sync.WaitGroup

	wg.Go(func() { main() })

	time.Sleep(20 * time.Millisecond)

	curr := time.Now()

	requests := []queue.UpdateBrightnessEvent{
		{
			Time:   curr.Add(20 * time.Millisecond),
			BulbNo: 1,
			Power:  1,
		},
		{
			Time:   curr.Add(20 * time.Millisecond),
			BulbNo: -1,
			Power:  1,
		},
	}

	for i := range requests {
		b, err := json.Marshal(requests[i : i+1])

		if err != nil {
			t.Errorf("JSON marshal failure")
		}

		http.Post("http://localhost:8080", "application/json", bytes.NewBuffer(b))

		time.Sleep(30 * time.Millisecond)
	}

	wg.Wait()
}

func TestIntegrationOutOfOrderRequests(t *testing.T) {
	var wg sync.WaitGroup

	wg.Go(func() { main() })

	time.Sleep(20 * time.Millisecond)

	curr := time.Now()

	requests := []queue.UpdateBrightnessEvent{
		{
			Time:   curr.Add(20 * time.Millisecond),
			BulbNo: 1,
			Power:  1,
		},
		{
			Time:   curr.Add(100 * time.Millisecond),
			BulbNo: 1,
			Power:  1,
		},
		{
			Time:   curr.Add(90 * time.Millisecond),
			BulbNo: 1,
			Power:  1,
		},
		{
			Time:   curr.Add(100 * time.Millisecond),
			BulbNo: -1,
			Power:  1,
		},
	}

	for i := range requests {
		b, err := json.Marshal(requests[i : i+1])

		if err != nil {
			t.Errorf("JSON marshal failure")
		}

		http.Post("http://localhost:8080", "application/json", bytes.NewBuffer(b))

		time.Sleep(40 * time.Millisecond)
	}

	wg.Wait()
}
