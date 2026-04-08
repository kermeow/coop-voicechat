package coop

import (
	"log"
	"time"
)

const BRIDGE_VERSION uint16 = 1

const POLL_INTERVAL = 33           // 1/30
const POLL_INTERVAL_INACTIVE = 100 // 1/10

type BridgeEvent int

const (
	BridgeConnect BridgeEvent = iota
	BridgeDisconnect
)

type Bridge struct {
	Connected bool
	Running   bool
	Event     chan BridgeEvent

	audio *audioBridge

	syncLocalFrame      uint32
	syncRemoteFrame     uint32
	syncRemoteAckFrame  uint32
	syncLastRemoteFrame uint32
	syncTimeoutCounter  uint8

	sendFS *ModFS
	recvFS *ModFS
}

func NewBridge() *Bridge {
	send_modfs, err := ModFSGet("coop-voicechat-recv")
	if err != nil {
		panic(err)
	}
	send_modfs.Write()

	recv_modfs, err := ModFSGet("coop-voicechat")
	if err != nil {
		panic(err)
	}
	recv_modfs.Write()

	return &Bridge{
		Connected: false,
		Running:   false,
		Event:     make(chan BridgeEvent),

		syncLocalFrame:      1,
		syncRemoteFrame:     0,
		syncRemoteAckFrame:  0,
		syncLastRemoteFrame: 0,
		syncTimeoutCounter:  0,

		sendFS: send_modfs,
		recvFS: recv_modfs,
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
	b.recvFS.Read(false)

	syncFile, err := b.recvFS.Get("sync")
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

func (b *Bridge) recv() {
	b.audio.recv()
}

func (b *Bridge) send() {
	b.audio.send()
}

func (b *Bridge) update() {
	if b.poll() {
		b.recv()
		b.send()
	}

	syncFile := b.sendFS.Create("sync")
	syncFile.WriteUint16(BRIDGE_VERSION)
	syncFile.WriteUint32(b.syncLocalFrame)
	syncFile.WriteUint32(b.syncRemoteFrame)

	b.sendFS.Write()
}

func (b *Bridge) Run() {
	if b.Running {
		return
	}

	log.Println("Bridge running")

	b.Running = true

	b.audio = newAudioBridge(b)
	b.audio.Run()

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
	b.audio = nil
}
