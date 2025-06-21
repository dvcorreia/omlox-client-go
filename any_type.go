package omlox

import (
	"bytes"
	"encoding/json"
	"io"
)

type AnyType json.RawMessage

// MarshalJSON implements json.Marshaler
func (a AnyType) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}
	return json.RawMessage(a).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler
func (a *AnyType) UnmarshalJSON(data []byte) error {
	return (*json.RawMessage)(a).UnmarshalJSON(data)
}

// Reader returns an io.Reader for the JSON data
func (a AnyType) Reader() io.Reader {
	return bytes.NewReader([]byte(a))
}
