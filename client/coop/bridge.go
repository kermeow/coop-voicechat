package coop

import (
	"encoding/binary"
	"time"
)

const POLL_FREQUENCY = 60
const POLL_INTERVAL = 1000 / POLL_FREQUENCY
const SLEEPING_POLL_INTERVAL = 1000
const TIMEOUT_THRESHOLD = POLL_FREQUENCY / 2

type Bridge struct {
	Active bool

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

func (b *Bridge) poll() {
	b.recv_modfs.Read(false)
	if !b.recvSync() {
		b.Active = false
		return
	}

	b.Active = true

	if (b.clientFlushed) {
		
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

	ack := binary.BigEndian.Uint32(syncData)

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
	binary.BigEndian.PutUint32(syncFile.Data, b.localSyncId)
}
