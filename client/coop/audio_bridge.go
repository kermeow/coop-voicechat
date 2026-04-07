package coop

import (
	"log"
	"math"

	"github.com/gordonklaus/portaudio"
	"github.com/quartercastle/vector"
	"gopkg.in/hraban/opus.v2"
)

type vec = vector.Vector

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

type voiceState struct {
	volume      float32
	panL        float32
	panR        float32
	attenuation float32

	pos   vec
	level uint8
	area  uint8
}

func (v *voiceState) read(f *ModFSFile) {
	v.volume, _ = f.ReadFloat32()
	x, _ := f.ReadFloat64()
	y, _ := f.ReadFloat64()
	z, _ := f.ReadFloat64()
	v.pos = vec{x, y, z}
	v.level, _ = f.ReadUint8()
	v.area, _ = f.ReadUint8()
}

type speaker struct {
	*portaudio.Stream

	state    *voiceState
	fileName string

	pcmBuf  []float32
	rms     float32
	decoder *opus.Decoder
}

func newSpeaker() *speaker {
	s := &speaker{
		state: &voiceState{
			volume:      0,
			panL:        0,
			panR:        0,
			attenuation: 1,
		},

		pcmBuf: make([]float32, 0),
		rms:    0,
	}

	s.Stream, _ = portaudio.OpenDefaultStream(0, 2, SAMPLE_RATE, FRAMES_PER_BUFFER, s.processAudio)
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

		pcm *= s.state.volume * s.state.attenuation

		out[0][i] = pcm * s.state.panL
		out[1][i] = pcm * s.state.panR
	}
	s.rms = float32(math.Sqrt(ms / float64(FRAMES_PER_BUFFER)))
	s.pcmBuf = s.pcmBuf[FRAMES_PER_BUFFER:]
}

type audioBridge struct {
	bridge *Bridge

	localState   *voiceState
	localForward vec
	localRight   vec
	speakers     map[uint8]*speaker

	inFrames []*opusFrame
	inStream *portaudio.Stream
	inBuf    []float32
	inRms    float32
	encBuf   []byte

	encoder *opus.Encoder
}

func newAudioBridge(bridge *Bridge) *audioBridge {
	inBuf := make([]float32, FRAMES_PER_BUFFER)
	inStream, _ := portaudio.OpenDefaultStream(1, 0, SAMPLE_RATE, FRAMES_PER_BUFFER, inBuf)
	encoder, _ := opus.NewEncoder(SAMPLE_RATE, CHANNELS, opus.AppVoIP)
	encoder.SetBitrate(OPUS_BITRATE)
	return &audioBridge{
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

func (b *audioBridge) Run() {
	go b.inputLoop()
}

func (b *audioBridge) recv() {
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

		fileNameLen, _ := states.ReadUint8()
		fileNameBytes, _ := states.ReadBytes(int(fileNameLen))
		speaker.fileName = string(fileNameBytes)

		speaker.state.read(states)
	}

	{
		local, err := b.bridge.recvFS.Get("local")
		if err != nil {
			return
		}

		b.localState.read(local)

		px, _ := local.ReadFloat64()
		py, _ := local.ReadFloat64()
		pz, _ := local.ReadFloat64()
		lakituPos := vec{px, py, pz}

		fx, _ := local.ReadFloat64()
		fy, _ := local.ReadFloat64()
		fz, _ := local.ReadFloat64()
		lakituFoc := vec{fx, fy, fz}

		b.localForward = lakituFoc.Sub(lakituPos).Unit()
		b.localRight, _ = b.localForward.Cross(vec{0, 1, 0})
	}

	for _, speaker := range b.speakers {
		difference := speaker.state.pos.Sub(b.localState.pos)
		distance := difference.Magnitude()
		speaker.state.attenuation = min(1, 1/float32(distance/256))

		direction := difference.Unit()
		pan := direction.Dot(b.localRight)
		angle := (pan + 1) * (math.Pi / 4)
		speaker.state.panL = float32(math.Cos(angle))
		speaker.state.panR = float32(math.Sin(angle))

		inFile, err := b.bridge.recvFS.Get(speaker.fileName)
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

func (b *audioBridge) send() {
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

func (b *audioBridge) inputLoop() error {
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
