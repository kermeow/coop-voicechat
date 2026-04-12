package audio

import (
	"errors"
	"log"
	"sync"
)

const JITTER_BUFFER_MAX = 10
const JITTER_BUFFER_MIN = 3

var (
	ErrJitterBuffering = errors.New("jitter buffer not ready")
	ErrJitterUnderrun  = errors.New("jitter buffer empty")
	ErrJitterMissing   = errors.New("jitter buffer packet lost")
)

type JitterPacket struct {
	Timestamp int
	Data      []byte
}

type JitterBuffer struct {
	packets     []*JitterPacket
	playoutHead int
	buffering   bool
	mutex       sync.Mutex
}

func NewJitterBuffer() *JitterBuffer {
	j := &JitterBuffer{
		packets:     make([]*JitterPacket, 0),
		playoutHead: -1,
		buffering:   true,
	}
	return j
}

func (j *JitterBuffer) Push(packet *JitterPacket) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	i := max(0, len(j.packets)-(JITTER_BUFFER_MAX-1))
	j.packets = append(j.packets[i:], packet)

	if len(j.packets) > JITTER_BUFFER_MIN && j.buffering {
		j.playoutHead = j.packets[0].Timestamp
		j.buffering = false
	}
}

func (j *JitterBuffer) Pop() (*JitterPacket, error) {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	if j.buffering {
		return nil, ErrJitterBuffering
	}

	if len(j.packets) == 0 {
		return nil, ErrJitterUnderrun
	}

	var packet *JitterPacket = j.packets[0]

	for _, p := range j.packets {
		if p.Timestamp == j.playoutHead {
			packet = p
			break
		}
	}

	if packet == nil {
		return nil, ErrJitterUnderrun
	}

	if packet.Timestamp > j.playoutHead {
		j.playoutHead = packet.Timestamp
		return nil, ErrJitterMissing
	}

	j.playoutHead++

	log.Printf("playout head is %d frames behind", j.packets[len(j.packets)-1].Timestamp-j.playoutHead)

	return packet, nil
}
