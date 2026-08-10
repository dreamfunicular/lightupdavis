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
		newTime := start.Add(time.Duration(delays[i]) * time.Millisecond)
		e := generateUpdateBrightnessEvent(newTime)
		q = append(q, e)
	}

	return q
}

func setupEnv() (context.Context, context.CancelFunc, time.Time, *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())

	curr := time.Now().Truncate(0)

	var wg sync.WaitGroup
	return ctx, cancel, curr, &wg
}

func compare(t *testing.T, expected []queue.UpdateBrightnessEvent, actual []queue.UpdateBrightnessEvent) {
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
	ctx, cancel, _, wg := setupEnv()

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

func TestOneMessage(t *testing.T) {
	ctx, cancel, curr, wg := setupEnv()

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

	compare(t, expected, actual)

	wg.Wait()
}

func TestThreeMessages(t *testing.T) {
	ctx, cancel, curr, wg := setupEnv()

	expected := generateArray(curr, []int{20, 60, 100})

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

	compare(t, expected, actual)

	wg.Wait()
}
