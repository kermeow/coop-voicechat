package coop

import (
	"log"
	"math"

	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
)

const SAMPLE_RATE = 48000 // opus is really annoying about this :(
const CHANNELS = 1        // mono

const OPUS_BITRATE = 64000      // 64 kbps
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

	state *voiceState

	pcmBuf  []float32
	rms     float32
	decoder *opus.Decoder
}

type voiceState struct {
	volume   float32
	fileName string
}

func newSpeaker() *speaker {
	s := &speaker{
		state: &voiceState{
			volume: 0,
		},

		pcmBuf: make([]float32, 0),
		rms:    0,
	}

	s.Stream, _ = portaudio.OpenDefaultStream(0, 1, SAMPLE_RATE, FRAMES_PER_BUFFER, s.processAudio)
	s.decoder, _ = opus.NewDecoder(SAMPLE_RATE, CHANNELS)

	return s
}

func (s *speaker) processAudio(out [][]float32) {
	if len(s.pcmBuf) < FRAMES_PER_BUFFER {
		return
	}

	var ms float64 = 0
	for i := range FRAMES_PER_BUFFER {
		pcm := s.pcmBuf[i]
		ms += math.Pow(float64(pcm), 2)
		out[0][i] = pcm * s.state.volume
	}
	s.rms = float32(math.Sqrt(ms / float64(FRAMES_PER_BUFFER)))
	s.pcmBuf = s.pcmBuf[FRAMES_PER_BUFFER:]
}

type AudioBridge struct {
	bridge *Bridge

	localState *voiceState
	speakers   map[uint8]*speaker

	inFrames []*opusFrame
	inStream *portaudio.Stream
	inBuf    []float32
	inRms    float32
	encBuf   []byte

	encoder *opus.Encoder
}

func NewAudioBridge(bridge *Bridge) *AudioBridge {
	inBuf := make([]float32, FRAMES_PER_BUFFER)
	inStream, _ := portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, FRAMES_PER_BUFFER, inBuf)
	encoder, _ := opus.NewEncoder(SAMPLE_RATE, CHANNELS, opus.AppVoIP)
	encoder.SetBitrate(OPUS_BITRATE)
	return &AudioBridge{
		bridge: bridge,

		localState: &voiceState{
			volume: 1,
		},
		speakers: make(map[uint8]*speaker),

		inFrames: make([]*opusFrame, 0),
		inStream: inStream,
		inBuf:    inBuf,
		inRms:    0,
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

	for states.Cursor < len(states.Data) {
		rawi, _ := states.ReadUint8()
		i, disconnected := rawi&0x7f, rawi&0x80

		speaker := b.speakers[i]

		if disconnected > 0 {
			if speaker != nil {
				speaker.Abort()
				delete(b.speakers, i)
				log.Println("Speaker", i, "removed")
			}
			continue
		} else if speaker == nil {
			speaker = newSpeaker()
			speaker.Start()
			b.speakers[i] = speaker
			log.Println("Speaker", i, "added")
		}

		speaker.state.volume, _ = states.ReadFloat32()

		fileNameLen, _ := states.ReadUint8()
		fileNameBytes, _ := states.ReadBytes(int(fileNameLen))
		speaker.state.fileName = string(fileNameBytes)
	}

	{
		localFile, err := b.bridge.recvFS.Get("local")
		if err != nil {
			return
		}

		b.localState.volume, _ = localFile.ReadFloat32()
	}

	for _, speaker := range b.speakers {
		inFile, err := b.bridge.recvFS.Get(speaker.state.fileName)
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
			for i, pcm := range pcmBuf {
				// soft clip
				pcmBuf[i] = float32(math.Tanh(float64(pcm)))
			}
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

	volumes := b.bridge.sendFS.Create("volumes")
	volumes.WriteUint8(0)
	volumes.WriteFloat32(b.inRms)
	for i, speaker := range b.speakers {
		volumes.WriteUint8(i)
		volumes.WriteFloat32(speaker.rms)
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

		pcmBuf := make([]float32, FRAMES_PER_BUFFER)

		var ms float64 = 0
		for i, pcm := range b.inBuf {
			ms += math.Pow(float64(pcm), 2)
			pcmBuf[i] = pcm * b.localState.volume
		}
		b.inRms = float32(math.Sqrt(ms / float64(FRAMES_PER_BUFFER)))

		i := max(0, len(b.inFrames)-MAX_OPUS_FRAMES)
		n, err := b.encoder.EncodeFloat32(pcmBuf, b.encBuf)
		frame := &opusFrame{
			syncFrame: 0,
			data:      make([]byte, n),
		}
		copy(frame.data, b.encBuf[:n])
		b.inFrames = append(b.inFrames[i:], frame)
	}
	return nil
}
