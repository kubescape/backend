package utils

import (
	"bytes"
	"encoding/json"
)

func EncodeCommandBody[T any](data T) ([]byte, error) {
	buffer := bytes.Buffer{}
	encoder := json.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecodeCommandBody[T any](b []byte) (T, error) {
	buffer := bytes.Buffer{}
	buffer.Write(b)
	decoder := json.NewDecoder(&buffer)
	var data T
	err := decoder.Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}
