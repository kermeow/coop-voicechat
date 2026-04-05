package audio

import (
	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
)

type Recorder struct {
	OpusFrames [][]byte
	Running    bool

	rawBuf  []int16
	encBuf  []byte
	stream  *portaudio.Stream
	encoder *opus.Encoder
}

func NewRecorder() (*Recorder, error) {
	rawBuf := make([]int16, SAMPLE_RATE*OPUS_FRAME_SIZE_MS*CHANNELS)
	encBuf := make([]byte, 2000)

	encoder, err := opus.NewEncoder(SAMPLE_RATE, CHANNELS, opus.AppVoIP)
	if err != nil {
		return nil, err
	}
	err = encoder.SetBitrate(OPUS_BITRATE)
	if err != nil {
		return nil, err
	}

	stream, err := portaudio.OpenDefaultStream(CHANNELS, 0, SAMPLE_RATE, len(rawBuf), rawBuf)
	if err != nil {
		return nil, err
	}

	return &Recorder{
		OpusFrames: make([][]byte, 0),
		Running:    false,

		rawBuf:  rawBuf,
		encBuf:  encBuf,
		stream:  stream,
		encoder: encoder,
	}, nil
}

func (r *Recorder) encode() error {
	n, err := r.encoder.Encode(r.rawBuf, r.encBuf)
	if err != nil {
		return err
	}
	frame := r.encBuf[:n]
	r.OpusFrames = append(r.OpusFrames, frame)
	return nil
}

func (r *Recorder) Start() error {
	if r.Running {
		return nil
	}

	err := r.stream.Start()
	if err != nil {
		return err
	}
	defer r.stream.Stop()

	r.Running = true

	for r.Running {
		err := r.stream.Read()
		if err != nil {
			return err
		}

		err = r.encode()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Recorder) Stop() {
	if !r.Running {
		return
	}

	r.Running = false
}

func (r *Recorder) Read() ([][]byte, error) {
	// if len(r.OpusFrames) == 0 {
	// 	r.waitingForFrame = true
	// 	<-r.frameAdded
	// }
	frames := make([][]byte, len(r.OpusFrames))
	copy(frames, r.OpusFrames)
	r.OpusFrames = make([][]byte, 0)
	return frames, nil
}
