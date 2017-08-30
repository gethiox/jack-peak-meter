# ULTIMATE SOUND VISUALIZER 2,000,000

Welcome!  
Well, it's not as ultimate as it could be but also it's not so bad so far.

![visualizer animation](/doc/animation.gif)

# Requirements

`ULTIMATE SOUND VISUALIZER 2,000,000`® uses [JACK Audio Connection Kit](http://jackaudio.org/) as base sound system,
if you are using Ubuntu or another distribution with PulseAudio I have good news for you — You even don't need to
download this piece of script on your computer. I made it for myself as my first attempt into [Go](https://golang.org/)
programming language.  
Code uses [xthexder](https://github.com/xthexder)'s [go-jack](https://github.com/xthexder/go-jack) library/bindings for JACK Audio Connection Kit.  

# Usage

    ./jack-peak-meter -title
    
# Build

    go build jack-peak-meter.go
    
