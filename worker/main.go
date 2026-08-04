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
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	for {
		for len(q) > 0 {
			e := q[0]

			if e.timestamp.Compare(time.Now()) == -1 {
				q = q[1:]
				continue
			}

			go func() {
				timer := time.NewTimer(time.Until(e.timestamp))
				<-timer.C
				handler(e)
			}()
			q = q[1:]
		}

		select {
		case new := <-ch:
			q = append(q, new)
		case <-ctx.Done():
			return
		}
	}
}
