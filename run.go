package main

import (
	"fmt"
	"github.com/xthexder/go-jack"
	"os"
	"syscall"
	"unsafe"
	"flag"
)

var buffer_size int
var PortsIn []*jack.Port

const channels int = 2

var chuj int = 0
var avg_main [channels]float32

var additional_buffer = 16

func process(nframes uint32) int {
	chuj += 1
	for i, in := range PortsIn {
		samplesIn := in.GetBuffer(nframes)

		var avg float32 = 0.0
		for _, sample := range samplesIn {
			if sample < 0 {
				avg = avg - float32(sample)
			} else {
				avg = avg + float32(sample)
			}
		}
		avg = avg * 3 // some arbitrary multiplier
		avg = float32(avg) / float32(buffer_size)

		avg_main[i] += float32(avg)

		if chuj > additional_buffer {

			printBar(avg_main[i]/float32(additional_buffer), i, int(getWidth()-19))
			avg_main[i] = 0.0
		}

		fmt.Print("\n")
	}
	if chuj > additional_buffer {
		chuj = 0
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

func main() {
	boolPtr := flag.Bool("title", false, "Print name of my ultimate visualizer")
	flag.Parse()

	var client *jack.Client
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

	if code := client.SetProcessCallback(process); code != 0 {
		fmt.Println("Failed to set process callback:", code)
		return
	}
	client.OnShutdown(shutdown)

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
	if *boolPtr && terminal_widh >= title_length {
		title_message := "%s>>> ULTIMATE SOUND VISUALIZER 2,000,000 <<<%s\n"
		fill := ""
		for i := 0; i < (terminal_widh-title_length)/2; i++ {
			fill += " "
		}
		fmt.Print(fmt.Sprintf(title_message, fill, fill))
	}
	<-make(chan struct{})
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
