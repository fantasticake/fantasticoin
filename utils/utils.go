package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ToBytes(v any) []byte {
	var buf bytes.Buffer
	HandleErr(gob.NewEncoder(&buf).Encode(v))
	return buf.Bytes()
}

func FromBytes(v any, b []byte) {
	HandleErr(gob.NewDecoder(bytes.NewReader(b)).Decode(v))
}

func Hash(v any) string {
	hash := sha256.Sum256([]byte(fmt.Sprint(v)))
	return fmt.Sprintf("%x", hash)
}

func ToJson(v any) []byte {
	b, err := json.Marshal(v)
	HandleErr(err)
	return b
}

func FromJson(v any, b []byte) {
	HandleErr(json.Unmarshal(b, v))
}
