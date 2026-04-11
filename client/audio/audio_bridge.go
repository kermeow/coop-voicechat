package audio

import (
	"context"
	"coop-voicechat/coop"
	"log"

	"github.com/gopxl/beep/v2/speaker"
	"github.com/gordonklaus/portaudio"
)

const MAX_INPUT_FRAMES = 16

type AudioBridge struct {
	streamers map[uint8]*PlayerStreamer

	paInStream  *portaudio.Stream
	paInBuffer  []float32
	inTimestamp int
	opusEncoder *OpusEncoder
}

func NewAudioBridge() *AudioBridge {
	a := &AudioBridge{
		streamers: make(map[uint8]*PlayerStreamer),

		paInBuffer:  make([]float32, OPUS_FRAME_SAMPLES),
		inTimestamp: 0,
		opusEncoder: NewOpusEncoder(),
	}
	paInStream, _ := portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, OPUS_FRAME_SAMPLES, a.paInBuffer)
	a.paInStream = paInStream
	return a
}

func (a *AudioBridge) encodeNext() {
	err := a.paInStream.Read()
	if err != nil {
		log.Println("Error reading input:", err)
		return
	}

	timestamp := a.inTimestamp
	a.inTimestamp += len(a.paInBuffer)

	data, err := a.opusEncoder.Encode(a.paInBuffer)
	if err != nil {
		log.Println("Error encoding input:", err)
		return
	}

	
}

func (a *AudioBridge) AddPlayer(player *coop.Player) {
	a.RemovePlayer(player.LocalIndex)

	log.Println("Adding speaker", player.LocalIndex)

	s := NewPlayerStreamer(player)
	a.streamers[player.LocalIndex] = s

	speaker.Play(s)
}

func (a *AudioBridge) RemovePlayer(localIndex uint8) {
	s := a.streamers[localIndex]
	if s != nil {
		log.Println("Removing speaker", localIndex)
		s.player = &coop.Player{LocalIndex: 0}
		delete(a.streamers, localIndex)
	}
}

func (a *AudioBridge) Run(ctx context.Context) {
	log.Println("Audio bridge running")

	a.paInStream.Start()

running:
	for {
		select {
		case <-ctx.Done():
			a.stop()
			break running
		default:
			a.encodeNext()
		}
	}
}

func (a *AudioBridge) stop() {
	log.Println("Audio bridge stopping")

	a.paInStream.Abort()

	for i := range a.streamers {
		a.RemovePlayer(i)
	}
}
