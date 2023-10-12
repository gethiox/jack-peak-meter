package midi

import (
	"fmt"
	"log"
	"time"

	"github.com/rakyll/portmidi"
	gomidi "gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/writer"

	// driver "gitlab.com/gomidi/midi/testdrv"
	// driver "gitlab.com/gomidi/portmididrv"
	driver "gitlab.com/gomidi/portmididrv"
)

func PlayNotes(out gomidi.Out, notes []MidiNote, duration time.Duration) error {
	var err error

	wr := writer.New(out)
	wr.SetChannel(0)

	for _, note := range notes {
		err = writer.NoteOn(wr, uint8(note.Note), uint8(note.Velocity))
		if err != nil {
			return err
		}
	}

	fmt.Printf("Playing %v for %v seconds\n", notes, duration.Seconds())
	time.Sleep(duration)

	for _, note := range notes {
		err = writer.NoteOff(wr, uint8(note.Note))
		if err != nil {
			return err
		}
	}

	fmt.Printf("Finished %v for %v seconds\n", notes, duration.Seconds())

	return nil
}

func GetMidiDevices() ([]gomidi.In, []gomidi.Out) {
	err := portmidi.Initialize()
	must(err)

	// drv := driver.New("fake midi device")
	drv, err := driver.New()
	must(err)

	// make sure to close all open ports at the end
	defer drv.Close()

	ins, err := drv.Ins()
	must(err)

	outs, err := drv.Outs()
	must(err)

	return ins, outs
	// in, out := ins[0], outs[0]

	// must(in.Open())
	// must(out.Open())

	// defer in.Close()
	// defer out.Close()

	// // the writer we are writing to
	// wr := writer.New(out)

	// log.Printf("Logging midi devices\nin %v\nout %v\n", in.String(), out.String())

	// // to disable logging, pass mid.NoLogger() as option
	// rd := reader.New(
	// 	reader.NoLogger(),
	// 	// write every message to the out port
	// 	reader.Each(func(pos *reader.Position, msg midi.Message) {
	// 		fmt.Printf("got %s: %v\n", msg, msg.Raw())
	// 	}),
	// )

	// // listen for MIDI
	// err = rd.ListenTo(in)
	// must(err)

	// err = writer.NoteOn(wr, 60, 100)
	// must(err)

	// time.Sleep(1)
	// err = writer.NoteOff(wr, 60)

	// must(err)
	// // Output: got channel.NoteOn channel 0 key 60 velocity 100
	// // got channel.NoteOff channel 0 key 60

	// return []MidiDevice{}
}

func must(err error) {
	if err != nil {
		log.Println(err.Error())
		// panic(err.Error())
	}
}
