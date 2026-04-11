package audio

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
	nSamples, filled := len(samples), len(s.leftover)

	opusSamples := make([]float32, len(samples))
	copy(opusSamples, s.leftover)

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
			copied := copy(opusSamples[filled:], dec[offset:])
			if copied < len(dec) {
				s.leftover = make([]float32, len(dec)-copied)
				copy(s.leftover, dec[copied:])
			}
			filled += len(dec) - offset
		default:
			dec, s.err = s.decoder.DecodePLC(packet.Span)
			if s.err != nil {
				return filled, false
			}
			copy(opusSamples[filled:], dec)
			filled += len(dec)
		}
	}

	for i, sample := range opusSamples {
		f64 := float64(sample)
		samples[i][0] = f64
		samples[i][1] = f64
	}

	return filled, true
}

func (s *OpusStreamer) Err() error {
	return s.err
}
