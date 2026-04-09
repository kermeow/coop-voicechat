package audio

import "github.com/hraban/opus"

type OpusEncoder struct {
	encoder *opus.Encoder
	buffer  []byte
}

func NewOpusEncoder() *OpusEncoder {
	o := &OpusEncoder{}
	o.encoder, _ = opus.NewEncoder(SAMPLE_RATE, 1, opus.AppVoIP)
	o.encoder.SetBitrate(OPUS_BITRATE)
	o.buffer = make([]byte, 1276) // recommended size by opus
	return o
}

func (o *OpusEncoder) Encode(pcm []float32) ([]byte, error) {
	n, err := o.encoder.EncodeFloat32(pcm, o.buffer)
	if err != nil {
		return nil, err
	}
	data := make([]byte, n)
	copy(data, o.buffer[:n])
	return data, err
}
