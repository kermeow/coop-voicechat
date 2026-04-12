package coop

import "github.com/quartercastle/vector"

type vec = vector.Vector

type Player struct {
	LocalIndex uint8

	HeadPosition vec

	CurrentLevel uint16
	CurrentArea  uint16
	CurrentRoom  uint16

	Cap        uint8
	WaterLevel uint16
}

var LocalPlayer *Player = &Player{}
