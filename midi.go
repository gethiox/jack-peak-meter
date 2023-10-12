package main

import (
	"encoding/json"
	"fmt"

	"gitlab.com/gomidi/midi"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver

	mid "jack-peak-meter/midi"
)

func initMidi() (midi.Out, func()) {
	midiIO := &mid.PortMidiIO{}
	ms, err := mid.NewMidiSystem(midiIO)
	must(err)

	fmt.Println(ms.String())

	return ms.GetOutputByName("IAC Driver Bus 1"), ms.Close
	// return ms.GetOutputByName("Juno USB Midi "), ms.Close
}

var printf = fmt.Printf

func must(err error, strs ...string) {
	if err != nil {
		printf(err.Error())
		if (len(strs)) > 0 {
			printf(strs[0])
		}
		// panic(err.Error())
	}
}

func assert(cond bool, strs ...string) {
	if !cond {
		printf("assertion failed")
		s := ""
		if (len(strs)) > 0 {
			s = strs[0]
		}

		panic(s)
	}
}

func logJSON(o interface{}) {
	b, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		printf("error marshaling json %s", err.Error())
	}

	printf("%T\n%s\n", o, string(b))
}
