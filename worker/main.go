package main

import (
	"fmt"
	"time"
)

type updateBrightnessEvent struct {
	timestamp time.Time
	bulbNo int
	brightness float32
}

func updateBrightness(e updateBrightnessEvent) {
	fmt.Println(time.Now(), "- Updating bulb", e.bulbNo, "to brightness", e.brightness)
}

func colorControl(q []updateBrightnessEvent) {
	for ; len(q) > 0; {
		e := q[0]

		if (e.timestamp.Compare(time.Now()) == -1) {
			q = q[1:]
			continue
		}

		diff := e.timestamp.Sub(time.Now())
		
		time.Sleep(diff)
		updateBrightness(e)
		q = q[1:]
	}
}

func main() {
	// colorControl()
	q := []updateBrightnessEvent {
		updateBrightnessEvent {
			timestamp: time.Now().Add(1 * time.Second),
			bulbNo: 1,
			brightness: 1,
		},
		updateBrightnessEvent {
			timestamp: time.Now().Add(2 * time.Second),
			bulbNo: 1,
			brightness: 0.5,
		},
		updateBrightnessEvent {
			timestamp: time.Now().Add(3 * time.Second),
			bulbNo: 1,
			brightness: 0,
		},
		updateBrightnessEvent {
			timestamp: time.Now().Add(4 * time.Second),
			bulbNo: 1,
			brightness: 1,
		},
	}

	colorControl(q)
}