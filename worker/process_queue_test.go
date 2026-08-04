package main

import (
	"context"
	"fmt"
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

		if len(expected) != len(actual) {
			fmt.Println("Expected:", len(expected), "Got:", len(actual))
		} else {
			for i := range len(expected) {
				fmt.Println("Expected and actual are off by:", expected[i].timestamp.Sub(actual[i].timestamp))
			}
		}

	}
}

func TestSingleEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
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

		if len(expected) != len(actual) {
			fmt.Println("Expected:", len(expected), "Got:", len(actual))
		} else {
			for i := range len(expected) {
				fmt.Println("Expected and actual are off by:", expected[i].timestamp.Sub(actual[i].timestamp))
			}
		}

	}
}

func TestTwoEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	curr := time.Now()

	expected := generateQueue(curr, 4, 1)
	expected = append(expected, generateQueue(curr, 44, 1)...)

	actual := make([]UpdateBrightnessEvent, 0)
	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	q := generateQueue(curr, 4, 1)
	q = append(q, generateQueue(curr, 44, 1)...)

	var wg sync.WaitGroup
	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	close(ch)
	wg.Wait()

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

func TestZeroEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
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

		if len(expected) != len(actual) {
			fmt.Println("Expected:", len(expected), "Got:", len(actual))
		} else {
			for i := range len(expected) {
				fmt.Println("Expected and actual are off by:", expected[i].timestamp.Sub(actual[i].timestamp))
			}
		}

	}
}

func TestTwoEntryQueueWithImmediateMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	curr := time.Now()

	expected := generateQueue(curr, 40, 1)
	expected = append(expected, generateQueue(curr, 80, 1)...)
	expected = append(expected, generateQueue(curr, 120, 1)...)
	expected = append(expected, generateQueue(curr, 160, 1)...)

	var actual []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := generateQueue(curr, 40, 1)
	q = append(q, generateQueue(curr, 80, 1)...)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(120)*time.Millisecond), 120)
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(160)*time.Millisecond), 160)

	close(ch)
	wg.Wait()

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

func TestTwoEntryQueueWithDelayedMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	curr := time.Now()

	expected := generateQueue(curr, 40, 1)
	expected = append(expected, generateQueue(curr, 80, 1)...)
	expected = append(expected, generateQueue(curr, 120, 1)...)
	expected = append(expected, generateQueue(curr, 160, 1)...)

	var actual []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)
	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := generateQueue(curr, 40, 1)
	q = append(q, generateQueue(curr, 80, 1)...)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	<-time.NewTimer(time.Millisecond * 50).C
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(120)*time.Millisecond), 120)
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(160)*time.Millisecond), 160)

	close(ch)
	wg.Wait()

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

func TestZeroEntryQueueWithDelayedMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	curr := time.Now()

	expected := generateQueue(curr, 120, 1)
	expected = append(expected, generateQueue(curr, 160, 1)...)

	var actual []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)
	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := generateQueue(curr, 0, 0)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { processQueue(ctx, mockHandler, q, ch) })

	<-time.NewTimer(time.Millisecond * 50).C
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(120)*time.Millisecond), 120)
	ch <- generateUpdateBrightnessEvent(curr.Add(time.Duration(160)*time.Millisecond), 160)

	close(ch)
	wg.Wait()

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
