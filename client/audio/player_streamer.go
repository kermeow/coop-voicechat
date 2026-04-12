package audio

import (
	"coop-voicechat/audio/effects"
	"coop-voicechat/coop"
)

type PlayerStreamer struct {
	Player *coop.Player

	streamer *OpusStreamer
	analyzer *effects.Analyzer
	volume   *effects.Volume
}

func NewPlayerStreamer(player *coop.Player) *PlayerStreamer {
	s := &PlayerStreamer{
		Player:   player,
		streamer: NewOpusStreamer(),
	}
	s.analyzer = &effects.Analyzer{
		Streamer: s.streamer,
	}
	s.volume = &effects.Volume{
		Streamer: s.analyzer,
		Volume:   0,
	}
	return s
}

func (s *PlayerStreamer) Push(data []byte, timestamp int) {
	s.streamer.Push(data, timestamp)
}

func (s *PlayerStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if s.Player.LocalIndex < 1 {
		return 0, true // stops streaming
	}
	s.volume.Volume = s.Player.State.Volume
	n, ok = s.volume.Stream(samples)
	return n, ok
}

func (s *PlayerStreamer) Err() error {
	return s.streamer.Err()
}

func (s *PlayerStreamer) Rms() float64 {
	return s.analyzer.Rms()
}
