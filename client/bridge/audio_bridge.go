package bridge

import (
	"context"
	"coop-voicechat/audio"
	"coop-voicechat/audio/effects"
	"coop-voicechat/coop"
	"log"
	"time"

	"github.com/gopxl/beep/v2/speaker"
)

const MAX_INPUT_FRAMES = 100

type frame struct {
	syncFrame uint32
	timestamp int
	data      []byte
}

type AudioBridge struct {
	bridge    *Bridge
	streamers map[uint8]*audio.PlayerStreamer

	inputStreamer *audio.InputStreamer
	denoiser      *effects.Denoiser
	analyzer      *effects.Analyzer

	inTimestamp int
	inQueue     []*frame
	opusBuffer  []float32
	opusEncoder *audio.OpusEncoder
}

func NewAudioBridge(b *Bridge) *AudioBridge {
	a := &AudioBridge{
		bridge:    b,
		streamers: make(map[uint8]*audio.PlayerStreamer),

		inTimestamp: 0,
		inQueue:     make([]*frame, 0),
		opusBuffer:  make([]float32, audio.OPUS_FRAME_SAMPLES),
		opusEncoder: audio.NewOpusEncoder(),
	}

	a.inputStreamer = &audio.InputStreamer{}
	a.inputStreamer.StartDefault()
	a.denoiser = effects.NewDenoiser(a.inputStreamer)
	a.analyzer = &effects.Analyzer{Streamer: a.denoiser, WindowSize: audio.SAMPLE_RATE_BEEP.N(50 * time.Millisecond)}

	return a
}

func (a *AudioBridge) encodeNext() {
	samples := make([][2]float64, audio.OPUS_FRAME_SAMPLES)

	n, ok := a.analyzer.Stream(samples)
	if n < audio.OPUS_FRAME_SAMPLES || !ok {
		log.Println("Error streaming input:", a.analyzer.Err())
		return
	}

	if !a.bridge.Connected {
		return
	}

	for i := range samples {
		a.opusBuffer[i] = float32(samples[i][0]+samples[i][1]) / 2
	}

	timestamp := a.inTimestamp
	a.inTimestamp++

	data, err := a.opusEncoder.Encode(a.opusBuffer)
	if err != nil {
		log.Println("Error encoding input:", err)
		return
	}

	f := &frame{
		syncFrame: 0,
		timestamp: timestamp,
		data:      data,
	}
	i := max(0, len(a.inQueue)-(MAX_INPUT_FRAMES-1))
	a.inQueue = append(a.inQueue[i:], f)
}

func (a *AudioBridge) connect() {
	a.inQueue = a.inQueue[:0]
}

func (a *AudioBridge) disconnect() {
	for i := range a.streamers {
		a.removePlayer(i)
	}
}

func (a *AudioBridge) addPlayer(player *coop.Player) {
	a.removePlayer(player.LocalIndex)

	log.Println("Adding speaker", player.LocalIndex)

	s := audio.NewPlayerStreamer(player)
	a.streamers[player.LocalIndex] = s

	speaker.Play(s)
}

func (a *AudioBridge) removePlayer(localIndex uint8) {
	s := a.streamers[localIndex]
	if s != nil {
		log.Println("Removing speaker", localIndex)
		s.Player = &coop.Player{LocalIndex: 0}
		delete(a.streamers, localIndex)
	}
}

func (a *AudioBridge) run(ctx context.Context) {
	log.Println("Audio bridge running")

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

	for i := range a.streamers {
		a.removePlayer(i)
	}
}

func (a *AudioBridge) recv() {
	for i, s := range a.streamers {
		f, err := a.bridge.RecvFs.Get(a.bridge.audioFiles[i])
		if err != nil {
			continue
		}

		f.Cursor = 4 // skip header
		for f.Cursor < len(f.Data) {
			sf, _ := f.ReadUint32()
			t, _ := f.ReadUint32()
			l, _ := f.ReadUint32()
			data, _ := f.ReadBytes(int(l))
			if sf <= a.bridge.syncLastRemoteFrame {
				continue
			}
			s.Push(data, int(t))
		}
	}
}

func (a *AudioBridge) send() {
	in := a.bridge.SendFs.Create("stream")
	in.WriteBytes(FILE_HEADER_BYTES)

	for _, f := range a.inQueue {
		if f.syncFrame == 0 {
			f.syncFrame = a.bridge.syncLocalFrame
		}
		if f.syncFrame < a.bridge.syncRemoteAckFrame-1 {
			continue
		}
		in.WriteUint32(f.syncFrame)
		in.WriteUint32(uint32(f.timestamp))
		in.WriteUint32(uint32(len(f.data)))
		in.WriteBytes(f.data)
	}

	vols := a.bridge.SendFs.Create("loudness")
	vols.WriteBytes(FILE_HEADER_BYTES)

	vols.WriteUint8(0)
	vols.WriteFloat64(a.analyzer.Rms())

	for i, s := range a.streamers {
		vols.WriteUint8(i)
		vols.WriteFloat64(s.Rms())
	}
}
