package coop

import (
	"bytes"
	"coop-voicechat/audio"
	"encoding/binary"
	"time"
)

const POLL_FREQUENCY = 60
const POLL_INTERVAL = 1000 / POLL_FREQUENCY
const SLEEPING_POLL_INTERVAL = 100
const TIMEOUT_THRESHOLD = POLL_FREQUENCY / 2

type Bridge struct {
	Active   bool
	Recorder *audio.Recorder

	ackSyncId     uint32
	localSyncId   uint32
	clientFlushed bool

	send_modfs *ModFS
	recv_modfs *ModFS
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
		Active: false,

		ackSyncId:   0,
		localSyncId: 0,

		send_modfs: send_modfs,
		recv_modfs: recv_modfs,
	}
}

func (b *Bridge) Run() {
	for {
		b.poll()
		interval := POLL_INTERVAL * time.Millisecond
		if !b.Active {
			interval = SLEEPING_POLL_INTERVAL * time.Millisecond
		}
		time.Sleep(interval)
	}
}

func (b *Bridge) activated() {
	if b.Recorder == nil {
		recorder, err := audio.NewRecorder()
		if err == nil {
			b.Recorder = recorder
		}
	}
	if b.Recorder != nil {
		go b.Recorder.Start()
	}
	b.localSyncId = 0
}

func (b *Bridge) deactivated() {
	if b.Recorder != nil {
		b.Recorder.Stop()
	}
	b.localSyncId = 0
}

func (b *Bridge) poll() {
	b.recv_modfs.Read(false)
	if !b.recvSync() {
		if b.Active {
			b.deactivated()
			b.Active = false
		}
		return
	}

	if !b.Active {
		b.activated()
		b.Active = true
	}

	if b.clientFlushed {
		if b.Recorder != nil && b.Recorder.Running {
			buf := bytes.Buffer{}
			opusFrames, err := b.Recorder.Read()
			if err == nil {
				for _, v := range opusFrames {
					data := make([]byte, 4+len(v))
					binary.NativeEndian.PutUint32(data, uint32(len(v)))
					copy(data[4:], v)
					buf.Write(data)
				}
			}
			file := b.send_modfs.Create("recording")
			file.Data = buf.Bytes()
		}
	}

	b.sendSync()
	b.send_modfs.Write()
}

func (b *Bridge) recvSync() bool {
	syncFile, err := b.recv_modfs.Get("sync")
	if err != nil {
		return false
	}

	syncData := syncFile.Data
	if len(syncData) == 0 {
		return false
	}

	ack := binary.NativeEndian.Uint32(syncData)

	b.clientFlushed = ack > b.ackSyncId
	b.ackSyncId = ack

	if b.ackSyncId+TIMEOUT_THRESHOLD <= b.localSyncId {
		return false
	}

	return true
}

func (b *Bridge) sendSync() {
	b.localSyncId++

	syncFile := b.send_modfs.Create("sync")
	syncFile.Data = make([]byte, 4)
	binary.NativeEndian.PutUint32(syncFile.Data, b.localSyncId)
}
