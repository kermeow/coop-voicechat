package audio

import (
	"math"

	"github.com/gordonklaus/portaudio"
)

type InputStreamer struct {
	paInStream *portaudio.Stream
	paInBuffer []float32
	overBuffer []float32
	err        error
}

const bufferSize = 512

func (i *InputStreamer) StartDefault() error {
	i.paInBuffer = make([]float32, bufferSize)
	paInStream, err := portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, bufferSize, i.paInBuffer)
	if err != nil {
		return err
	}
	paInStream.Start()
	i.paInStream = paInStream
	return nil
}

func (i *InputStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	nSamples := len(samples)

	filled := i.convert(i.overBuffer, samples)
	i.overBuffer = i.overBuffer[filled:]

	for filled < nSamples {
		i.err = i.paInStream.Read()
		copied := i.convert(i.paInBuffer, samples[filled:])
		i.overBuffer = make([]float32, bufferSize-copied)
		copy(i.overBuffer, i.paInBuffer[copied:])
		filled += copied
	}

	return filled, true
}

func (i *InputStreamer) Err() error {
	return i.err
}

func (i *InputStreamer) convert(in []float32, out [][2]float64) int {
	fill := min(len(in), len(out))
	for i := range fill {
		f64 := math.Tanh(float64(in[i]))
		out[i][0] = f64
		out[i][1] = f64
	}
	return fill
}
