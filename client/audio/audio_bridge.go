package audio

import (
	"context"
	"coop-voicechat/coop"
	"log"

	"github.com/gopxl/beep/v2/speaker"
)

type AudioBridge struct {
	streamers map[uint8]*PlayerStreamer
}

func NewAudioBridge() *AudioBridge {
	a := &AudioBridge{
		streamers: make(map[uint8]*PlayerStreamer),
	}
	return a
}

func (b *AudioBridge) AddPlayer(player *coop.Player) {
	b.RemovePlayer(player.LocalIndex)

	log.Println("Adding speaker", player.LocalIndex)

	s := NewPlayerStreamer(player)
	b.streamers[player.LocalIndex] = s
	
	speaker.Play(s)
}

func (b *AudioBridge) RemovePlayer(localIndex uint8) {
	s := b.streamers[localIndex]
	if s != nil {
		log.Println("Removing speaker", localIndex)

		delete(b.streamers, localIndex)
	}
}

func (b *AudioBridge) Run(ctx context.Context) {
	log.Println("Audio bridge running")

running:
	for {
		select {
		case <-ctx.Done():
			b.stop()
			break running
		default:

		}
	}
}

func (b *AudioBridge) stop() {
	log.Println("Audio bridge stopping")
}
