package coop

import (
	"log"
	"time"
)

const BRIDGE_VERSION uint16 = 1

const POLL_INTERVAL = 33           // 1/30
const POLL_INTERVAL_INACTIVE = 100 // 1/10

type Bridge struct {
	Connected bool
	Running   bool

	audio *AudioBridge

	syncLocalFrame      uint32
	syncRemoteFrame     uint32
	syncRemoteAckFrame  uint32
	syncLastRemoteFrame uint32

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

		syncLocalFrame:      1,
		syncRemoteFrame:     0,
		syncRemoteAckFrame:  0,
		syncLastRemoteFrame: 0,

		sendFS: send_modfs,
		recvFS: recv_modfs,
	}
}

func (b *Bridge) connect() {
	if b.Connected {
		return
	}
	b.Connected = true

	log.Println("Bridge connected")
}

func (b *Bridge) disconnect() {
	if !b.Connected {
		return
	}
	b.Connected = false

	log.Println("Bridge disconnected")
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
	b.syncRemoteFrame, _ = syncFile.ReadUint32()
	b.syncRemoteAckFrame, _ = syncFile.ReadUint32()

	ackFrameValid := b.syncRemoteAckFrame <= b.syncLocalFrame
	ackFrameThreshold := b.syncLocalFrame-b.syncRemoteAckFrame < 6

	b.syncLocalFrame++
	b.syncLastRemoteFrame = lastRemoteFrame

	if b.syncRemoteFrame > lastRemoteFrame {
		// active means coop is running and acknowledging us
		shouldActivate := ackFrameValid && ackFrameThreshold
		if shouldActivate && !lastActive {
			b.connect()
		}
		return b.Connected
	}

	if lastActive && !(ackFrameValid && ackFrameThreshold) {
		b.disconnect()
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
	syncFile.WriteUint32(b.syncLocalFrame)
	syncFile.WriteUint32(b.syncRemoteFrame)

	b.sendFS.Write()
}

func (b *Bridge) Run() {
	if b.Running {
		return
	}

	b.Running = true

	b.audio = NewAudioBridge(b)
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

	b.audio = nil

	b.Running = false
}
