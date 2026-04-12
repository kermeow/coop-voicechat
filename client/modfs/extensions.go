package modfs

import (
	"coop-voicechat/coop"

	"github.com/quartercastle/vector"
)

type vec = vector.Vector

// Non-standard ModFs io

func (f *ModFsFile) ReadVec(n int) (vec, error) {
	v := vec{}
	for range n {
		x, err := f.ReadFloat64()
		if err != nil {
			return v, err
		}
		v = append(v, x)
	}
	return v, nil
}

func (f *ModFsFile) ReadPlayer(p *coop.Player) error {
	pos, _ := f.ReadVec(3)

	level, _ := f.ReadUint8()
	area, _ := f.ReadUint8()
	room, _ := f.ReadUint8()

	cap, _ := f.ReadUint8()
	water, _ := f.ReadUint8()

	// todo: check errors

	p.HeadPosition = pos

	p.CurrentLevel = level
	p.CurrentArea = area
	p.CurrentRoom = room

	p.Cap = cap
	p.WaterLevel = water

	return nil
}
