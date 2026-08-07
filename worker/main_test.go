package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

type UpdateBrightnessEvent struct {
	Timestamp  time.Time
	BulbNo     int
	Brightness float32
}

func generateUpdateBrightnessEvent(t time.Time, i int) (e UpdateBrightnessEvent) {
	return UpdateBrightnessEvent{
		Timestamp:  t,
		BulbNo:     1,
		Brightness: float32(i) * 0.1,
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

func TestPing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	called := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}

	var wg sync.WaitGroup
	wg.Go(func() { startServer(ctx, handler) })

	time.Sleep(50 * time.Millisecond)

	http.Post("http://localhost:8080/", "application/json", nil)

	time.Sleep(50 * time.Millisecond)

	cancel()

	if !called {
		t.Errorf("HTTP ping to server failed")
	}

	wg.Wait()
}

func TestOneMessage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	curr := time.Now().Truncate(0)

	expected := generateQueue(curr, 40, 1)

	actual := make([]UpdateBrightnessEvent, 0)

	handler := func(w http.ResponseWriter, r *http.Request) {
		var e UpdateBrightnessEvent
		var b []byte

		b, err := io.ReadAll(r.Body)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = json.Unmarshal(b, &e)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		actual = append(actual, e)
	}

	var wg sync.WaitGroup
	wg.Go(func() { startServer(ctx, handler) })

	e := expected[0]
	body, err := json.Marshal(e)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	time.Sleep(50 * time.Millisecond)

	http.Post("http://localhost:8080/", "application/json", bytes.NewBuffer(body))

	time.Sleep(50 * time.Millisecond)

	cancel()

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Contents of expected and actual calls to handler differed.")

		if len(expected) != len(actual) {
			fmt.Println("Expected:", len(expected), "Got:", len(actual))
		} else {
			for i := range len(expected) {
				fmt.Println("Expected and actual are off by:", expected[i].Timestamp.Sub(actual[i].Timestamp))
			}
		}

	}

	wg.Wait()
}
