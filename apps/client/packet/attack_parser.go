package packet

import (
	"encoding/binary"
	"fmt"
)

type AttackData struct {
	UserID   uint32
	TargetID uint32
	Key1     uint32
	Key2     uint32
	Flags    map[string]bool
}

const attackDataLength = 35

var damageFlagDefs = []struct {
	index uint8
	name  string
	mask  byte
}{
	{0, "crit", 1},
	{0, "what1", 2},
	{0, "unguarded", 4},
	{0, "break", 8},
	{0, "what05", 16},
	{0, "what06", 32},
	{0, "first_hit", 64},
	{0, "default_attack", 128},

	{1, "multi_attack", 1},
	{1, "power", 2},
	{1, "fast", 4},
	{1, "dot1", 8},
	{1, "dot2", 128},

	{2, "dot3", 1},

	{3, "add_hit", 8},
	{3, "bleed", 16},
	{3, "dark", 32},
	{3, "fire", 64},
	{3, "holy", 128},

	{4, "ice", 1},
	{4, "electric", 2},
	{4, "poison", 4},
	{4, "mind", 8},
	{4, "dot4", 16},
}

func parseAttack(data []byte) (AttackData, error) {
	if len(data) != attackDataLength {
		return AttackData{}, fmt.Errorf("invalid attack packet size: %d (want %d)", len(data), attackDataLength)
	}

	userID := binary.LittleEndian.Uint32(data[0:4])
	// data[4:8] is present but unused
	targetID := binary.LittleEndian.Uint32(data[8:12])
	// data[12:16] is present but unused
	key1 := binary.LittleEndian.Uint32(data[16:20])
	key2 := binary.LittleEndian.Uint32(data[20:24])

	flagData := data[24:31]

	flags := make(map[string]bool, len(damageFlagDefs))
	for _, def := range damageFlagDefs {
		flags[def.name] = false

		if def.index < uint8(len(flagData)) {
			flags[def.name] = (flagData[def.index] & def.mask) != 0
		}
	}

	return AttackData{
		UserID:   userID,
		TargetID: targetID,
		Key1:     key1,
		Key2:     key2,
		Flags:    flags,
	}, nil
}
