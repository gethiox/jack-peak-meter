package main

import (
	"fmt"
	"github.com/xthexder/go-jack"
	"os"
	"syscall"
	"unsafe"
	"flag"
	"os/signal"
)

const channels = 2             // Amount of input channels
const channelBuffer = 10       // Smoothing graph with last n printed samples, set 1 to disable
const arbitraryAmplifier = 3.5 // Compensate weak audio signal with this ultimate amplifier value

const disableCursor = "\033[?25l"
const enableCursor = "\033[?25h"
const moveCursorUp = "\033[F"

var additionalBuffer int

var bufferSize int
var PortsIn []*jack.Port
var client *jack.Client
var avgMain [channels]float32
var counter int
var avg float32

var printTitle *bool
var printValues *bool

func process(nframes uint32) int {
	counter += 1
	for i, port := range PortsIn {
		samples := port.GetBuffer(nframes)

		avg = 0
		for _, sample := range samples {
			if sample < 0 {
				avg = avg - float32(sample)
			} else {
				avg = avg + float32(sample)
			}
		}
		avg = avg * arbitraryAmplifier
		avg = float32(avg) / float32(bufferSize)
		
		avgMain[i] += avg
		
		if counter >= additionalBuffer {
			printBar(avgMain[i]/float32(additionalBuffer), i, int(getWidth()))
			avgMain[i] = 0
		}
		
		fmt.Print("\n")
	}
	if counter >= additionalBuffer {
		counter = 0
	}
	for i := 0; i < channels; i++ {
		fmt.Print(moveCursorUp)
	}
	return 0
}

func shutdown() {
	fmt.Print(enableCursor + "\n")
	client.Close()
}

func main() {
	printTitle = flag.Bool("title", false, "Print name of my ultimate visualizer")
	printValues = flag.Bool("values", false, "Print value before each channel of visualizer")
	flag.Parse()
	
	var status int
	var clientName string
	
	for i := 0; i < 10; i++ {
		clientName = fmt.Sprintf("spectrum analyser %d", i)
		client, status = jack.ClientOpen(clientName, jack.NoStartServer)
		if status == 0 {
			break
		}
	}
	if status != 0 {
		fmt.Println("Status:", status)
		return
	}
	
	defer client.Close()
	
	fmt.Print(disableCursor)
	bufferSize = int(client.GetBufferSize())
	additionalBuffer = calculateAdditionalBuffer(bufferSize)
	
	if code := client.SetProcessCallback(process); code != 0 {
		fmt.Println("Failed to set process callback:", code)
		return
	}
	client.OnShutdown(shutdown)
	
	//client.SetBufferSizeCallback(foo)
	
	if code := client.Activate(); code != 0 {
		fmt.Println("Failed to activate client:", code)
		return
	}
	
	for i := 0; i < channels; i++ {
		port := client.PortRegister(fmt.Sprintf("input_%d", i), jack.DEFAULT_AUDIO_TYPE, jack.PortIsInput, 0)
		if i%2 == 0 {
			client.Connect("system:monitor_1", fmt.Sprintf("%s:input_%d", clientName, i))
		} else {
			client.Connect("system:monitor_2", fmt.Sprintf("%s:input_%d", clientName, i))
		}
		PortsIn = append(PortsIn, port)
	}
	
	fmt.Print("\n\n")
	terminalWidh := int(getWidth())
	titleLength := len(">>> ULTIMATE SOUND VISUALIZER 2,000,000 <<<")
	if *printTitle && terminalWidh >= titleLength {
		titleMessage := "%s>>> ULTIMATE SOUND VISUALIZER 2,000,000 <<<%s\n"
		fill := ""
		for i := 0; i < (terminalWidh-titleLength)/2; i++ {
			fill += " "
		}
		fmt.Print(fmt.Sprintf(titleMessage, fill, fill))
	}
	
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		shutdown()
		os.Exit(0)
	}()
	
	<-make(chan struct{})
}

var fill = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}
var lastValues [channels][channelBuffer]float32

func updateCache(value float32, channel int) {
	l := channelBuffer - 1
	for i := l; i > 0; i-- {
		lastValues[channel][i] = lastValues[channel][i-1]
	}
	lastValues[channel][0] = value
}

func getAvg(channel int) float32 {
	var avg float32
	for _, v := range lastValues[channel] {
		avg += v
	}
	avg = avg / float32(channelBuffer)
	if avg > 1 {
		avg = 1
	}
	return avg
}

func calculateAdditionalBuffer(frameSize int) int {
	if frameSize > 512 {
		return 1
	}
	return int(512 / frameSize)
}

func printBar(value float32, channel int, width int) {
	updateCache(value, channel)
	value = getAvg(channel)
	
	var bar = ""
	if *printValues {
		width -= 10
		bar = fmt.Sprintf("\r %.3f |", value)
	} else {
		width -= 4
		bar = "\r |"
	}
	
	chars := int(float32(width) * value)
	for i := 0; i < chars; i++ {
		bar += fill[8]
	}
	
	if chars < width {
		fillIndex := (float32(width)*value - float32(chars)) * 8
		bar += fill[int(fillIndex)]
	}
	
	for i := 0; i <= width-chars-2; i++ {
		bar += fill[0]
	}
	
	fmt.Print(bar + "| ")
}

// SOD Section (Stack Overflow Development)
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getWidth() uint {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))
	
	if int(retCode) == -1 {
		panic(errno)
	}
	return uint(ws.Col)
}

// End of SOD
