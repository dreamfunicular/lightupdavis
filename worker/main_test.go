package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/dreamfunicular/lightupdavis/worker/queue"
)

func generateUpdateBrightnessEvent(t time.Time) (e queue.UpdateBrightnessEvent) {
	return queue.UpdateBrightnessEvent{
		Time:   t,
		BulbNo: 1,
		Power:  1,
	}
}

func generateArray(start time.Time, delays []int) (q []queue.UpdateBrightnessEvent) {
	q = make([]queue.UpdateBrightnessEvent, 0)

	for i := range delays {
		newTime := start.Add(time.Duration(delays[i]) * time.Millisecond).Truncate(0)
		e := generateUpdateBrightnessEvent(newTime)
		q = append(q, e)
	}

	return q
}

func generateEnv() (context.Context, context.CancelFunc, time.Time, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())

	curr := time.Now().Truncate(0)

	var wg sync.WaitGroup
	return ctx, cancel, curr, &wg
}

func compare(t *testing.T, expected []queue.UpdateBrightnessEvent, actual []queue.UpdateBrightnessEvent) {
	if len(expected) != len(actual) {
		t.Error("Expected array of:", len(expected), "Got array of:", len(actual))
	} else {
		for i := range len(expected) {
			if expected[i].Time != actual[i].Time {
				t.Error("Expected and actual time on element", i, "are off by:", expected[i].Time.Sub(actual[i].Time))
			}

			if expected[i].BulbNo != actual[i].BulbNo {
				t.Error("Expected on element", i, "BulbNo", expected[i].BulbNo, "Got:", actual[i].BulbNo)
			}

			if expected[i].Power != actual[i].Power {
				t.Error("Expected on element", i, "Power", expected[i].BulbNo, "Got:", actual[i].BulbNo)
			}
		}
	}
}

func generateTestHandler() (*[]queue.UpdateBrightnessEvent, func(w http.ResponseWriter, r *http.Request)) {
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

	return &actual, handler
}

func TestGenerateArray(t *testing.T) {
	curr := time.Now()

	expected := []queue.UpdateBrightnessEvent{
		{
			Time:   curr.Add(20 * time.Millisecond).Truncate(0),
			BulbNo: 1,
			Power:  1,
		},
		{
			Time:   curr.Add(40 * time.Millisecond).Truncate(0),
			BulbNo: 1,
			Power:  1,
		},
	}

	actual := generateArray(curr, []int{20, 40})

	if !reflect.DeepEqual(actual, expected) {
		t.Error("Actual and expected differed")
	}
}

// startServer method tests

func TestStartServerPing(t *testing.T) {
	ctx, cancel, _, wg := generateEnv()

	called := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		called = true
	}

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

func testStartServerWithRequests(t *testing.T, delays []int) {
	ctx, cancel, curr, wg := generateEnv()

	expected := generateArray(curr, delays)

	actual, handler := generateTestHandler()

	wg.Go(func() { startServer(ctx, handler) })

	for i := range expected {
		e := expected[i]
		body, err := json.Marshal(e)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		time.Sleep(50 * time.Millisecond)

		http.Post("http://localhost:8080/", "application/json", bytes.NewBuffer(body))

	}

	time.Sleep(50 * time.Millisecond)

	cancel()

	compare(t, expected, *actual)

	wg.Wait()
}

func TestStartServerWithZeroRequests(t *testing.T) {
	delays := []int{}

	testStartServerWithRequests(t, delays)
}

func TestStartServerOneMessage(t *testing.T) {
	delays := []int{20}

	testStartServerWithRequests(t, delays)
}

func TestStartServerWithThreeRequests(t *testing.T) {
	delays := []int{20, 60, 100}

	testStartServerWithRequests(t, delays)
}

// channelHandler unit tests

func testHandlerWithRequests(t *testing.T, delays []int) {
	ch, handler := makeHandler(nil)

	curr := time.Now()

	expected := generateArray(curr, delays)

	for i := range expected {
		b, err := json.Marshal(expected[i : i+1])

		if err != nil {
			t.Errorf("JSON marshal failure")
		}

		r, err := http.NewRequest(
			"POST",
			"https://localhost:8080",
			bytes.NewBuffer(b),
		)

		if err != nil {
			t.Errorf("Request creation")
		}

		w := httptest.NewRecorder()

		handler(w, r)
	}

	var actual []queue.UpdateBrightnessEvent

	close(ch)
	for e := range ch {
		actual = append(actual, e)
	}

	compare(t, expected, actual)
}

func TestHandlerWithZeroRequests(t *testing.T) {
	delays := []int{0}

	testHandlerWithRequests(t, delays)
}

func TestHandlerWithOneRequest(t *testing.T) {
	delays := []int{20}

	testHandlerWithRequests(t, delays)
}

func TestHandlerWithTwoRequests(t *testing.T) {
	delays := []int{20, 60}

	testHandlerWithRequests(t, delays)
}

// May not properly interface with context cancellation for ListenAndServe. Integration testing will be required.
func TestHandlerShutoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ch, handler := makeHandler(cancel)

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

	r, err := http.NewRequest(
		"POST",
		"https://localhost:8080",
		bytes.NewBuffer(b),
	)

	if err != nil {
		t.Errorf("Request creation")
	}

	w := httptest.NewRecorder()

	handler(w, r)

	var actual []queue.UpdateBrightnessEvent

	close(ch)
	for e := range ch {
		actual = append(actual, e)
	}

	expected := []queue.UpdateBrightnessEvent{}

	compare(t, expected, actual)

	select {
	case <-ctx.Done():
	default:
		t.Error("Context wasn't cancelled")
	}
}
