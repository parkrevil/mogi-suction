package packet

import (
	"bytes"
	"encoding/binary"
)

var (
	startDelimiter        = []byte{0x68, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	startDelimiterLength  = len(startDelimiter)
	endDelimiter          = []byte{0xe3, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	endDelimiterLength    = len(endDelimiter)
	segmentMetadataLength = 9
)

type AnalyzedData struct {
	Type    int
	Content []byte
}

func AnalyzePayload(payload []byte) []AnalyzedData {
	payloadLength := len(payload)

	if payloadLength == 0 {
		return nil
	}

	var analyzed []AnalyzedData
	consumed := 0

	for consumed < payloadLength {
		relStart := bytes.Index(payload[consumed:], startDelimiter)
		if relStart < 0 {
			break
		}
		startIdx := consumed + relStart

		scanFrom := startIdx + startDelimiterLength
		if scanFrom > payloadLength {
			break
		}
		relEnd := bytes.Index(payload[scanFrom:], endDelimiter)
		if relEnd < 0 {
			break
		}
		endIdx := scanFrom + relEnd

		for segStart := scanFrom; ; {
			metaEnd := segStart + segmentMetadataLength
			if metaEnd > endIdx {
				break
			}

			dataType := int(binary.LittleEndian.Uint32(payload[segStart : segStart+4]))
			dataLength := int(binary.LittleEndian.Uint32(payload[segStart+4 : segStart+8]))
			dataEncoding := payload[segStart+8]

			if dataType == 0 {
				break
			}

			segStart = metaEnd + dataLength

			if segStart > endIdx {
				break
			}

			if dataEncoding != 0 {
				//println("data encoding: ", dataEncoding)
				continue
			}

			analyzed = append(analyzed, AnalyzedData{
				Type:    dataType,
				Content: payload[metaEnd:segStart],
			})
		}

		consumed = endIdx + endDelimiterLength
	}

	return analyzed
}

func analyzePayload(payload []byte) {
	packets := AnalyzePayload(payload)
	for _, packet := range packets {
		println(packet.Type)
		switch packet.Type {
		case 10308:
			parseAttack(packet.Content)
		case 100041:
			parseAction(packet.Content)
		case 10299:
			parseDamage(packet.Content)
		case 100178:
			parseHPChanged(packet.Content)
		case 10701, 10719:
			parseSelfDamage(packet.Content)
		case 100321, 100322:
			parseItem(packet.Content)
		}
	}
}
