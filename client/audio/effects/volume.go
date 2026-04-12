package effects

import (
	"math"

	"github.com/gopxl/beep/v2"
)

type Volume struct {
	Streamer beep.Streamer
	Volume   float64

	gain float64
}

func (v *Volume) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.Streamer.Stream(samples)

	gain := math.Pow(v.Volume, 2)

	for i := range samples {
		v.gain = (v.gain + v.gain + gain) / 3
		if math.Abs(v.gain-gain) < 0.01 {
			v.gain = gain
		}
		samples[i][0] *= v.gain
		samples[i][1] *= v.gain
	}

	return n, ok
}

func (v *Volume) Err() error {
	return v.Streamer.Err()
}
