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

	"github.com/dreamfunicular/lightupdavis/worker/queue"
)

func generateUpdateBrightnessEvent(t time.Time, i int) (e queue.UpdateBrightnessEvent) {
	return queue.UpdateBrightnessEvent{
		Time:   t,
		BulbNo: 1,
		Power:  1,
	}
}
func generateArray(start time.Time, delays []int) (q []queue.UpdateBrightnessEvent) {
	q = make([]queue.UpdateBrightnessEvent, 0)

	for i := range delays {
		newTime := start.Add(time.Duration(delays[i]) * time.Millisecond)
		e := generateUpdateBrightnessEvent(newTime, 1)
		q = append(q, e)
	}

	return q
}

func TestGenerateArray(t *testing.T) {
	curr := time.Now()

	expected := []queue.UpdateBrightnessEvent{
		{
			Time:   curr.Add(20 * time.Millisecond),
			BulbNo: 1,
			Power:  1,
		},
		{
			Time:   curr.Add(40 * time.Millisecond),
			BulbNo: 1,
			Power:  1,
		},
	}

	actual := generateArray(curr, []int{20, 40})

	if !reflect.DeepEqual(actual, expected) {
		t.Error("Actual and expected differed")
	}
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

	expected := generateArray(curr, []int{20})

	actual := make([]queue.UpdateBrightnessEvent, 0)

	handler := func(w http.ResponseWriter, r *http.Request) {
		var e queue.UpdateBrightnessEvent
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
				fmt.Println("Expected and actual are off by:", expected[i].Time.Sub(actual[i].Time))
			}
		}

	}

	wg.Wait()
}
