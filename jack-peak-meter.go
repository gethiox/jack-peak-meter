package main

import (
	"fmt"
	"github.com/xthexder/go-jack"
	"os"
	"syscall"
	"unsafe"
	"flag"
)

const channels = 2              // Amount of input channels
const channel_buffer = 5        // Smoothing graph with last n printed samples, set 1 to disable
const arbitrary_amplifier = 3.8 // Compensate weak audio signal with this ultimate amplifier value

var additional_buffer int

var buffer_size int
var PortsIn []*jack.Port
var client *jack.Client
var avg_main [channels]float32
var counter int
var avg float32

var print_title *bool
var print_values *bool

func process(nframes uint32) int {
	counter += 1
	for i, port := range PortsIn {
		samples := port.GetBuffer(nframes)

		for _, sample := range samples {
			if sample < 0 {
				avg = avg - float32(sample)
			} else {
				avg = avg + float32(sample)
			}
		}
		avg = avg * arbitrary_amplifier
		avg = float32(avg) / float32(buffer_size)

		avg_main[i] += avg

		if counter >= additional_buffer {
			printBar(avg_main[i]/float32(additional_buffer), i, int(getWidth()-19))
			avg_main[i] = 0
		}

		fmt.Print("\n")
	}
	if counter >= additional_buffer {
		counter = 0
	}
	for i := 0; i < channels; i++ {
		fmt.Print("\033[F")
	}
	return 0
}

func shutdown() {
	fmt.Println("Shutting down")
	os.Exit(1)
}

//func foo(buffer_size int) {
//	buffer_size = int(client.GetBufferSize())
//	additional_buffer = calculate_additional_buffer(buffer_size)
//}

func main() {
	print_title = flag.Bool("title", false, "Print name of my ultimate visualizer")
	print_values = flag.Bool("values", false, "Print value before each channel of visualizer")
	flag.Parse()

	var status int
	var client_name string

	for i := 0; i < 10; i++ {
		client_name = fmt.Sprintf("spectrum analyser %d", i)
		client, status = jack.ClientOpen(client_name, jack.NoStartServer)
		if status == 0 {
			break
		}
	}
	if status != 0 {
		fmt.Println("Status:", status)
		return
	}

	defer client.Close()
	
	buffer_size = int(client.GetBufferSize())
	additional_buffer = calculate_additional_buffer(buffer_size)
	
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
			client.Connect("system:monitor_1", fmt.Sprintf("%s:input_%d", client_name, i))
		} else {
			client.Connect("system:monitor_2", fmt.Sprintf("%s:input_%d", client_name, i))
		}
		PortsIn = append(PortsIn, port)
	}

	fmt.Print("\n\n")
	terminal_widh := int(getWidth())
	title_length := len(">>> ULTIMATE SOUND VISUALIZER 2,000,000 <<<")
	if *print_title && terminal_widh >= title_length {
		title_message := "%s>>> ULTIMATE SOUND VISUALIZER 2,000,000 <<<%s\n"
		fill := ""
		for i := 0; i < (terminal_widh-title_length)/2; i++ {
			fill += " "
		}
		fmt.Print(fmt.Sprintf(title_message, fill, fill))
	}
	<-make(chan struct{})
}

var fill_h = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}
var last_values [channels][channel_buffer]float32

func update_cache(value float32, channel int) {
	l := channel_buffer - 1
	for i := l; i > 0; i-- {
		last_values[channel][i] = last_values[channel][i-1]
	}
	last_values[channel][0] = value
}

func get_avg(channel int) float32 {
	var avg float32
	for _, v := range last_values[channel] {
		avg += v
	}
	avg = avg / float32(channel_buffer)
	if avg > 1 {
		avg = 1
	}
	return avg
}

func calculate_additional_buffer(frame_size int) int {
	if frame_size == 16 {
		return 64
	}
	if frame_size == 32 {
		return 32
	}
	if frame_size == 64 {
		return 16
	}
	if frame_size == 128 {
		return 8
	}
	if frame_size == 256 {
		return 4
	}
	if frame_size == 512 {
		return 2
	}
	return 1
}


func printBar(value float32, channel int, width int) {
	update_cache(value, channel)
	value = get_avg(channel)

	var bar = ""
	if *print_values {
		bar = fmt.Sprintf("\r  %.3f  |", value)
	} else {
		bar = "\r         |"
	}

	chars := int(float32(width) * value)
	for i := 0; i < chars; i++ {
		bar += fill_h[8]
	}

	if chars < width {
		fill_index := (float32(width)*value - float32(chars)) * 8
		bar += fill_h[int(fill_index)]
	}

	for i := 0; i <= width-chars-2; i++ {
		bar += fill_h[0]
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
