package coop

import (
	"coop-voicechat/audio"
	"time"
)

const BRIDGE_VERSION uint16 = 1

const POLL_INTERVAL = 33           // 1/30
const POLL_INTERVAL_INACTIVE = 100 // 1/10

type Bridge struct {
	Active   bool
	Running  bool
	Recorder *audio.Recorder

	syncLocalFrame     uint32
	syncRemoteFrame    uint32
	syncRemoteAckFrame uint32

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
		Active:  false,
		Running: false,

		syncLocalFrame:     1,
		syncRemoteFrame:    0,
		syncRemoteAckFrame: 0,

		sendFS: send_modfs,
		recvFS: recv_modfs,
	}
}

func (b *Bridge) activated() {
	println("connected")
	if b.Recorder == nil {
		recorder, err := audio.NewRecorder()
		if err == nil {
			b.Recorder = recorder
		}
	}
	if b.Recorder != nil {
		go b.Recorder.Start()
	}
}

func (b *Bridge) deactivated() {
	println("disconnected")
	if b.Recorder != nil {
		b.Recorder.Stop()
	}
}

func (b *Bridge) poll() bool {
	b.recvFS.Read(false)

	syncFile, err := b.recvFS.Get("sync")
	if err != nil {
		return false
	}

	lastActive := b.Active
	lastRemoteFrame := b.syncRemoteFrame

	syncFile.Cursor = 0
	b.syncRemoteFrame, _ = syncFile.ReadUint32()
	b.syncRemoteAckFrame, _ = syncFile.ReadUint32()

	ackFrameValid := b.syncRemoteAckFrame <= b.syncLocalFrame
	ackFrameThreshold := b.syncLocalFrame-b.syncRemoteAckFrame < 6

	if b.syncRemoteFrame > lastRemoteFrame {
		// active means coop is running and acknowledging us
		b.Active = ackFrameValid && ackFrameThreshold
		if b.Active && !lastActive {
			b.activated()
		}
	}

	if !(ackFrameValid && ackFrameThreshold) {
		b.Active = false
		if lastActive {
			b.deactivated()
		}
	}

	return b.Active
}

func (b *Bridge) recv() {

}

func (b *Bridge) send() {

}

func (b *Bridge) update() {
	if b.poll() {
		b.recv()
		b.send()
	}

	b.syncLocalFrame++

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

	for b.Running {
		b.update()
		interval := POLL_INTERVAL * time.Millisecond
		if !b.Active {
			interval = POLL_INTERVAL_INACTIVE * time.Millisecond
		}
		time.Sleep(interval)
	}
}

func (b *Bridge) Stop() {
	if !b.Running {
		return
	}

	b.Running = false
}
