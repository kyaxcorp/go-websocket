package msg

import (
	"encoding/json"
	"strconv"
)

func TextToBytes(message string) []byte {
	return []byte(strconv.Itoa(Text) + message)
}

func TextBytesToBytes(message []byte) []byte {
	return append([]byte(strconv.Itoa(Text)), message...)
}

func JsonToBytes(message interface{}) ([]byte, error) {
	encoded, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return append([]byte(strconv.Itoa(Text)), encoded...), nil
}

func ToBinary(message []byte) []byte {
	return append([]byte(strconv.Itoa(Binary)), message...)
}
