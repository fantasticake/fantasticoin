package utils

import (
	"bytes"
	"encoding/gob"
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
