package queue

import (
	"context"
	"fmt"
	"time"
)

type UpdateBrightnessEventHandler func(UpdateBrightnessEvent)

type UpdateBrightnessEvent struct {
	Time   time.Time
	BulbNo int
	Power  float32
}

func handleUpdateBrightnessEvent(e UpdateBrightnessEvent) {
	fmt.Println(time.Now(), "- Updating bulb", e.BulbNo, "to brightness", e.Power)
}

func processQueue(ctx context.Context, handler UpdateBrightnessEventHandler, q []UpdateBrightnessEvent, ch chan UpdateBrightnessEvent) {
	i := 0
	for {
		i++
		for len(q) > 0 {
			e := q[0]

			if e.Time.Compare(time.Now()) == -1 {
				q = q[1:]
				continue
			}

			go func() {
				<-time.NewTimer(time.Until(e.Time)).C
				handler(e)
			}()
			q = q[1:]
		}

		select {
		case new := <-ch:
			if new.Time.Before(time.Now()) {
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
