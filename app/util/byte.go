package util

import (
	"bytes"
	"encoding/gob"
)

// FromBytes converts the given bytes to object.
func FromBytes(b []byte, v interface{}) error {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	return dec.Decode(v)
}

// ToBytes converts the given object to bytes.
func ToBytes(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
