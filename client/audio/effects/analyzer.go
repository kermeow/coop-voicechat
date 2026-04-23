package effects

import (
	"math"

	"github.com/gopxl/beep/v2"
)

type Analyzer struct {
	Streamer   beep.Streamer
	WindowSize int

	window []float64
}

func (a *Analyzer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = a.Streamer.Stream(samples)
	win := make([]float64, n)
	for i := range samples[:n] {
		win[i] = float64(samples[i][0]+samples[i][1]) / 2
	}
	a.window = append(a.window, win...)
	if len(a.window) > a.WindowSize {
		a.window = a.window[len(a.window)-a.WindowSize:]
	}
	return n, ok
}

func (a *Analyzer) Err() error {
	return a.Streamer.Err()
}

func (a *Analyzer) Rms() float64 {
	ms := 0.0
	for _, sample := range a.window {
		ms += sample * sample
	}
	return math.Sqrt(ms / float64(a.WindowSize))
}
