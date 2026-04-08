package coop

import (
	"coop-voicechat/config"
	"log"
	"time"
)

const BRIDGE_VERSION uint16 = 1

const POLL_INTERVAL = 33           // 1/30
const POLL_INTERVAL_INACTIVE = 100 // 1/10

type Bridge struct {
	Running   bool
	Connected bool
	Event     chan BridgeEvent
	Options   *config.Config

	SendFs *ModFS
	RecvFs *ModFS

	syncLocalFrame      uint32
	syncRemoteFrame     uint32
	syncRemoteAckFrame  uint32
	syncLastRemoteFrame uint32
	syncTimeoutCounter  uint8
}

func NewBridge(options *config.Config) *Bridge {
	send_modfs, err := ModFsGet("coop-voicechat-recv")
	if err != nil {
		panic(err)
	}
	send_modfs.Write()

	recv_modfs, err := ModFsGet("coop-voicechat")
	if err != nil {
		panic(err)
	}
	recv_modfs.Write()

	return &Bridge{
		Running:   false,
		Connected: false,
		Event:     make(chan BridgeEvent),
		Options:   options,

		SendFs: send_modfs,
		RecvFs: recv_modfs,

		syncLocalFrame:      1,
		syncRemoteFrame:     0,
		syncRemoteAckFrame:  0,
		syncLastRemoteFrame: 0,
		syncTimeoutCounter:  0,
	}
}

func (b *Bridge) event(e BridgeEvent) {
	select {
	case b.Event <- e:
	default:
	}
}

func (b *Bridge) connect() {
	if b.Connected {
		return
	}
	b.Connected = true
	b.event(BridgeConnect)
}

func (b *Bridge) disconnect() {
	if !b.Connected {
		return
	}
	b.Connected = false
	b.event(BridgeDisconnect)
}

func (b *Bridge) poll() bool {
	b.RecvFs.Read(false)

	syncFile, err := b.RecvFs.Get("sync")
	if err != nil {
		return false
	}

	lastActive := b.Connected
	lastRemoteFrame := b.syncRemoteFrame

	syncFile.Cursor = 0

	remoteVersion, _ := syncFile.ReadUint16()
	if remoteVersion != BRIDGE_VERSION {
		return false
	}

	b.syncRemoteFrame, _ = syncFile.ReadUint32()
	b.syncRemoteAckFrame, _ = syncFile.ReadUint32()

	ackFrameValid := b.syncRemoteAckFrame <= b.syncLocalFrame
	ackFrameThreshold := b.syncLocalFrame-b.syncRemoteAckFrame < 6

	b.syncLocalFrame++
	b.syncLastRemoteFrame = lastRemoteFrame

	if b.syncRemoteFrame > lastRemoteFrame {
		// active means coop is running and acknowledging us
		shouldActivate := ackFrameValid && ackFrameThreshold
		if shouldActivate {
			b.syncTimeoutCounter = 0
			if !lastActive {
				b.connect()
				log.Printf("Bridge connected\n")
			}
		}
		return b.Connected
	}

	if lastActive && !(ackFrameValid && ackFrameThreshold) {
		b.syncTimeoutCounter++
		if b.syncTimeoutCounter > 6 {
			log.Printf("Bridge disconnected - av:%t aft:%t slf:%d srf:%d sraf:%d stc:%d\n", ackFrameValid, ackFrameThreshold, b.syncLocalFrame, b.syncRemoteFrame, b.syncRemoteAckFrame, b.syncTimeoutCounter)
			b.disconnect()
		}
	}

	return false
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
	b.Connected = false
}
