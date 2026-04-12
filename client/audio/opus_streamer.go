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

func (s *OpusStreamer) Put(packet []byte, timestamp int) {
	s.jitter.Put(&JitterPacket{
		Data:      packet,
		Timestamp: timestamp,
		Span:      OPUS_FRAME_SAMPLES,
	})
}

func (s *OpusStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	nSamples := len(samples)

	opusSamples := make([]float32, len(samples))
	filled := copy(opusSamples, s.leftover)
	s.leftover = make([]float32, 0)

	for filled < nSamples {
		packet := &JitterPacket{}
		offset := 0
		j := s.jitter.Get(packet, nSamples-filled, &offset)

		var dec []float32
		switch j {
		case JitterOk:
			dec, s.err = s.decoder.Decode(packet.Data)
			if s.err != nil {
				return filled, false
			}
		default:
			dec, s.err = s.decoder.DecodePLC(OPUS_FRAME_SAMPLES)
			if packet.Span < OPUS_FRAME_SAMPLES {
				dec = dec[:packet.Span]
			}
			if s.err != nil {
				return filled, false
			}
		}
		copied := copy(opusSamples[filled:], dec[offset:])
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
