package packet

import (
	"bytes"
	"encoding/binary"
	"log"
)

const (
	attackDataType      = 10308
	hpDataType          = 100178
	actionDataType      = 100041
	selfDamageDataType1 = 10701
	selfDamageDataType2 = 10719
	itemDataType1       = 100321
	itemDataType2       = 100322
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
				// TODO Brotli 압축 파싱
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
	datas := AnalyzePayload(payload)
	for _, data := range datas {
		switch data.Type {
		case attackDataType:
			parsed, err := parseAttack(data.Content)
			if err != nil {
				log.Printf("attack error: %v", err)
			} else {
				log.Printf("attack: %+v", parsed)
			}
		case hpDataType:
			parsed, err := parseHP(data.Content)
			if err != nil {
				log.Printf("HP error: %v", err)
			} else {
				log.Printf("HP: %+v", parsed)
			}
		case actionDataType:
			if err := parseAction(data.Content); err != nil {
				log.Printf("parseAction error: %v", err)
			}
		case selfDamageDataType1, selfDamageDataType2:
			if err := parseSelfDamage(data.Content); err != nil {
				log.Printf("parseSelfDamage error: %v", err)
			}
		case itemDataType1, itemDataType2:
			if err := parseItem(data.Content); err != nil {
				log.Printf("parseItem error: %v", err)
			}
		}
	}
}
