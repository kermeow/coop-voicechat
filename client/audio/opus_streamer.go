package audio

import "math"

type OpusStreamer struct {
	jitter   *JitterBuffer
	decoder  *OpusDecoder
	leftover []float32
	err      error
}

func NewOpusStreamer() *OpusStreamer {
	s := &OpusStreamer{}
	s.jitter = NewJitterBuffer()
	s.decoder = NewOpusDecoder()
	s.leftover = make([]float32, 0)
	return s
}

func (o *OpusStreamer) Push(packet []byte, timestamp int) {
	o.jitter.Push(&JitterPacket{
		Data:      packet,
		Timestamp: timestamp,
	})
}

func (o *OpusStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	nSamples := len(samples)

	filled := o.convert(o.leftover, samples)
	o.leftover = o.leftover[filled:]

	for filled < nSamples {
		packet, _ := o.jitter.Pop()

		var dec []float32
		if packet != nil {
			dec, o.err = o.decoder.Decode(packet.Data)
			if o.err != nil {
				return filled, false
			}
		} else {
			dec, o.err = o.decoder.DecodePLC(OPUS_FRAME_SAMPLES)
			if o.err != nil {
				return filled, false
			}
		}

		copied := o.convert(dec, samples[filled:])
		o.leftover = dec[copied:]
		filled += copied
	}

	return filled, true
}

func (o *OpusStreamer) Err() error {
	return o.err
}

func (o *OpusStreamer) convert(in []float32, out [][2]float64) int {
	fill := min(len(in), len(out))
	for i := range fill {
		f64 := math.Tanh(float64(in[i])) // soft clip
		out[i][0] = f64
		out[i][1] = f64
	}
	return fill
}
