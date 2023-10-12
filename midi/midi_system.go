package midi

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	gomidi "gitlab.com/gomidi/midi"
)

type MidiDevice struct {
	In  gomidi.In
	Out gomidi.Out
}

func (md *MidiDevice) GetName() string {
	if md.In != nil {
		return md.In.String()
	}
	if md.Out != nil {
		return md.Out.String()
	}

	return ""
}

type MidiSystem struct {
	Supervisors map[string]*Supervisor
	midiIO      MidiIO
}

func NewMidiSystem(midiIO MidiIO) (*MidiSystem, error) {
	ms := &MidiSystem{
		Supervisors: map[string]*Supervisor{},
		midiIO:      midiIO,
	}

	devices := map[string]*MidiDevice{}

	ins, outs := midiIO.Initialize()
	for _, in := range ins {
		found := devices[in.String()]
		if found == nil {
			devices[in.String()] = &MidiDevice{
				In: in,
			}
			continue
		}

		if found.In != nil {
			return nil, errors.Errorf("Duplicate midi in instrument '%s'", in.String())
		}

		found.In = in
	}

	for _, out := range outs {
		found := devices[out.String()]
		if found == nil {
			devices[out.String()] = &MidiDevice{
				Out: out,
			}
			continue
		}

		if found.Out != nil {
			return nil, errors.Errorf("Duplicate midi out instrument '%s'", out.String())
		}

		found.Out = out
	}

	for _, device := range devices {
		sup, err := NewSupervisor(device)
		if err != nil {
			ms.Close()
			return nil, errors.Wrapf(err, "error creating supervisor for midi device '%s'", device.GetName())
		}

		ms.Supervisors[device.GetName()] = sup
	}

	return ms, nil
}

func (ms *MidiSystem) String() string {
	deviceStrs := []string{}
	for _, sup := range ms.Supervisors {
		inStr := ""
		if sup.Device.In != nil {
			inStr = "In"
		}

		outStr := ""
		if sup.Device.Out != nil {
			outStr = "Out"
		}

		str := fmt.Sprintf("- '%s' (%s) (%s)", sup.Device.GetName(), inStr, outStr)
		deviceStrs = append(deviceStrs, str)
	}

	return fmt.Sprintf("%d midi supervisors connected:\n%s", len(ms.Supervisors), strings.Join(deviceStrs, "\n"))
}

func (ms *MidiSystem) Close() {
	for _, sup := range ms.Supervisors {
		err := sup.Close()
		if err != nil {
			fmt.Printf("Error closing supervisor %q: %v\n", sup.Device.GetName(), err)
		}
	}

	err := ms.midiIO.Close()
	if err != nil {
		fmt.Printf("Error closing midi IO %v\n", err)
	}
}

func (ms *MidiSystem) GetInputByName(name string) gomidi.In {
	sup := ms.Supervisors[name]
	if sup == nil {
		return nil
	}

	if sup.Device.In != nil {
		return sup.Device.In
	}

	return nil
}

func (ms *MidiSystem) GetOutputByName(name string) gomidi.Out {
	sup := ms.Supervisors[name]
	if sup == nil {
		return nil
	}

	if sup.Device.Out != nil {
		return sup.Device.Out
	}

	return nil
}

func (ms *MidiSystem) ListenForEvents(callback func(name string, midiNumber int)) error {
	for name, sup := range ms.Supervisors {
		if sup.Device.In == nil {
			continue
		}

		err := sup.Device.In.SetListener(func(data []byte, deltaMicroseconds int64) {
			callback(name, int(data[1]))
		})
		if err != nil {
			return err
		}
	}

	return nil
}
