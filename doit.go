package main

import (
	"fmt"
	"time"

	mid "jack-peak-meter/midi"

	"github.com/pkg/errors"
)

// const volumeThreshold = 10
// const lowThreshold = 5

const volumeThreshold = 60

// const volumeThreshold = 70
const lowThreshold = 40

// if the new note is higher than the previous one, and it's been more than 200 ms since the last increase, this is a peak. if it's also above some threshold
const waitTime = 200 * time.Millisecond

var lastTimePeak *time.Time
var lastTimeValley *time.Time

var hitCount = 0

var t = time.Now()

var storedValues = []float32{}

func processFrames(value float32, percentage int) {
	storedValues = append(storedValues, value)

	// fmt.Println(value)

	now := time.Now()
	if lastTimePeak == nil {
		if percentage >= volumeThreshold {
			lastTimePeak = &now
			hit(percentage)
			hitCount++
			// ptof(hitCount)
		}
		return
	}

	if time.Since(*lastTimePeak) < waitTime {
		return
	}

	if lastTimeValley == nil {
		if percentage < lowThreshold {
			lastTimeValley = &now
		}
		return
	}

	if time.Since(*lastTimeValley) < waitTime {
		return
	}

	diff := lastTimePeak.Sub(*lastTimeValley)
	waitingForValley := diff > 0

	if waitingForValley {
		if percentage < lowThreshold {
			// fmt.Println(percentage)
			lastTimeValley = &now
		}
		return
	}

	if percentage >= volumeThreshold {
		lastTimePeak = &now
		hit(percentage)
		hitCount++
		// ptof(hitCount)
	}
}

func hit(percentage interface{}) {
	go func() {
		err := mid.PlayNotes(midiOut, []mid.MidiNote{
			{
				Note:     50,
				Velocity: 127,
			},
		}, time.Millisecond*10)

		if err != nil {
			fmt.Println(errors.Wrap(err, "error playing midi note"))
		}

		// ptof(percentage)
	}()
}
