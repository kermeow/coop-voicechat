package coop

type BridgeEvent int

const (
	BridgeConnect BridgeEvent = iota
	BridgeDisconnect
)
