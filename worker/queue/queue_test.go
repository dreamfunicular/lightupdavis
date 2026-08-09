package queue

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func GenerateUpdateBrightnessEvent(t time.Time, i int) (e UpdateBrightnessEvent) {
	return UpdateBrightnessEvent{
		Time:   t,
		BulbNo: 1,
		Power:  float32(i) * 0.1,
	}
}

func TestGenerateUpdateBrightnessEvent(t *testing.T) {
	curr := time.Now()
	expected := UpdateBrightnessEvent{
		Time:   curr,
		BulbNo: 1,
		Power:  0.1,
	}
	actual := GenerateUpdateBrightnessEvent(curr, 1)

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Unequal!")
	}
}

func GenerateQueue(curr time.Time, start int, len int) (q []UpdateBrightnessEvent) {
	for i := start; i < start+len; i++ {
		newTime := curr.Add(time.Duration(i) * time.Millisecond)
		new := GenerateUpdateBrightnessEvent(newTime, i)
		q = append(q, new)
	}

	return q
}

func TestGenerateQueue(t *testing.T) {
	curr := time.Now()

	expected := []UpdateBrightnessEvent{
		{
			Time:   curr.Add(1 * time.Millisecond),
			BulbNo: 1,
			Power:  .1,
		},
		{
			Time:   curr.Add(2 * time.Millisecond),
			BulbNo: 1,
			Power:  .2,
		},
		{
			Time:   curr.Add(3 * time.Millisecond),
			BulbNo: 1,
			Power:  .3,
		},
	}

	actual := GenerateQueue(curr, 1, 3)

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

func TestSingleEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()

	curr := time.Now()
	expected := GenerateQueue(curr, 1, 1)
	actual := make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := GenerateQueue(curr, 1, 1)
	ch := make(chan UpdateBrightnessEvent, 100)

	var wg sync.WaitGroup
	wg.Go(func() { StartProcessQueue(ctx, mockHandler, q, ch) })

	close(ch)
	wg.Wait()

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

func TestTwoEntryQueue(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	curr := time.Now()

	expected := GenerateQueue(curr, 4, 1)
	expected = append(expected, GenerateQueue(curr, 44, 1)...)

	actual := make([]UpdateBrightnessEvent, 0)
	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	q := GenerateQueue(curr, 4, 1)
	q = append(q, GenerateQueue(curr, 44, 1)...)

	var wg sync.WaitGroup
	wg.Go(func() { StartProcessQueue(ctx, mockHandler, q, ch) })

	close(ch)
	wg.Wait()

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

	StartProcessQueue(ctx, mockHandler, q, ch)

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

func TestTwoEntryQueueWithImmediateMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	curr := time.Now()

	expected := GenerateQueue(curr, 40, 1)
	expected = append(expected, GenerateQueue(curr, 80, 1)...)
	expected = append(expected, GenerateQueue(curr, 120, 1)...)
	expected = append(expected, GenerateQueue(curr, 160, 1)...)

	var actual []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)

	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := GenerateQueue(curr, 40, 1)
	q = append(q, GenerateQueue(curr, 80, 1)...)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { StartProcessQueue(ctx, mockHandler, q, ch) })

	ch <- GenerateUpdateBrightnessEvent(curr.Add(time.Duration(120)*time.Millisecond), 120)
	ch <- GenerateUpdateBrightnessEvent(curr.Add(time.Duration(160)*time.Millisecond), 160)

	close(ch)
	wg.Wait()

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

func TestTwoEntryQueueWithDelayedMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	curr := time.Now()

	expected := GenerateQueue(curr, 40, 1)
	expected = append(expected, GenerateQueue(curr, 80, 1)...)
	expected = append(expected, GenerateQueue(curr, 120, 1)...)
	expected = append(expected, GenerateQueue(curr, 160, 1)...)

	var actual []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)
	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := GenerateQueue(curr, 40, 1)
	q = append(q, GenerateQueue(curr, 80, 1)...)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { StartProcessQueue(ctx, mockHandler, q, ch) })

	<-time.NewTimer(time.Millisecond * 50).C
	ch <- GenerateUpdateBrightnessEvent(curr.Add(time.Duration(120)*time.Millisecond), 120)
	ch <- GenerateUpdateBrightnessEvent(curr.Add(time.Duration(160)*time.Millisecond), 160)

	close(ch)
	wg.Wait()

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

func TestZeroEntryQueueWithDelayedMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	curr := time.Now()

	expected := GenerateQueue(curr, 120, 1)
	expected = append(expected, GenerateQueue(curr, 160, 1)...)

	var actual []UpdateBrightnessEvent = make([]UpdateBrightnessEvent, 0)
	mockHandler := func(e UpdateBrightnessEvent) {
		actual = append(actual, e)
	}

	q := GenerateQueue(curr, 0, 0)

	var wg sync.WaitGroup
	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() { StartProcessQueue(ctx, mockHandler, q, ch) })

	<-time.NewTimer(time.Millisecond * 50).C
	ch <- GenerateUpdateBrightnessEvent(curr.Add(time.Duration(120)*time.Millisecond), 120)
	ch <- GenerateUpdateBrightnessEvent(curr.Add(time.Duration(160)*time.Millisecond), 160)

	close(ch)
	wg.Wait()

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
