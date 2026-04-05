package coop

import (
	"coop-voicechat/audio"
	"time"
)

const POLL_FREQUENCY = 60
const POLL_INTERVAL = 1000 / POLL_FREQUENCY
const SLEEPING_POLL_INTERVAL = 100
const TIMEOUT_THRESHOLD = POLL_FREQUENCY / 2

type Bridge struct {
	Active   bool
	Recorder *audio.Recorder
	Running  bool

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

		sendFS: send_modfs,
		recvFS: recv_modfs,
	}
}

func (b *Bridge) activate() {
	if b.Active {
		return
	}
	b.Active = true

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

func (b *Bridge) deactivate() {
	if !b.Active {
		return
	}
	b.Active = false

	if b.Recorder != nil {
		b.Recorder.Stop()
	}
}

func (b *Bridge) poll() bool {
	return false
}

func (b *Bridge) recv() {

}

func (b *Bridge) send() {

}

func (b *Bridge) update() {
	if b.poll() {
		b.recv()
	}
	b.send()
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
			interval = SLEEPING_POLL_INTERVAL * time.Millisecond
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
