package audio

import "github.com/gordonklaus/portaudio"

type Recorder struct {
	paStream *portaudio.Stream
	buffer   []float32
}

func NewRecorder() *Recorder {
	r := &Recorder{}
	r.buffer = make([]float32, OPUS_FRAME_SAMPLES)
	r.paStream, _ = portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, OPUS_FRAME_SAMPLES, r.buffer)
	return r
}

func (r *Recorder) Start() error {
	return r.paStream.Start()
}

func (r *Recorder) Stop() error {
	return r.paStream.Stop()
}

func (r *Recorder) Read() ([]float32, error) {
	err := r.paStream.Read()
	if err != nil {
		return nil, err
	}
	data := make([]float32, len(r.buffer))
	copy(data, r.buffer)
	return data, err
}
