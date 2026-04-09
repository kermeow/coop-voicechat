package audio

import (
	"github.com/hraban/opus"
)

type OpusDecoder struct {
	decoder *opus.Decoder
	buffer  []float32
}

func NewOpusDecoder() *OpusDecoder {
	o := &OpusDecoder{}
	o.decoder, _ = opus.NewDecoder(SAMPLE_RATE, 1)
	o.buffer = make([]float32, OPUS_FRAME_SAMPLES)
	return o
}

func (o *OpusDecoder) Decode(data []byte) ([]float32, error) {
	n, err := o.decoder.DecodeFloat32(data, o.buffer)
	if err != nil {
		return nil, err
	}
	pcm := make([]float32, n)
	copy(pcm, o.buffer[:n])
	return pcm, err
}

func (o *OpusDecoder) DecodePLC(size int) ([]float32, error) {
	pcm := make([]float32, size)
	err := o.decoder.DecodePLCFloat32(pcm)
	return pcm, err
}
