package encodedecode

import (
	"encoding/base64"
	"log"
)

func EncodeData(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)
	_, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Printf("EncodeData failed input = %s, encodes as %s, err = %s\n", string(data), encoded, err)
	}
	return encoded
}

func DecodeData(data string) string {
	rawDecodedText, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		log.Printf("DecodeData failed input = %s, err = %s", data, err)
		return ""
	}
	return string(rawDecodedText)
}
