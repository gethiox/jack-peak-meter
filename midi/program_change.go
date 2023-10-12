package midi

import (
	"fmt"

	gomidi "gitlab.com/gomidi/midi"
)

type ProgramChanger struct {
	MSB     uint8
	LSB     uint8
	Program uint8
}

// 0xB0 - 0x00 - 0x05 (MSB sound bank selection)
// 0xB0 - 0x20 - 0x01 (LSB sound bank selection)
// 0xC0 - 0x02 (sound selection in the current sound bank)

func PerformProgramChange(out gomidi.Out, msb uint8, lsb uint8, program uint8) {
	fmt.Printf("\nPerforming program change\n%v %v %v\n", lsb, lsb, program)

	msbMsg := []byte{
		byte(0xB0),
		byte(0x00),
		byte(msb),
	}
	out.Write(msbMsg)

	lsbMsg := []byte{
		byte(0xB0),
		byte(0x20),
		byte(lsb),
	}
	out.Write(lsbMsg)

	programMsg := []byte{
		byte(0xC0),
		byte(program - 1),
	}
	out.Write(programMsg)

	fmt.Printf("Program changed\n\n")
}
