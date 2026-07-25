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

		// Non-blocking poll of the channel
		select {
		case new := <-ch:
			// Catches all buffer-created, late, or bugged updates.
			if new.timestamp.Before(time.Now()) {
				continue
			}

			// TODO: Leaves the possibility of new, more urgent, instructions being behind other ones in the queue. Needs to iterate through all additons.
			q = append(q, new)
			for each := range ch {
				q = append(q, each)
			}
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
			brightness: .1,
		},
		{
			timestamp:  time.Now().Add(500 * time.Millisecond),
			bulbNo:     1,
			brightness: 0.2,
		},
		{
			timestamp:  time.Now().Add(650 * time.Millisecond),
			bulbNo:     1,
			brightness: 0.3,
		},
		{
			timestamp:  time.Now().Add(720 * time.Millisecond),
			bulbNo:     1,
			brightness: 0.4,
		},
	}

	ch := make(chan UpdateBrightnessEvent, 100)

	wg.Go(func() {
		processQueue(q, ch)
	})

	ch <- UpdateBrightnessEvent{
		timestamp:  time.Now().Add(800 * time.Millisecond),
		bulbNo:     1,
		brightness: 0.5,
	}
	// ch <- UpdateBrightnessEvent{
	// 	timestamp:  time.Now().Add(900 * time.Millisecond),
	// 	bulbNo:     1,
	// 	brightness: 0.6,
	// }
	// ch <- UpdateBrightnessEvent{
	// 	timestamp:  time.Now().Add(1000 * time.Millisecond),
	// 	bulbNo:     1,
	// 	brightness: 0.7,
	// }

	close(ch)
	wg.Wait()
}
