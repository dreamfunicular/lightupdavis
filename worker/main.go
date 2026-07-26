package main

import (
	"fmt"
	"time"
)

type UpdateBrightnessEventHandler func(UpdateBrightnessEvent)

type UpdateBrightnessEvent struct {
	timestamp  time.Time
	bulbNo     int
	brightness float32
}

func handleUpdateBrightnessEvent(e UpdateBrightnessEvent) {
	fmt.Println(time.Now(), "- Updating bulb", e.bulbNo, "to brightness", e.brightness)
}

func processQueue(handler UpdateBrightnessEventHandler, q []UpdateBrightnessEvent, ch chan UpdateBrightnessEvent) {
	for len(q) > 0 {
		e := q[0]

		if e.timestamp.Compare(time.Now()) == -1 {
			q = q[1:]
			continue
		}

		time.Sleep(time.Until(e.timestamp))
		// TODO: Leaves the possibility of a new, more urgent, instruction arriving while the current upcoming instruction is waiting.
		handler(e)
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
