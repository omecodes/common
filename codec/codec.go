package codec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

var Default Codec

func init() {
	Default = &gsonCodec{}
}

func GSONDecode(data []byte, o interface{}) error {
	buff := bytes.NewReader(data)
	return gob.NewDecoder(buff).Decode(o)
}

func GSONEncode(o interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buff).Encode(o)
	return buff.Bytes(), err
}

func JSONDecode(data []byte, o interface{}) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(o)
}

func JSONEncode(o interface{}) ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	err := json.NewEncoder(buff).Encode(o)
	return buff.Bytes(), err
}

type Codec interface {
	Encoder
	Decoder
}

type Encoder interface {
	Encode(o interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(data []byte, o interface{}) error
}

type gsonCodec struct{}

func (g *gsonCodec) Encode(o interface{}) ([]byte, error) {
	return GSONEncode(o)
}

func (g *gsonCodec) Decode(data []byte, o interface{}) error {
	return GSONDecode(data, o)
}

func NewGSONCodec() Codec {
	return &gsonCodec{}
}
