package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unsafe"

	"github.com/xthexder/go-jack"
	"gitlab.com/gomidi/midi"
)

const (
	disableCursor = "\033[?25l"
	enableCursor  = "\033[?25h"
	moveCursorUp  = "\033[F"
)

var (
	counter int
)

var fillBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

type portStrings []string

func (p *portStrings) String() string {
	return strings.Join(*p, ",")
}

func (p *portStrings) Set(value string) error {
	*p = append(*p, value)
	return nil
}

type visualizer struct {
	channels    int     // Amount of input channels
	offset      int     // Number of matching channels to skip over
	buffer      int     // Smoothing graph with last n printed samples, set 1 to disable
	amplifer    float64 // Compensate weak audio signal with this ultimate amplifier value
	printValues bool
	printChnIdx bool
	printNames  bool
	verbose     bool
	portMatches portStrings // Array of port match patterns used for input bindings

	additionalBuffer int
	avg              float32
	avgMain          []float32
	lastValues       [][]float32

	client       *jack.Client
	PortsIn      []*jack.Port
	srcPortNames portStrings
}

func (v *visualizer) Start() error {
	var status int
	var clientName string

	// trying to establish JACK client
	for i := 0; i < 1000; i++ {
		clientName = fmt.Sprintf("spectrum analyser %d", i)
		// v.client, status = jack.ClientOpen(clientName, jack.NullOption)
		v.client, status = jack.ClientOpen(clientName, jack.NoStartServer)
		if status == 0 {
			break
		}
	}
	if status != 0 {
		return fmt.Errorf("failed to initialize client, errcode: %d", status)
	}
	defer v.client.Close()

	// registering JACK callback
	if code := v.client.SetProcessCallback(v.process); code != 0 {
		return fmt.Errorf("failed to set process callback: %d", code)
	}
	v.client.OnShutdown(v.shutdown)

	// Activating client
	if code := v.client.Activate(); code != 0 {
		return fmt.Errorf("failed to activate client: %d", code)
	}

	// find jack input ports
	for i := range v.portMatches {
		foundNames := v.client.GetPorts(v.portMatches[i], "", jack.PortIsOutput)
		if len(foundNames) == 0 {
			return fmt.Errorf("failed to find matching jack ports: %s", v.portMatches[i])
		}
		for n := range foundNames {
			v.srcPortNames = append(v.srcPortNames, foundNames[n])
		}
	}

	// adjust for offset, if any
	if v.offset > 0 {
		if v.offset >= len(v.srcPortNames) {
			return fmt.Errorf("offset exceeds number of matching jack ports: %d >= %d", v.offset, len(v.srcPortNames))
		}
		v.srcPortNames = v.srcPortNames[v.offset:]
	}

	// print warning if # channels < # found
	if v.channels < len(v.srcPortNames) {
		printfe(">> Capturing the first %d channels of %d found <<\r", v.channels, len(v.srcPortNames))
	}

	// registering audio channels inputs and connecting them automatically to system monitor output
	for i := 1; i <= v.channels && i <= len(v.srcPortNames); i++ {
		portName := fmt.Sprintf("input_%d", i)
		port := v.client.PortRegister(portName, jack.DEFAULT_AUDIO_TYPE, jack.PortIsInput, 0)
		v.PortsIn = append(v.PortsIn, port)

		srcPortName := v.srcPortNames[i-1]
		dstPortName := fmt.Sprintf("%s:input_%d", clientName, i)

		code := v.client.Connect(srcPortName, dstPortName)
		if code != 0 {
			return fmt.Errorf("Failed connecting port \"%s\" to \"%s\"\n", srcPortName, dstPortName)
		}
		if v.verbose {
			printfe("connected port \"%s\" to \"%s\"\n", srcPortName, dstPortName)
		}
	}

	interrupted := make(chan bool)

	// signal handler
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("custom shutdown 1")
		v.shutdown()
		interrupted <- true
	}()

	// buffer := int(v.client.GetBufferSize())
	// v.additionalBuffer = v.calculateAdditionalBuffer(buffer)

	<-interrupted

	printe(disableCursor) // disablingCursorblink
	printe("\n")
	return nil
}

func getHighestSpread(samples []jack.AudioSample) jack.AudioSample {
	var winner jack.AudioSample
	for _, s := range samples {
		if s < 0 {
			s = -s
		}

		if s > winner {
			winner = s
		}
	}
	return winner
}

var processCounter = 0

// JACK callback
func (v *visualizer) process(nframes uint32) int {
	counter += 1
	processCounter++
	// fmt.Printf("Process counter: %d\n", processCounter)

	// fmt.Println("length of portin", len(v.PortsIn))

	// for i, port := range v.PortsIn {
	for i, port := range v.PortsIn[0:1] {
		// for i, port := range v.PortsIn[0:1] {
		samples := port.GetBuffer(nframes)
		// fmt.Println(port)

		highest := float32(getHighestSpread(samples))
		highest *= float32(v.amplifer)

		v.avgMain[i] += highest

		if counter >= v.additionalBuffer {
			value := v.avgMain[i] / float32(v.additionalBuffer)
			v.updateCache(value, i)

			termWidth, termHeight := getTermWidthHeight()

			if termHeight < v.channels {
				printfe(">> Not sufficient space for bars <<\r")
			} else {
				v.printBar(v.getAvg(i), termWidth, i)

				if i+1 != v.channels { // do not print newline for last bar
					printe("\n")
				}
				v.avgMain[i] = 0
			}
		}

	}
	if counter >= v.additionalBuffer {
		counter = 0
		for i := 1; i < v.channels; i++ {
			printe(moveCursorUp)
		}
	}

	return 0
}

// JACK callback
func (v *visualizer) shutdown() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	printe(enableCursor + "\n")
	// v.client.Close()
}

func newVisualizer(channels, offset, buffer int, amplifier float64, portMatches portStrings, verbose, printValues, printChnIdx, printNames bool) visualizer {
	var lastValues [][]float32
	var avgMin []float32

	// preparing fixed-size lastValues struct
	for channel := 0; channel < channels; channel++ {
		var tmp []float32
		for frame := 0; frame < buffer; frame++ {
			tmp = append(tmp, 0.0)
		}

		lastValues = append(lastValues, tmp)
		avgMin = append(avgMin, 0.0)
	}

	// set default input ports pattern, if none provided
	if len(portMatches) == 0 {
		portMatches = portStrings{"system:(capture|monitor)_"}
	}

	return visualizer{
		channels,
		offset,
		buffer,
		amplifier,
		printValues,
		printChnIdx,
		printNames,
		verbose,
		portMatches,
		1,
		0.0,
		avgMin,
		lastValues,
		nil,
		[]*jack.Port{},
		portStrings{},
	}
}

func (v *visualizer) updateCache(value float32, channel int) {
	l := v.buffer - 1
	for i := l; i > 0; i-- {
		v.lastValues[channel][i] = v.lastValues[channel][i-1]
	}
	v.lastValues[channel][0] = value
}

func (v *visualizer) getAvg(channel int) float32 {
	var avg float32
	for _, v := range v.lastValues[channel] {
		avg += v
	}
	avg = avg / float32(v.buffer)
	if avg > 1 {
		avg = 1
	}
	return avg
}

func (v *visualizer) calculateAdditionalBuffer(frameSize int) int {
	if frameSize > 512 {
		return 1
	}
	return 512 / frameSize
}

func (v *visualizer) printBar(value float32, width, chanNumber int) {
	var bar = ""
	if v.printValues {
		width -= 10
		bar = fmt.Sprintf(" %.3f |", value)
	} else {
		width -= 4
		bar = " |"
	}

	if v.printNames {
		width -= 26
		bar = fmt.Sprintf(" %25s%s", v.srcPortNames[chanNumber], bar)
	}

	if v.printChnIdx {
		width -= 4
		bar = fmt.Sprintf(" %3d%s", chanNumber, bar)
	}

	bar = "\r" + bar

	fullBlocks := int(float32(width) * value)
	for i := 0; i < fullBlocks; i++ {
		bar += fillBlocks[8] // full block fill
	}

	if fullBlocks < width {
		fillBlockIdx := int((float32(width)*value - float32(fullBlocks)) * 8)
		bar += fillBlocks[fillBlockIdx] // transition block fill
	}

	for i := 0; i <= width-fullBlocks-2; i++ {
		bar += fillBlocks[0] // empty block fill
	}

	printe(bar + "| ")

	processFrames(value, fullBlocks)
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getTermWidthHeight() (x, y int) {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		panic(errno)
	}
	x = int(ws.Col)
	y = int(ws.Row)
	return
}

var midiOut midi.Out

func main() {
	var close func()
	midiOut, close = initMidi()
	defer close()
	// hit()

	// if true {
	// 	return
	// }

	var (
		verbose       *bool
		printValues   *bool
		printChnIdx   *bool
		printNames    *bool
		flagChannels  *int
		flagOffset    *int
		flagBuffer    *int
		flagAmplifier *float64
		portMatches   portStrings
	)

	verbose = flag.Bool("verbose", false, "Print verbose messages for troubleshooting")
	printValues = flag.Bool("values", false, "Print value before each channel of visualizer")
	printChnIdx = flag.Bool("index", false, "Print channel index before each channel of visualizer")
	printNames = flag.Bool("names", false, "Print channel names before each channel of visualizer")

	flagChannels = flag.Int("channels", 2, "Maximum amount of input channels to meter")
	flagOffset = flag.Int("offset", 0, "Number of matching channels to skip over")
	flagBuffer = flag.Int("buffer", 10, "Smoothing graph with last n printed samples, set 1 to disable")
	flagAmplifier = flag.Float64("amplify", 3.5, "Compensate weak audio signal with this ultimate amplifier value")
	flag.Var(&portMatches, "port", "Name or regex pattern matching one or more jack ports.")
	flag.Parse()

	v := newVisualizer(*flagChannels, *flagOffset, *flagBuffer, *flagAmplifier, portMatches, *verbose, *printValues, *printChnIdx, *printNames)
	go v.Start()

	interrupted := make(chan bool)

	// signal handler
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("custom shutdown 2")
		// v.shutdown()
		interrupted <- true
	}()

	// buffer := int(v.client.GetBufferSize())
	// v.additionalBuffer = v.calculateAdditionalBuffer(buffer)

	<-interrupted

	// err := visualizer.Start()
	// if err != nil {
	// 	panic(err)
	// }
	// printlne("Bye!")
}

var printfe = func(s string, toFormat ...interface{}) {
	// if true {
	// 	return
	// }
	fmt.Printf(s, toFormat)
}

var printe = func(s string) {
	if true {
		return
	}
	fmt.Print(s)
}
var printlne = func(s string) {
	// if true {
	// 	return
	// }
	fmt.Println(s)
}
