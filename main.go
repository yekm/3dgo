// Copyright 2018 Jacques Supcik / HEIA-FR
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"image"
	_ "image/png"
	"time"

	ws2811 "github.com/rpi-ws281x/rpi-ws281x-go"
)

var (
	brightness = 90
	w          = 32
	h          = 16
	dw         = 32
	dh         = 8
	maxCount   = 50
)

type wsEngine interface {
	Init() error
	Render() error
	Wait() error
	Fini()
	Leds(channel int) []uint32
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type display struct {
	r     Renderer
	ws    wsEngine
	frame uint
}

func coordinatesToIndex(bounds image.Rectangle, x int, y int) int {
	if x%2 == 0 {
		return (x-bounds.Min.X)*h + (y - bounds.Min.Y)
	}
	return (x-bounds.Min.X)*h + (h - 1) - (y - bounds.Min.Y)
}

func coordinatesToIndex3(x int, y int) int {
	if x%2 == 0 {
		return x*dh + y
	}
	return x*dh + (dh - 1) - y
}

func coordinatesToIndex2(x int, y int) int {
	c3 := coordinatesToIndex3(x%dw, y%dh)
	// TODO: add support for horisontal pieces
	c := c3 + (dw*dh)*(y/dh)
	return c
}

func rgbToColor(r uint32, g uint32, b uint32) uint32 {
	return ((r>>8)&0xff)<<16 + ((g>>8)&0xff)<<8 + ((b >> 8) & 0xff)
}

func (disp *display) display() error {
	img := disp.r.frame()
	//return nil
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			disp.ws.Leds(0)[coordinatesToIndex2(x, y)] = rgbToColor(r, g, b)
		}
	}
	return disp.ws.Render()
}

func (disp *display) clear() {
	for i := 0; i < w*h; i++ {
		disp.ws.Leds(0)[i] = 0
	}
}

func main() {

	fb := flag.Int("b", 255, "brightness")
	fw := flag.Int("w", 32, "w")
	fh := flag.Int("h", 16, "h")
	fs := flag.Int("s", 2, "sleep between frames")
	fc := flag.Int("c", 2, "cycles")
	ff := flag.String("f", "", "stl file")
	flag.Parse()
	w = *fw
	h = *fh

	opt := ws2811.DefaultOptions
	opt.Channels[0].Brightness = *fb
	opt.Channels[0].LedCount = w * h

	dev, err := ws2811.MakeWS2811(&opt)
	check(err)

	disp := &display{
		ws: dev,
		r:  get_renderer(*ff, w, h),
	}

	check(dev.Init())

	defer dev.Fini()

	for count := 0; count < (*fc); count++ {
		disp.display()
		time.Sleep(time.Duration(*fs) * time.Millisecond)
	}

	disp.clear()
	dev.Render()
}
