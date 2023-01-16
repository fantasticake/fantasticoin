package utils

import (
	"reflect"
	"testing"
)

func TestToBytes(t *testing.T) {
	t.Run("should return a slice of bytes", func(t *testing.T) {
		b := ToBytes("test")
		if reflect.TypeOf(b).Kind() != reflect.Slice {
			t.Errorf("got: %v", b)
		}
	})
	t.Run("return value should be able to decode correctly", func(t *testing.T) {
		b := ToBytes("test")
		var v string
		FromBytes(&v, b)
		if v != "test" {
			t.Errorf("Expected: test, Got: %v", v)
		}
	})
}

func TestHash(t *testing.T) {
	h1 := Hash("test")
	h2 := Hash("test")
	if h1 != h2 {
		t.Errorf("should return same value for same input, Hash1: %v, Hash2: %v", h1, h2)
	}
}

func TestToJson(t *testing.T) {
	type test struct {
		Key string
	}
	data := &test{"data"}
	jsonAsB := ToJson(data)
	v := &test{}
	FromJson(v, jsonAsB)
	if !reflect.DeepEqual(data, v) {
		t.Errorf("return value sould be able to decode correctly")
	}
}
