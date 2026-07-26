package main

import (
	"reflect"
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
			t.Error("Actual call to handler did not match expectation")
		}
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	processQueue(mockHandler, q, ch)
}

func TestTwoEntryQueue(t *testing.T) {
	curr := time.Now()
	expectedOrder := []UpdateBrightnessEvent{
		{
			timestamp:  curr.Add(300 * time.Millisecond),
			bulbNo:     1,
			brightness: .1,
		},
		{
			timestamp:  curr.Add(400 * time.Millisecond),
			bulbNo:     1,
			brightness: .1,
		},
	}

	var actualOrder []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actualOrder = append(actualOrder, e)
	}

	q := []UpdateBrightnessEvent{
		{
			timestamp:  curr.Add(300 * time.Millisecond),
			bulbNo:     1,
			brightness: .1,
		},
		{
			timestamp:  curr.Add(400 * time.Millisecond),
			bulbNo:     1,
			brightness: .1,
		},
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	processQueue(mockHandler, q, ch)

	if len(expectedOrder) != len(actualOrder) {
		t.Errorf("Expected %d calls to handler; got %d.\n", len(expectedOrder), len(actualOrder))
	}

	if !reflect.DeepEqual(expectedOrder, actualOrder) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}

func TestZeroEntryQueue(t *testing.T) {
	expectedOrder := []UpdateBrightnessEvent{}

	var actualOrder []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actualOrder = append(actualOrder, e)
	}

	q := []UpdateBrightnessEvent{}

	ch := make(chan UpdateBrightnessEvent, 100)

	processQueue(mockHandler, q, ch)

	if len(expectedOrder) != len(actualOrder) {
		t.Errorf("Expected %d calls to handler; got %d.\n", len(expectedOrder), len(actualOrder))
	}

	if !reflect.DeepEqual(expectedOrder, actualOrder) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}
