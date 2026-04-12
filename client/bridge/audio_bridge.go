package bridge

import (
	"context"
	"coop-voicechat/audio"
	"coop-voicechat/coop"
	"log"

	"github.com/gopxl/beep/v2/speaker"
	"github.com/gordonklaus/portaudio"
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

	paInStream  *portaudio.Stream
	paInBuffer  []float32
	inTimestamp int
	inQueue     []*frame
	opusEncoder *audio.OpusEncoder
}

func NewAudioBridge(b *Bridge) *AudioBridge {
	a := &AudioBridge{
		bridge:    b,
		streamers: make(map[uint8]*audio.PlayerStreamer),

		paInBuffer:  make([]float32, audio.OPUS_FRAME_SAMPLES),
		inTimestamp: 0,
		inQueue:     make([]*frame, 0),
		opusEncoder: audio.NewOpusEncoder(),
	}
	paInStream, _ := portaudio.OpenDefaultStream(1, 0, audio.SAMPLE_RATE, audio.OPUS_FRAME_SAMPLES, a.paInBuffer)
	a.paInStream = paInStream
	return a
}

func (a *AudioBridge) encodeNext() {
	err := a.paInStream.Read()
	if err != nil {
		log.Println("Error reading input:", err)
		return
	}

	if !a.bridge.Connected {
		return
	}

	timestamp := a.inTimestamp
	a.inTimestamp += len(a.paInBuffer)

	data, err := a.opusEncoder.Encode(a.paInBuffer)
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
			if sf < a.bridge.syncRemoteFrame {
				continue
			}
			s.Put(data, int(t))
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
		if f.syncFrame < a.bridge.syncRemoteAckFrame {
			continue
		}
		in.WriteUint32(f.syncFrame)
		in.WriteUint32(uint32(f.timestamp))
		in.WriteUint32(uint32(len(f.data)))
		in.WriteBytes(f.data)
	}
}
