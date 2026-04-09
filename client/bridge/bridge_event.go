package bridge

type BridgeEvent int

const (
	BridgeConnect BridgeEvent = iota
	BridgeDisconnect
)
