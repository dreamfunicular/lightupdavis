package main

import (
	"context"
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

func processQueue(ctx context.Context, handler UpdateBrightnessEventHandler, q []UpdateBrightnessEvent, ch chan UpdateBrightnessEvent) {
	i := 0
	for {
		i++
		for len(q) > 0 {
			e := q[0]

			if e.timestamp.Compare(time.Now()) == -1 {
				q = q[1:]
				continue
			}

			go func() {
				<-time.NewTimer(time.Until(e.timestamp)).C
				handler(e)
			}()
			q = q[1:]
		}

		select {
		case new := <-ch:
			if new.timestamp.Before(time.Now()) {
				time.Sleep(20 * time.Millisecond)
				continue
			} else {
				q = append(q, new)
			}
		case <-ctx.Done():
			fmt.Println(i, "loops")
			return
		}
	}
}
