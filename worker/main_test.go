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

func TestOneMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

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

	go runServer(ctx, handler)

	e := expected[0]
	body, err := json.Marshal(e)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	time.Sleep(50 * time.Millisecond)

	http.Post("http://localhost:8080/", "application/json", bytes.NewBuffer(body))

	time.Sleep(50 * time.Millisecond)

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
}
