package codec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

type Encoder interface {
	Encode(o interface{}) ([]byte, error)
}

type Decoder interface {
	Decode(data []byte, o interface{}) error
}

type Codec interface {
	Encoder
	Decoder
}

var Default Codec
var Gob Codec
var Json Codec

func init() {
	Default = &gobCodec{}
	Gob = &gobCodec{}
	Json = &jsonc{}
}

func GobDecode(data []byte, o interface{}) error {
	buff := bytes.NewReader(data)
	return gob.NewDecoder(buff).Decode(o)
}

func GobEncode(o interface{}) ([]byte, error) {
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

type gobCodec struct{}

func (g *gobCodec) Encode(o interface{}) ([]byte, error) {
	return GobEncode(o)
}

func (g *gobCodec) Decode(data []byte, o interface{}) error {
	return GobDecode(data, o)
}

func NewGSONCodec() Codec {
	return &gobCodec{}
}

type jsonc struct{}

func (j *jsonc) Encode(o interface{}) ([]byte, error) {
	return json.Marshal(o)
}
func (j *jsonc) Decode(data []byte, o interface{}) error {
	return json.Unmarshal(data, o)
}
