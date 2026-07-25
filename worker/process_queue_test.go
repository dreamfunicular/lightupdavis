package main

import (
	"testing"
	"time"
)

func TestSingleEntryQueue(t *testing.T) {
	q := []UpdateBrightnessEvent{
		{
			timestamp:  time.Now().Add(300 * time.Millisecond),
			bulbNo:     1,
			brightness: .1,
		},
	}

	mockHandler := func(e UpdateBrightnessEvent) {
		if e != q[0] {
			t.Errorf("Unexpected call")
		}
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	processQueue(mockHandler, q, ch)
}
