package main

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func generateUpdateBrightnessEvent(t time.Time, i int) (e UpdateBrightnessEvent) {
	return UpdateBrightnessEvent{
		timestamp:  t,
		bulbNo:     1,
		brightness: float32(i) * 0.1,
	}
}

func TestGenerateUpdateBrightnessEvent(t *testing.T) {
	curr := time.Now()
	expected := UpdateBrightnessEvent{
		timestamp:  curr,
		bulbNo:     1,
		brightness: 0.1,
	}
	actual := generateUpdateBrightnessEvent(curr, 1)

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Unequal!")
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

func TestGenerateQueue(t *testing.T) {
	curr := time.Now()

	expected := []UpdateBrightnessEvent{
		{
			timestamp:  curr.Add(1 * time.Millisecond),
			bulbNo:     1,
			brightness: .1,
		},
		{
			timestamp:  curr.Add(2 * time.Millisecond),
			bulbNo:     1,
			brightness: .2,
		},
		{
			timestamp:  curr.Add(3 * time.Millisecond),
			bulbNo:     1,
			brightness: .3,
		},
	}

	actual := generateQueue(curr, 1, 3)

	if len(expected) != len(actual) {
		t.Errorf("Expected %d calls to handler; got %d.\n", len(expected), len(actual))
	}

	for i := range len(expected) {
		if expected[i].timestamp != actual[i].timestamp {
			t.Errorf("Timestamp mismatch")
		}

		if expected[i].brightness != actual[i].brightness {
			t.Errorf("brightness mismatch")
		}

		if expected[i].bulbNo != actual[i].bulbNo {
			t.Errorf("bulbNo mismatch")
		}
	}
}

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

func TestTwoEntryQueueWithSustainingChannels(t *testing.T) {
	curr := time.Now()
	expectedOrder := []UpdateBrightnessEvent{
		{
			timestamp:  curr.Add(300 * time.Millisecond),
			bulbNo:     1,
			brightness: .3,
		},
		{
			timestamp:  curr.Add(400 * time.Millisecond),
			bulbNo:     1,
			brightness: .4,
		},
		{
			timestamp:  curr.Add(500 * time.Millisecond),
			bulbNo:     1,
			brightness: 0.5,
		},
		{
			timestamp:  curr.Add(600 * time.Millisecond),
			bulbNo:     1,
			brightness: 0.6,
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
			brightness: .3,
		},
		{
			timestamp:  curr.Add(400 * time.Millisecond),
			bulbNo:     1,
			brightness: .4,
		},
	}

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { processQueue(mockHandler, q, ch) })

	ch <- UpdateBrightnessEvent{
		timestamp:  curr.Add(500 * time.Millisecond),
		bulbNo:     1,
		brightness: 0.5,
	}
	ch <- UpdateBrightnessEvent{
		timestamp:  curr.Add(600 * time.Millisecond),
		bulbNo:     1,
		brightness: 0.6,
	}

	close(ch)
	wg.Wait()

	if len(expectedOrder) != len(actualOrder) {
		t.Errorf("Expected %d calls to handler; got %d.\n", len(expectedOrder), len(actualOrder))
	}

	if !reflect.DeepEqual(expectedOrder, actualOrder) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}
