package codec

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

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
