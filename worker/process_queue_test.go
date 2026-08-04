package main

import (
	"context"
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

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}

func TestSingleEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	curr := time.Now()
	expected := generateQueue(curr, 1, 1)
	actual := make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := generateQueue(curr, 1, 1)
	ch := make(chan UpdateBrightnessEvent, 100)

	var wg sync.WaitGroup
	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	close(ch)
	wg.Wait()

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}

func TestTwoEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	curr := time.Now()
	expected := generateQueue(curr, 1, 2)
	actual := make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := generateQueue(curr, 1, 2)
	ch := make(chan UpdateBrightnessEvent, 100)

	var wg sync.WaitGroup
	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	close(ch)
	wg.Wait()

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}

func TestZeroEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	expected := []UpdateBrightnessEvent{}
	actual := make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := []UpdateBrightnessEvent{}

	ch := make(chan UpdateBrightnessEvent, 100)

	processQueue(ctx, mockHandler, q, ch)

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}

func TestTwoEntryQueueWithImmediateMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	curr := time.Now()
	expected := generateQueue(curr, 1, 4)

	var actualOrder []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actualOrder = append(actualOrder, e)
	}

	q := generateQueue(curr, 1, 2)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(3)*time.Millisecond), 3)
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(4)*time.Millisecond), 4)

	close(ch)
	wg.Wait()

	if !reflect.DeepEqual(expected, actualOrder) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}

func TestTwoEntryQueueWithDelayedMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	curr := time.Now()
	expected := generateQueue(curr, 1, 2)
	expected = append(expected, generateQueue(curr, 1005, 2)...)

	var actualOrder []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actualOrder = append(actualOrder, e)
	}

	q := generateQueue(curr, 1, 2)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	<-time.NewTimer(500 * time.Millisecond).C
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(1005)*time.Millisecond), 1005)
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(1006)*time.Millisecond), 1006)

	close(ch)
	wg.Wait()

	if !reflect.DeepEqual(expected, actualOrder) {
		t.Error("Contents of expected and actual calls to handler differed.")
	}
}
