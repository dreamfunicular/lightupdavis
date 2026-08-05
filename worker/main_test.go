package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
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
		var e UpdateBrightnessEvent
		var b []byte

		r.Body.Read(b)
		json.Unmarshal(b, &e)

		actual = append(actual, e)
	}
}

func TestPing(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	expected := make([]UpdateBrightnessEvent, 0)

	actual := make([]UpdateBrightnessEvent, 0)

	go runServer(ctx, handlerFixture(actual))

	time.Sleep(10 * time.Millisecond)

	http.NewRequest("GET", "localhost:8080", nil)

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Contents of expected and actual calls to handler differed.")

		if len(expected) != len(actual) {
			fmt.Println("Expected:", len(expected), "Got:", len(actual))
		} else {
			for i := range len(expected) {
				fmt.Println("Expected and actual are off by:", expected[i].timestamp.Sub(actual[i].timestamp))
			}
		}

	}
}
