package packet

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type ActionData struct {
	UserID    uint32
	SkillName string
	Key1      uint32
}

const actionDataMinLength = 12

func parseAction(data []byte) (ActionData, error) {
	if len(data) < actionDataMinLength {
		return ActionData{}, fmt.Errorf("action packet too short: %d bytes, need at least %d", len(data), actionDataMinLength)
	}

	userID := binary.LittleEndian.Uint32(data[0:4])
	skillNameLen := binary.LittleEndian.Uint32(data[8:12])
	skillNameBytes := removeNullBytes(data[12 : 12+skillNameLen])
	skillName := strings.TrimSpace(string(skillNameBytes))
	key1 := binary.LittleEndian.Uint32(data[12+skillNameLen+8 : 12+skillNameLen+12])

	return ActionData{
		UserID:    userID,
		SkillName: skillName,
		Key1:      key1,
	}, nil
}

func removeNullBytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	writeIdx := 0
	for readIdx := 0; readIdx < len(data); readIdx++ {
		if data[readIdx] != 0 {
			data[writeIdx] = data[readIdx]
			writeIdx++
		}
	}

	return data[:writeIdx]
}
