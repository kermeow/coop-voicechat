package bridge

import "log"

func (b *Bridge) poll() bool {
	b.RecvFs.Read(false)

	syncFile, err := b.RecvFs.Get("sync")
	if err != nil {
		return false
	}

	lastActive := b.Connected
	lastRemoteFrame := b.syncRemoteFrame

	syncFile.Cursor = 4

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
