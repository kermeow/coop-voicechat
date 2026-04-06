package coop

import (
	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
)

const SAMPLE_RATE = 24000 // opus is really annoying about this :(
const CHANNELS = 1        // mono

const OPUS_BITRATE = 32000      // 32 kbps
const OPUS_FRAME_SIZE_MS = 0.02 // 20 ms frames, we can go smaller but i dont want to

const MAX_AUDIO_FRAMES = 10 // 200 ms backlog

type audioFrame struct {
	syncFrame uint32
	data      []byte
}

type AudioBridge struct {
	bridge *Bridge

	inFrames []*audioFrame
	inStream *portaudio.Stream
	inBuf    []int16
	encBuf   []byte

	encoder *opus.Encoder
}

func NewAudioBridge(bridge *Bridge) *AudioBridge {
	framesPerBuffer := int(SAMPLE_RATE * OPUS_FRAME_SIZE_MS * CHANNELS)
	inBuf := make([]int16, framesPerBuffer)
	inStream, _ := portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, framesPerBuffer, inBuf)
	encoder, _ := opus.NewEncoder(SAMPLE_RATE, CHANNELS, opus.AppVoIP)
	encoder.SetBitrate(OPUS_BITRATE)
	return &AudioBridge{
		bridge: bridge,

		inFrames: make([]*audioFrame, 0),
		inStream: inStream,
		inBuf:    inBuf,
		encBuf:   make([]byte, 2048), // hungry hungry opus

		encoder: encoder,
	}
}

func (b *AudioBridge) Run() {
	go b.inputLoop()
	go b.outputLoop()
}

func (b *AudioBridge) send() {
	n := 0
	j := 0
	for i := len(b.inFrames); i > 0; i-- {
		j = i - 1
		frame := b.inFrames[j]
		if frame.syncFrame == 0 {
			frame.syncFrame = b.bridge.syncLocalFrame
		}
		if frame.syncFrame <= b.bridge.syncRemoteAckFrame {
			// log.Println(frame.syncFrame, b.bridge.syncRemoteAckFrame)
			break
		}
		n++
	}

	// log.Println(n, "frames sent, clipped", j)

	b.inFrames = b.inFrames[j:]

	recording := b.bridge.sendFS.Create("recording")
	for _, v := range b.inFrames {
		recording.WriteUint32(v.syncFrame)
		recording.WriteUint32(uint32(len(v.data)))
		recording.WriteBytes(v.data)
	}
}

func (b *AudioBridge) inputLoop() error {
	b.inStream.Start()
	defer b.inStream.Abort()

	for b.bridge.Running {
		err := b.inStream.Read()
		if err != nil {
			return err
		}

		if !b.bridge.Connected {
			continue
		}

		i := max(0, len(b.inFrames)-MAX_AUDIO_FRAMES)
		n, err := b.encoder.Encode(b.inBuf, b.encBuf)
		b.inFrames = append(b.inFrames[i:], &audioFrame{
			syncFrame: 0,
			data:      b.encBuf[:n],
		})
	}
	return nil
}

func (b *AudioBridge) outputLoop() error {
	return nil
}
