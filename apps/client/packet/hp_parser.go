package packet

import (
	"encoding/binary"
	"fmt"
)

type HPData struct {
	TargetID uint32
	Prev     uint32
	Current  uint32
	Damage   uint32
}

const hpDataMinLength = 20

func parseHP(data []byte) (HPData, error) {
	if len(data) < hpDataMinLength {
		return HPData{}, fmt.Errorf("HP packet too short: got %d bytes, need at least %d bytes", len(data), hpDataMinLength)
	}

	targetID := binary.LittleEndian.Uint32(data[0:4])
	// Skip bytes 4-7 (unused)
	prev := binary.LittleEndian.Uint32(data[8:12])
	// Skip bytes 12-15 (unused)
	current := binary.LittleEndian.Uint32(data[16:20])

	damage := uint32(0)
	if prev > current {
		damage = prev - current
	}

	return HPData{
		TargetID: targetID,
		Prev:     prev,
		Current:  current,
		Damage:   damage,
	}, nil
}
