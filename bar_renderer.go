package main

import "fmt"

const channel_buffor = 10

var fill_h = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

var last_values [channels][channel_buffor]float32

func update_cache(value float32, channel int) {
	l := len(last_values[channel]) - 1
	for i := l; i > 1; i-- {
		last_values[channel][i] = last_values[channel][i-1]
	}
	last_values[channel][0] = value
	last_values[channel][1] = value
}

func get_avg(channel int) float32 {
	var avg float32
	for _, v := range last_values[channel] {
		avg += v
	}
	avg = avg / float32(len(last_values[channel]))

	return avg
}

func printBar(value float32, channel int, width int) {
	update_cache(value, channel)
	value = get_avg(channel)
	if value > 1.0 {
		value = 1.0
	}

	bar := fmt.Sprintf("\r  %.3f  |", value)

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
