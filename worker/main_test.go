package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

type UpdateBrightnessEvent struct {
	timestamp  time.Time
	bulbNo     int
	brightness float32
}

func generateUpdateBrightnessEvent(t time.Time, i int) (e UpdateBrightnessEvent) {
	return UpdateBrightnessEvent{
		timestamp:  t,
		bulbNo:     1,
		brightness: float32(i) * 0.1,
	}
}

func generateQueue(curr time.Time, start int, len int) (q []UpdateBrightnessEvent) {
	for i := start; i < start+len; i++ {
		newTime := curr.Add(time.Duration(i) * time.Millisecond)
		new := generateUpdateBrightnessEvent(newTime, i)
		q = append(q, new)
	}

	return q
}

func handlerFixture(actual []UpdateBrightnessEvent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func TestPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	handlerCalled := false

	pingHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}

	go runServer(ctx, pingHandler)

	time.Sleep(20 * time.Millisecond)

	r, err := http.NewRequest("POST", "http://localhost:8080", nil)
	if err != nil {
		t.Error(err)
		os.Exit(1)
	}
	_, err = http.DefaultClient.Do(r)
	if err != nil {
		t.Error(err)
		os.Exit(1)
	}

	time.Sleep(20 * time.Millisecond)

	if !handlerCalled {
		t.Errorf("Failed to ping the server")
	}
}

// func TestOneMessage(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
// 	defer cancel()

// 	expected := make([]UpdateBrightnessEvent, 1)

// 	actual := make([]UpdateBrightnessEvent, 0)

// 	go runServer(ctx, handlerFixture(actual))

// 	time.Sleep(500 * time.Millisecond)

// 	r, err := http.NewRequest("POST", "http://localhost:8080", nil)
// 	if err != nil {
// 		t.Error(err)
// 		os.Exit(1)
// 	}
// 	_, err = http.DefaultClient.Do(r)
// 	if err != nil {
// 		t.Error(err)
// 		os.Exit(1)
// 	}

// 	time.Sleep(500 * time.Millisecond)

// 	if !reflect.DeepEqual(expected, actual) {
// 		t.Error("Contents of expected and actual calls to handler differed.")

// 		if len(expected) != len(actual) {
// 			fmt.Println("Expected:", len(expected), "Got:", len(actual))
// 		} else {
// 			for i := range len(expected) {
// 				fmt.Println("Expected and actual are off by:", expected[i].timestamp.Sub(actual[i].timestamp))
// 			}
// 		}

// 	}
// }
