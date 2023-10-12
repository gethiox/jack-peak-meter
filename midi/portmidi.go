package midi

import (
	"fmt"

	"github.com/rakyll/portmidi"
	gomidi "gitlab.com/gomidi/midi"
	driver "gitlab.com/gomidi/portmididrv"
)

type MidiIO interface {
	Initialize() ([]gomidi.In, []gomidi.Out)
	Close() error
}

type PortMidiIO struct{}

func (p *PortMidiIO) Initialize() ([]gomidi.In, []gomidi.Out) {
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
}

func (p *PortMidiIO) Close() error {
	err := portmidi.Terminate()
	if err != nil {
		return fmt.Errorf("Error terminating portmidi: %v", err)
	}

	return nil
}
