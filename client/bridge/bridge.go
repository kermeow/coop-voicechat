package bridge

import (
	"context"
	"coop-voicechat/coop"
	"coop-voicechat/modfs"
	"log"
	"time"
)

const BRIDGE_VERSION uint16 = 2
const FILE_HEADER = "smvc"

var FILE_HEADER_BYTES = []byte(FILE_HEADER) // cant be const because go bruh

const UPDATE_INTERVAL = 16 // 1/60

type Bridge struct {
	Connected bool
	Players   map[uint8]*coop.Player

	SendFs *modfs.ModFs
	RecvFs *modfs.ModFs

	Event chan BridgeEvent

	audio      *AudioBridge
	audioFiles map[uint8]string

	updTicker *time.Ticker
	stopChan  chan bool

	syncLocalFrame      uint32
	syncRemoteFrame     uint32
	syncRemoteAckFrame  uint32
	syncLastRemoteFrame uint32
	syncTimeoutCounter  uint8
}

func NewBridge() *Bridge {
	send_modfs, err := modfs.Get("coop-voicechat-recv")
	if err != nil {
		panic(err)
	}
	send_modfs.Write()

	recv_modfs, err := modfs.Get("coop-voicechat")
	if err != nil {
		panic(err)
	}
	recv_modfs.Write()

	b := &Bridge{
		Connected: false,
		Players:   make(map[uint8]*coop.Player),

		SendFs: send_modfs,
		RecvFs: recv_modfs,

		Event: make(chan BridgeEvent),

		audioFiles: make(map[uint8]string),

		updTicker: time.NewTicker(time.Millisecond * UPDATE_INTERVAL),

		syncLocalFrame:      1,
		syncRemoteFrame:     0,
		syncRemoteAckFrame:  0,
		syncLastRemoteFrame: 0,
		syncTimeoutCounter:  0,
	}
	b.audio = NewAudioBridge(b)
	return b
}

func (b *Bridge) removePlayer(localIndex uint8) {
	p := b.Players[localIndex]
	if p == nil {
		return
	}
	b.audio.removePlayer(localIndex)
	delete(b.Players, localIndex)
	delete(b.audioFiles, localIndex)
	log.Println("Player", localIndex, "removed")
}

func (b *Bridge) connect() {
	if b.Connected {
		return
	}
	b.Connected = true
	b.Event <- BridgeConnect

	b.audio.connect()
}

func (b *Bridge) disconnect() {
	if !b.Connected {
		return
	}
	b.Connected = false
	b.Event <- BridgeDisconnect

	for i := range b.Players {
		b.removePlayer(i)
	}
	b.audio.disconnect()
}

func (b *Bridge) Run(ctx context.Context) {
	log.Println("Bridge running")

	go b.audio.run(ctx)

running:
	for {
		select {
		case <-b.updTicker.C:
			b.update()
		case <-ctx.Done():
			b.stop()
			break running
		}
	}
}

func (b *Bridge) stop() {
	log.Println("Bridge stopping")

	b.updTicker.Stop()
	b.disconnect()
}

func (b *Bridge) update() {
	if b.poll() {
		b.recv()
		b.send()
	}

	syncFile := b.SendFs.Create("sync")
	syncFile.WriteBytes(FILE_HEADER_BYTES)
	syncFile.WriteUint16(BRIDGE_VERSION)
	syncFile.WriteUint32(b.syncLocalFrame)
	syncFile.WriteUint32(b.syncRemoteFrame)

	b.SendFs.Write()
}

func (b *Bridge) recv() {
	f, err := b.RecvFs.Get("local_player")
	if err != nil {
		return
	}

	f.Cursor = 4 // skip header
	coop.LocalPlayer.LastState = coop.LocalPlayer.State
	f.ReadPlayer(&coop.LocalPlayer.State)

	f, err = b.RecvFs.Get("players")
	if err != nil {
		return
	}

	f.Cursor = 4 // skip header
	for f.Cursor < len(f.Data) {
		n, _ := f.ReadUint8()
		id, connected := n&0x7f, n&0x80 > 0
		p := b.Players[id]

		if !connected {
			if p != nil {
				b.removePlayer(id)
			}
			continue
		} else if p == nil {
			p = &coop.Player{
				LocalIndex: id,
			}

			l, _ := f.ReadUint8()
			sb, _ := f.ReadBytes(int(l))

			p.LastState = p.State
			f.ReadPlayer(&p.State)

			b.Players[id] = p
			b.audioFiles[id] = string(sb)
			b.audio.addPlayer(p)
			log.Println("Player", id, "added")
		} else {
			f.Cursor++
			f.Cursor += len(b.audioFiles[id])
			f.ReadPlayer(&p.State)
		}
	}

	b.audio.recv()
}

func (b *Bridge) send() {
	b.audio.send()
}
