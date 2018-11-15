package sse

import (
	"bytes"
	"encoding/json"
)

func serializeEvent(event interface{}) ([]byte, error) {
	buffer := bytes.Buffer{}
	if _, err := buffer.Write([]byte("data:")); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	if _, err := buffer.Write(payload); err != nil {
		return nil, err
	}

	if _, err := buffer.Write([]byte("\n\n")); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
