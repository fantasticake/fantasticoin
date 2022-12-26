package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ToBytes(a any) []byte {
	var buf bytes.Buffer
	HandleErr(gob.NewEncoder(&buf).Encode(a))
	return buf.Bytes()
}

func FromBytes(a any, b []byte) {
	HandleErr(gob.NewDecoder(bytes.NewReader(b)).Decode(a))
}

func Hash(a any) string {
	hash := sha256.Sum256([]byte(fmt.Sprint(a)))
	return fmt.Sprintf("%x", hash)
}
