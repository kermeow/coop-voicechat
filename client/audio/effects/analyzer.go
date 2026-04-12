package effects

import (
	"math"

	"github.com/gopxl/beep/v2"
)

type Analyzer struct {
	Streamer beep.Streamer

	rms float64
}

func (a *Analyzer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = a.Streamer.Stream(samples)
	rms := 0.0
	for i := range samples[:n] {
		vol := (samples[i][0] + samples[i][1]) / 2
		rms += vol * vol
	}
	rms = math.Sqrt(rms / float64(n))
	a.rms = rms
	return n, ok
}

func (a *Analyzer) Err() error {
	return a.Streamer.Err()
}

func (a *Analyzer) Rms() float64 {
	return a.rms
}
