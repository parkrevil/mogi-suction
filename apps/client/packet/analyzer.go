package packet

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/andybalholm/brotli"
)

var (
	payloadType        = []byte{0x68, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	payloadTypeLength  = len(payloadType)
	dataMetadataLength = 9
)

func analyzePayload(payload []byte) {
	payloadLength := len(payload)
	if payloadLength < payloadTypeLength {
		return
	}

	if !bytes.Equal(payload[0:payloadTypeLength], payloadType) {
		return
	}

	if payloadLength < payloadTypeLength+dataMetadataLength {
		return
	}

	dataType := int(binary.LittleEndian.Uint32(payload[payloadTypeLength : payloadTypeLength+4]))
	dataLength := int(binary.LittleEndian.Uint32(payload[payloadTypeLength+4 : payloadTypeLength+8]))
	dataEncoding := payload[payloadTypeLength+8]

	if payloadLength < payloadTypeLength+dataMetadataLength+dataLength {
		return
	}

	data := payload[payloadTypeLength+dataMetadataLength : payloadTypeLength+dataMetadataLength+dataLength]

	if dataEncoding == 1 {
		reader := brotli.NewReader(bytes.NewReader(data))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return
		}
		data = decompressed
	}

	switch dataType {
	case 10308:
		parseAttack(data)
	case 100041:
		parseAction(data)
	case 10299:
		parseDamage(data)
	case 100178:
		parseHPChanged(data)
	case 10701, 10719:
		parseSelfDamage(data)
	case 100321, 100322:
		parseItem(data)
	default:
		println("unknown packet: &x", dataType)
		return
	}
}
