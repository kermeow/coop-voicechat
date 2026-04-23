package effects

import (
	"math"

	"github.com/gopxl/beep/v2"
)

type AutoGain struct {
	Streamer   beep.Streamer
	Bypass     bool
	WindowSize int

	window [][2]float64
}

func (a *AutoGain) Stream(samples [][2]float64) (n int, ok bool) {
	if a.Bypass {
		return a.Streamer.Stream(samples)
	}

	n, ok = a.Streamer.Stream(samples)

	if denoiser, isDenoiser := a.Streamer.(*Denoiser); isDenoiser && !denoiser.Bypass {
		for i, sample := range samples[:n] {
			vad := denoiser.vadBuffer[i] >= 0.9
			peak, rms := a.getPeakAndRMS(sample)
			_, rmsDb := Amp2Db(peak), Amp2Db(rms)

			if vad && rmsDb < -12 {
				gain := Db2Amp(-18 - rmsDb)
				samples[i][0] *= gain
				samples[i][1] *= gain
			}
		}
	} else {
		for i, sample := range samples[:n] {
			peak, rms := a.getPeakAndRMS(sample)
			_, rmsDb := Amp2Db(peak), Amp2Db(rms)

			if rmsDb > -40 {
				gain := Db2Amp(-18 - rmsDb)
				samples[i][0] *= gain
				samples[i][1] *= gain
			}
		}
	}

	return n, ok
}

func (a *AutoGain) Err() error {
	return a.Streamer.Err()
}

func (a *AutoGain) getPeakAndRMS(sample [2]float64) (peak, rms float64) {
	a.window = append(a.window, sample)
	if len(a.window) > a.WindowSize {
		a.window = a.window[len(a.window)-a.WindowSize:]
	}
	peak, rms = 0, 0
	for _, s := range a.window {
		pcm := (s[0] + s[1]) / 2
		if math.Abs(pcm) > peak {
			peak = math.Abs(pcm)
		}
		rms += pcm * pcm
	}
	rms = math.Sqrt(rms / float64(a.WindowSize))
	return peak, rms
}
