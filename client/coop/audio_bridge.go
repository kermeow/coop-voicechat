package coop

import (
	"strconv"

	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
)

const SAMPLE_RATE = 24000 // opus is really annoying about this :(
const CHANNELS = 1        // mono

const OPUS_BITRATE = 32000      // 32 kbps
const OPUS_FRAME_SIZE_MS = 0.02 // 20 ms frames, we can go smaller but i dont want to

const MAX_OPUS_FRAMES = 10                  // 200 ms backlog
const MAX_DECODE_SAMPLES = SAMPLE_RATE * 10 // 200 ms backlog
const FRAMES_PER_BUFFER = int(SAMPLE_RATE * OPUS_FRAME_SIZE_MS * CHANNELS)

type opusFrame struct {
	syncFrame uint32
	data      []byte
}

type speaker struct {
	*portaudio.Stream
	pcmBuf  []float32
	decoder *opus.Decoder
}

func newSpeaker() *speaker {
	s := &speaker{}
	s.Stream, _ = portaudio.OpenDefaultStream(0, 1, SAMPLE_RATE, FRAMES_PER_BUFFER, s.processAudio)
	s.pcmBuf = make([]float32, 0)
	s.decoder, _ = opus.NewDecoder(SAMPLE_RATE, CHANNELS)
	return s
}

func (s *speaker) processAudio(out [][]float32) {
	if len(s.pcmBuf) < FRAMES_PER_BUFFER {
		return
	}
	for i := range FRAMES_PER_BUFFER {
		out[0][i] = s.pcmBuf[i]
	}
	s.pcmBuf = s.pcmBuf[FRAMES_PER_BUFFER:]
}

type AudioBridge struct {
	bridge *Bridge

	speakers map[uint8]*speaker

	inFrames []*opusFrame
	inStream *portaudio.Stream
	inBuf    []int16
	encBuf   []byte

	encoder *opus.Encoder
}

func NewAudioBridge(bridge *Bridge) *AudioBridge {
	inBuf := make([]int16, FRAMES_PER_BUFFER)
	inStream, _ := portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, FRAMES_PER_BUFFER, inBuf)
	encoder, _ := opus.NewEncoder(SAMPLE_RATE, CHANNELS, opus.AppVoIP)
	encoder.SetBitrate(OPUS_BITRATE)
	return &AudioBridge{
		bridge: bridge,

		speakers: make(map[uint8]*speaker),

		inFrames: make([]*opusFrame, 0),
		inStream: inStream,
		inBuf:    inBuf,
		encBuf:   make([]byte, 2048), // hungry hungry opus

		encoder: encoder,
	}
}

func (b *AudioBridge) Run() {
	go b.inputLoop()
}

func (b *AudioBridge) recv() {
	states, err := b.bridge.recvFS.Get("states")
	if err != nil {
		return
	}

	for _, v := range states.Data {
		speaker := b.speakers[v]
		if speaker == nil {
			speaker = newSpeaker()
			b.speakers[v] = speaker
			speaker.Start()
		}

		inFile, err := b.bridge.recvFS.Get(strconv.Itoa(int(v)))
		if err != nil {
			continue
		}
		for inFile.Cursor < len(inFile.Data) {
			syncFrame, _ := inFile.ReadUint32()
			dataLen, _ := inFile.ReadUint32()
			data, _ := inFile.ReadBytes(int(dataLen))
			if syncFrame <= b.bridge.syncLastRemoteFrame {
				continue
			}
			pcmBuf := make([]float32, FRAMES_PER_BUFFER)
			n, _ := speaker.decoder.DecodeFloat32(data, pcmBuf)
			speaker.pcmBuf = append(speaker.pcmBuf, pcmBuf[:n]...)
		}
	}
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

		i := max(0, len(b.inFrames)-MAX_OPUS_FRAMES)
		n, err := b.encoder.Encode(b.inBuf, b.encBuf)
		frame := &opusFrame{
			syncFrame: 0,
			data:      make([]byte, n),
		}
		copy(frame.data, b.encBuf[:n])
		b.inFrames = append(b.inFrames[i:], frame)
	}
	return nil
}
