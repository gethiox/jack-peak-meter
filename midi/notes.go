package midi

type MidiNote struct {
	Note     int `json:"note"`
	Velocity int `json:"velocity"`
}
