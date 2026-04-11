package coop

import "github.com/quartercastle/vector"

type vec = vector.Vector

type Player struct {
	LocalIndex uint8

	HeadPosition vec

	CurrentLevel uint8
	CurrentArea  uint8
	CurrentRoom  uint8

	Cap        uint8
	WaterLevel uint8
}

var LocalPlayer *Player = &Player{}
