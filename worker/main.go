package main

import (
	"fmt"
	"sync"
	"time"
)

type UpdateBrightnessEvent struct {
	timestamp  time.Time
	bulbNo     int
	brightness float32
}

func updateBrightness(e UpdateBrightnessEvent) {
	fmt.Println(time.Now(), "- Updating bulb", e.bulbNo, "to brightness", e.brightness)
}

func processQueue(q []UpdateBrightnessEvent, ch chan UpdateBrightnessEvent) {
	for len(q) > 0 {
		e := q[0]

		if e.timestamp.Compare(time.Now()) == -1 {
			q = q[1:]
			continue
		}

		time.Sleep(time.Until(e.timestamp))
		// TODO: Leaves the possibility of a new, more urgent, instruction arriving while the current upcoming instruction is waiting.
		updateBrightness(e)
		q = q[1:]

		// Non-blocking check on if new updates are in the channel
		select {
		case new := <-ch:
			// TODO: Leaves open the possibility
			q = append(q, new)
		default:
			continue
		}
	}
}

func main() {
	var wg sync.WaitGroup

	q := []UpdateBrightnessEvent{
		{
			timestamp:  time.Now().Add(300 * time.Millisecond),
			bulbNo:     1,
			brightness: 1,
		},
		{
			timestamp:  time.Now().Add(500 * time.Millisecond),
			bulbNo:     1,
			brightness: 0.5,
		},
		{
			timestamp:  time.Now().Add(650 * time.Millisecond),
			bulbNo:     1,
			brightness: 0,
		},
		{
			timestamp:  time.Now().Add(720 * time.Millisecond),
			bulbNo:     1,
			brightness: 1,
		},
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() {
		processQueue(q, ch)
	})

	ch <- UpdateBrightnessEvent{
		timestamp:  time.Now().Add(800 * time.Millisecond),
		bulbNo:     1,
		brightness: 0.6,
	}
	ch <- UpdateBrightnessEvent{
		timestamp:  time.Now().Add(900 * time.Millisecond),
		bulbNo:     1,
		brightness: 0.8,
	}
	ch <- UpdateBrightnessEvent{
		timestamp:  time.Now().Add(1000 * time.Millisecond),
		bulbNo:     1,
		brightness: 1.0,
	}

	wg.Wait()
}
