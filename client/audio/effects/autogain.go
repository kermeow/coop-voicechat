package effects

import (
	"math"

	"github.com/gopxl/beep/v2"
)

type AutoGain struct {
	Streamer beep.Streamer
	Bypass   bool

	WindowSize int
	Attack     int
	Release    int

	window [][2]float64
	gain   float64
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
			peakDb, rmsDb := Amp2Db(peak), Amp2Db(rms)

			if vad && peakDb > -60 {
				target, add := Db2Amp(-18-rmsDb), 0.0
				if target > a.gain {
					add = float64(target-a.gain) / float64(a.Attack)
				} else {
					add = float64(target-a.gain) / float64(a.Attack)
				}
				a.gain = min(target, a.gain+add)
			} else if !vad {
				a.gain = max(0, a.gain+float64(-a.gain)/float64(a.Release))
			}

			samples[i][0] *= a.gain
			samples[i][1] *= a.gain
		}
	} else {
		for i, sample := range samples[:n] {
			peak, rms := a.getPeakAndRMS(sample)
			_, rmsDb := Amp2Db(peak), Amp2Db(rms)

			if rmsDb > -40 {
				target, add := Db2Amp(-18-rmsDb), 0.0
				if target > a.gain {
					add = float64(target-a.gain) / float64(a.Attack)
				} else {
					add = float64(target-a.gain) / float64(a.Attack)
				}
				a.gain = min(target, a.gain+add)
			} else {
				a.gain = max(0, a.gain+float64(-a.gain)/float64(a.Release))
			}

			samples[i][0] = math.Tanh(sample[0] * a.gain)
			samples[i][1] = math.Tanh(sample[1] * a.gain)
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
