package audio

import (
	"context"
	"log"
)

type AudioBridge struct {
}

func NewAudioBridge() *AudioBridge {
	a := &AudioBridge{}
	return a
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
