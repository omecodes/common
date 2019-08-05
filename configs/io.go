package configs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

func SaveToFile(i interface{}, file string) error {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer([]byte{})
	if err := json.Indent(buffer, marshaled, "", "\t"); err == nil {
		return ioutil.WriteFile(file, buffer.Bytes(), os.ModePerm)
	} else {
		return ioutil.WriteFile(file, marshaled, os.ModePerm)
	}
}

func LoadFromFile(i interface{}, file string) (err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, i)
	return
}
