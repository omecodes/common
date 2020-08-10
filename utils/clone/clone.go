package clone

import (
	"encoding/json"
	"github.com/gogo/protobuf/proto"
	"reflect"
)

func New(o interface{}) interface{} {
	switch no := o.(type) {
	case proto.Message:
		encoded, _ := proto.Marshal(no)
		cp := reflect.New(reflect.TypeOf(o).Elem())
		i := cp.Interface()
		_ = proto.Unmarshal(encoded, i.(proto.Message))
		return i

	default:
		encoded, _ := json.Marshal(no)
		cp := reflect.New(reflect.TypeOf(o))
		i := cp.Interface()
		_ = json.Unmarshal(encoded, i)
		return i
	}
}
