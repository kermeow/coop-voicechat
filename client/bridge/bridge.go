package bridge

import (
	"context"
	"coop-voicechat/audio"
	"coop-voicechat/modfs"
	"log"
	"time"
)

const BRIDGE_VERSION uint16 = 2

const UPDATE_INTERVAL = 33 // 1/30

type Bridge struct {
	Connected bool

	SendFs *modfs.ModFs
	RecvFs *modfs.ModFs

	Event chan BridgeEvent

	audio *audio.AudioBridge

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

	return &Bridge{
		Connected: false,

		SendFs: send_modfs,
		RecvFs: recv_modfs,

		Event: make(chan BridgeEvent),

		audio: audio.NewAudioBridge(),

		updTicker: time.NewTicker(time.Millisecond * UPDATE_INTERVAL),

		syncLocalFrame:      1,
		syncRemoteFrame:     0,
		syncRemoteAckFrame:  0,
		syncLastRemoteFrame: 0,
		syncTimeoutCounter:  0,
	}
}

func (b *Bridge) connect() {
	if b.Connected {
		return
	}
	b.Connected = true
	b.Event <- BridgeConnect
}

func (b *Bridge) disconnect() {
	if !b.Connected {
		return
	}
	b.Connected = false
	b.Event <- BridgeDisconnect
}

func (b *Bridge) Run(ctx context.Context) {
	log.Println("Bridge running")

	go b.audio.Run(ctx)

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
	}

	syncFile := b.SendFs.Create("sync")
	syncFile.WriteUint16(BRIDGE_VERSION)
	syncFile.WriteUint32(b.syncLocalFrame)
	syncFile.WriteUint32(b.syncRemoteFrame)

	b.SendFs.Write()
}
