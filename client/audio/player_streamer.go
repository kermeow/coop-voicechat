package audio

import (
	"coop-voicechat/coop"

	"github.com/gopxl/beep/v2/effects"
)

type PlayerStreamer struct {
	Player *coop.Player

	streamer *OpusStreamer
	volume   *effects.Volume
}

func NewPlayerStreamer(player *coop.Player) *PlayerStreamer {
	s := &PlayerStreamer{
		Player:   player,
		streamer: NewOpusStreamer(),
	}
	s.volume = &effects.Volume{
		Streamer: s.streamer,
		Base:     2,
		Volume:   0,
		Silent:   false,
	}
	return s
}

func (s *PlayerStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.Player.LocalIndex < 1 {
		return 0, true // stops streaming
	}
	n, ok = s.volume.Stream(samples)
	return n, ok
}

func (s *PlayerStreamer) Err() error {
	return s.streamer.Err()
}
