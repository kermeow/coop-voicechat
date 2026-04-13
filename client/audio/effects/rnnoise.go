package effects

import (
	"github.com/gopxl/beep/v2"
	"github.com/kermeow/rnnoise"
)

type Denoiser struct {
	Streamer beep.Streamer
	Bypass   bool

	stream        [][2]float64
	denoiseState  *rnnoise.DenoiseState
	denoiseBuffer []float32
	denoiseErr    error
}

func NewDenoiser(streamer beep.Streamer) *Denoiser {
	denoiseState, _ := rnnoise.NewDenoiseState()
	return &Denoiser{
		Streamer: streamer,
		Bypass:   false,

		stream:        make([][2]float64, rnnoise.GetFrameSize()),
		denoiseState:  denoiseState,
		denoiseBuffer: make([]float32, 0),
		denoiseErr:    nil,
	}
}

func (d *Denoiser) Stream(samples [][2]float64) (n int, ok bool) {
	nSamples := len(samples)

	filled := d.convert(d.denoiseBuffer, samples)
	d.denoiseBuffer = d.denoiseBuffer[filled:]

	for filled < nSamples {
		fs := rnnoise.GetFrameSize()

		sn, sok := d.Streamer.Stream(d.stream)
		if sn < fs || !sok {
			return filled, sok
		}

		if d.Bypass {
			copied := copy(samples[filled:], d.stream)
			d.denoiseBuffer = d.denoiseBuffer[copied:]
			filled += copied
			continue
		}

		inBuffer := make([]float32, fs)
		d.denoiseBuffer = make([]float32, fs)
		for i := range fs {
			inBuffer[i] = 32767 * float32(d.stream[i][0]+d.stream[i][1]) / 2
		}

		_, d.denoiseErr = d.denoiseState.ProcessFrame(d.denoiseBuffer, inBuffer)

		if d.denoiseErr != nil {
			return filled, false
		}

		copied := d.convert(d.denoiseBuffer, samples[filled:])
		d.denoiseBuffer = d.denoiseBuffer[copied:]
		filled += copied
	}

	return filled, true
}

func (d *Denoiser) Err() error {
	if d.denoiseErr != nil {
		return d.denoiseErr
	}
	return d.Streamer.Err()
}

func (d *Denoiser) convert(in []float32, out [][2]float64) int {
	fill := min(len(in), len(out))
	for i := range fill {
		f64 := float64(in[i] / 32767)
		out[i][0] = f64
		out[i][1] = f64
	}
	return fill
}
