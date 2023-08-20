package utils

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

var json jsoniter.API

func init() {
	// NOTE(fredbi): attention, this configuration rounds floats down to 6 digits
	// For finer-grained config, see: https://pkg.go.dev/github.com/json-iterator/go#section-readme
	json = jsoniter.ConfigFastest
}

func Decode[T any](rdr io.Reader) (T, error) {
	var receiver T
	dec := newDecoder(rdr)
	err := dec.Decode(&receiver)

	return receiver, err
}

func newDecoder(rdr io.Reader) *jsoniter.Decoder {
	return json.NewDecoder(rdr)
}
