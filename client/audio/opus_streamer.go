package audio

import (
	"math"
)

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

func (s *OpusStreamer) Push(packet []byte, timestamp int) {
	s.jitter.Push(&JitterPacket{
		Data:      packet,
		Timestamp: timestamp,
	})
}

func (s *OpusStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	nSamples := len(samples)

	opusSamples := make([]float32, len(samples))
	filled := copy(opusSamples, s.leftover)
	s.leftover = s.leftover[filled:]

	for filled < nSamples {
		packet, _ := s.jitter.Pop()

		var dec []float32
		if packet != nil {
			dec, s.err = s.decoder.Decode(packet.Data)
			if s.err != nil {
				return filled, false
			}
		} else {
			dec, s.err = s.decoder.DecodePLC(OPUS_FRAME_SAMPLES)
			if s.err != nil {
				return filled, false
			}
		}

		copied := copy(opusSamples[filled:], dec)
		if copied < len(dec) {
			s.leftover = make([]float32, len(dec)-copied)
			copy(s.leftover, dec[copied:])
		}
		filled += copied
	}

	for i, sample := range opusSamples {
		f64 := math.Tanh(float64(sample)) // soft clip
		samples[i][0] = f64
		samples[i][1] = f64
	}

	return filled, true
}

func (s *OpusStreamer) Err() error {
	return s.err
}
