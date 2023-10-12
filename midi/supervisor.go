package midi

import (
	"strings"

	"github.com/pkg/errors"
)

type Supervisor struct {
	Device *MidiDevice
}

func NewSupervisor(device *MidiDevice) (*Supervisor, error) {
	sup := &Supervisor{device}

	if device.In != nil {
		err := device.In.Open()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to open midi in device '%s'", device.GetName())
		}
	}

	if device.Out != nil {
		err := device.Out.Open()
		if err != nil {
			if device.In != nil {
				device.In.Close()
			}

			return nil, errors.Wrapf(err, "failed to open midi out device '%s'", device.GetName())
		}
	}

	return sup, nil
}

func (sup *Supervisor) Close() error {
	errs := []string{}

	if sup.Device.In != nil {
		err := sup.Device.In.Close()
		if err != nil {
			errs = append(errs, errors.Wrap(err, "error closing midi in device").Error())
		}
	}

	if sup.Device.Out != nil {
		err := sup.Device.Out.Close()
		if err != nil {
			errs = append(errs, errors.Wrap(err, "error closing midi out device").Error())
		}
	}

	if len(errs) > 0 {
		return errors.Errorf("error closing supervisor for midi device. %s", strings.Join(errs, ". "))
	}

	return nil
}
