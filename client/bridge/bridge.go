package bridge

import (
	"coop-voicechat/config"
	"coop-voicechat/modfs"
	"log"
	"time"
)

const BRIDGE_VERSION uint16 = 1

const POLL_INTERVAL = 33           // 1/30
const POLL_INTERVAL_INACTIVE = 100 // 1/10

type Bridge struct {
	Running   bool
	Connected bool
	Options   *config.Config

	SendFs *modfs.ModFs
	RecvFs *modfs.ModFs

	Event chan BridgeEvent

	syncLocalFrame      uint32
	syncRemoteFrame     uint32
	syncRemoteAckFrame  uint32
	syncLastRemoteFrame uint32
	syncTimeoutCounter  uint8
}

func NewBridge(options *config.Config) *Bridge {
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
		Running:   false,
		Connected: false,
		Options:   options,

		SendFs: send_modfs,
		RecvFs: recv_modfs,

		Event: make(chan BridgeEvent),

		syncLocalFrame:      1,
		syncRemoteFrame:     0,
		syncRemoteAckFrame:  0,
		syncLastRemoteFrame: 0,
		syncTimeoutCounter:  0,
	}
}

func (b *Bridge) Start() {
	if b.Running {
		return
	}

	log.Println("Bridge running")

	b.Running = true

	for b.Running {
		b.update()

		interval := POLL_INTERVAL * time.Millisecond
		if !b.Connected {
			interval = POLL_INTERVAL_INACTIVE * time.Millisecond
		}
		time.Sleep(interval)
	}
}

func (b *Bridge) Stop() {
	if !b.Running {
		return
	}

	log.Println("Bridge stopping")

	b.Running = false
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
