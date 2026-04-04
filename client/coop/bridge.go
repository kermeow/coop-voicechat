package coop

import "time"

const POLL_FREQUENCY = 60
const POLL_INTERVAL = 1000 / POLL_FREQUENCY
const SLEEPING_POLL_INTERVAL = 1000

type Bridge struct {
	Active bool

	syncFails  int
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
		Active:     false,
		syncFails:  0,
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
		if b.Active {
			b.syncFails++
			if b.syncFails > POLL_FREQUENCY {
				b.Active = false
			}
		}
		return
	}

	b.Active = true
	b.syncFails = 0

	// todo: write data to send_modfs
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

	return true
}
