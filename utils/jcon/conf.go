package jcon

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

// Map wraps map type
type Map map[string]interface{}

func (item Map) getItem(key string) (interface{}, bool) {
	if !strings.Contains(key, "/") {
		i, ok := item[key]
		return i, ok
	}

	temp := item
	splits := strings.Split(key, "/")

	for i, s := range splits {
		o, exists := temp[s]
		if !exists {
			return nil, false
		}

		if i == len(splits)-1 {
			return o, true
		}

		if vItem, ok := o.(Map); ok {
			temp = vItem

		} else if vItem, ok := o.(map[string]interface{}); ok {
			temp = vItem

		} else {
			return nil, false
		}
	}
	return temp, true
}

func (item Map) Get(key string) interface{} {
	v, _ := item.getItem(key)
	return v
}

func (item Map) Del(key string) bool {
	if !strings.Contains(key, "/") {
		delete(item, key)
		return true
	}

	temp := item
	var lastKeyItem string

	splits := strings.Split(key, "/")
	for len(splits) > 1 {
		lastKeyItem = splits[0]
		if lastKeyItem == "" {
			continue
		}

		splits = splits[1:]

		temp = temp.GetConf(lastKeyItem)
		if temp == nil {
			return false
		}
	}
	delete(temp, splits[0])
	return true
}

func (item Map) Set(key string, val interface{}) {
	if !strings.Contains(key, "/") {
		item[key] = val
		return
	}

	temp := item
	var lastKeyItem string

	splits := strings.Split(key, "/")
	for len(splits) > 1 {
		lastKeyItem = splits[0]
		if lastKeyItem == "" {
			continue
		}

		splits = splits[1:]

		tmp := temp.GetConf(lastKeyItem)
		if tmp == nil {
			temp[lastKeyItem] = Map{}
			tmp = temp.GetConf(lastKeyItem)
		}
		temp = tmp
	}
	temp[splits[0]] = val
}

func (item Map) GetConf(key string) Map {
	v, ok := item.getItem(key)
	if !ok {
		return nil
	}

	if vm, ok := v.(Map); ok {
		return vm
	}

	if vm, ok := v.(map[string]interface{}); ok {
		return vm
	}

	return nil
}

func (item Map) GetBool(key string) (bool, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return false, false
	}
	s, ok := i.(bool)
	if ok {
		return s, ok
	}
	return false, false
}
func (item Map) GetString(key string) (string, bool) {
	i, ok := item.getItem(key)
	//log.Println("GetConf: ", key, " => ", i)
	if !ok {
		return "", false
	}

	s, ok := i.(string)
	if ok {
		return s, ok
	}
	return "", false
}
func (item Map) GetInt16(key string) (int16, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Int16Val(i)
}
func (item Map) GetInt32(key string) (int32, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Int32Val(i)
}
func (item Map) GetInt64(key string) (int64, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Int64Val(i)
}
func (item Map) GetUint16(key string) (uint16, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Uint16Val(i)
}
func (item Map) GetUint32(key string) (uint32, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Uint32Val(i)
}
func (item Map) GetUInt64(key string) (uint64, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return UInt64Val(i)
}
func (item Map) GetFloat32(key string) (float32, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Float32Val(i)
}
func (item Map) GetFloat64(key string) (float64, bool) {
	i, ok := item.getItem(key)
	if !ok {
		return 0, false
	}
	return Float64Val(i)
}

//Handle configs as JSON content in file
func (item Map) Save(filename string, mode os.FileMode) error {
	bytes, err := json.MarshalIndent(item, " ", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, bytes, mode)
}

func (item Map) Load(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &item)
}

func (item Map) String() string {
	configsBytes, _ := json.Marshal(item)
	buffer := bytes.NewBuffer([]byte{})
	if err := json.Indent(buffer, configsBytes, "", "\t"); err == nil {
		return string(buffer.Bytes())
	} else {
		return string(configsBytes)
	}
}

func (item Map) Bytes() ([]byte, error) {
	configsBytes, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	err = json.Indent(buffer, configsBytes, "", "\t")
	if err == nil {
		return buffer.Bytes(), nil
	} else {
		return configsBytes, nil
	}
}

//Init load content from json file
func Load(filename string, item *Map) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, item)
}
