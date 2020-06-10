package types

// String ...
type String string

// Int ...
type Int int

func BoolVal(i interface{}) (bool, bool) {
	if i == nil {
		return false, false
	}
	s, ok := i.(bool)
	return s, ok
}
func StringVal(i interface{}) (string, bool) {
	if i == nil {
		return "", false
	}
	str, ok := i.(string)
	return str, ok
}
func IntVal(i interface{}) (int, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return val, true
	}
	if val, ok := i.(int8); ok {
		return int(val), true
	}
	if val, ok := i.(int16); ok {
		return int(val), true
	}
	if val, ok := i.(int32); ok {
		return int(val), true
	}
	if val, ok := i.(int64); ok {
		return int(val), true
	}
	if val, ok := i.(uint8); ok {
		return int(val), true
	}
	if val, ok := i.(uint16); ok {
		return int(val), true
	}
	if val, ok := i.(uint32); ok {
		return int(val), true
	}
	if val, ok := i.(uint64); ok {
		return int(val), true
	}
	if val, ok := i.(float32); ok {
		return int(val), true
	}
	if val, ok := i.(float64); ok {
		return int(val), true
	}

	return 0, false
}
func Int16Val(i interface{}) (int16, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return int16(val), true
	}
	if val, ok := i.(int8); ok {
		return int16(val), true
	}
	if val, ok := i.(int16); ok {
		return val, true
	}
	if val, ok := i.(int32); ok {
		return int16(val), true
	}
	if val, ok := i.(int64); ok {
		return int16(val), true
	}
	if val, ok := i.(uint8); ok {
		return int16(val), true
	}
	if val, ok := i.(uint16); ok {
		return int16(val), true
	}
	if val, ok := i.(uint32); ok {
		return int16(val), true
	}
	if val, ok := i.(uint64); ok {
		return int16(val), true
	}
	if val, ok := i.(float32); ok {
		return int16(val), true
	}
	if val, ok := i.(float64); ok {
		return int16(val), true
	}

	return 0, false
}
func Int32Val(i interface{}) (int32, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return int32(val), true
	}
	if val, ok := i.(int8); ok {
		return int32(val), true
	}
	if val, ok := i.(int16); ok {
		return int32(val), true
	}
	if val, ok := i.(int32); ok {
		return val, true
	}
	if val, ok := i.(int64); ok {
		return int32(val), true
	}
	if val, ok := i.(uint8); ok {
		return int32(val), true
	}
	if val, ok := i.(uint16); ok {
		return int32(val), true
	}
	if val, ok := i.(uint32); ok {
		return int32(val), true
	}
	if val, ok := i.(uint64); ok {
		return int32(val), true
	}
	if val, ok := i.(float32); ok {
		return int32(val), true
	}
	if val, ok := i.(float64); ok {
		return int32(val), true
	}
	return 0, false
}
func Int64Val(i interface{}) (int64, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return int64(val), true
	}
	if val, ok := i.(int8); ok {
		return int64(val), true
	}
	if val, ok := i.(int16); ok {
		return int64(val), true
	}
	if val, ok := i.(int32); ok {
		return int64(val), true
	}
	if val, ok := i.(int64); ok {
		return val, true
	}
	if val, ok := i.(uint8); ok {
		return int64(val), true
	}
	if val, ok := i.(uint16); ok {
		return int64(val), true
	}
	if val, ok := i.(uint32); ok {
		return int64(val), true
	}
	if val, ok := i.(uint64); ok {
		return int64(val), true
	}
	if val, ok := i.(float32); ok {
		return int64(val), true
	}
	if val, ok := i.(float64); ok {
		return int64(val), true
	}

	return 0, false
}
func Uint16Val(i interface{}) (uint16, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return uint16(val), true
	}
	if val, ok := i.(int8); ok {
		return uint16(val), true
	}
	if val, ok := i.(int16); ok {
		return uint16(val), true
	}
	if val, ok := i.(int32); ok {
		return uint16(val), true
	}
	if val, ok := i.(int64); ok {
		return uint16(val), true
	}
	if val, ok := i.(uint8); ok {
		return uint16(val), true
	}
	if val, ok := i.(uint16); ok {
		return val, true
	}
	if val, ok := i.(uint32); ok {
		return uint16(val), true
	}
	if val, ok := i.(uint64); ok {
		return uint16(val), true
	}
	if val, ok := i.(float32); ok {
		return uint16(val), true
	}
	if val, ok := i.(float64); ok {
		return uint16(val), true
	}

	return 0, false
}
func Uint32Val(i interface{}) (uint32, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return uint32(val), true
	}
	if val, ok := i.(int8); ok {
		return uint32(val), true
	}
	if val, ok := i.(int16); ok {
		return uint32(val), true
	}
	if val, ok := i.(int32); ok {
		return uint32(val), true
	}
	if val, ok := i.(int64); ok {
		return uint32(val), true
	}
	if val, ok := i.(uint8); ok {
		return uint32(val), true
	}
	if val, ok := i.(uint16); ok {
		return uint32(val), true
	}
	if val, ok := i.(uint32); ok {
		return val, true
	}
	if val, ok := i.(uint64); ok {
		return uint32(val), true
	}
	if val, ok := i.(float32); ok {
		return uint32(val), true
	}
	if val, ok := i.(float64); ok {
		return uint32(val), true
	}

	return 0, false
}
func UInt64Val(i interface{}) (uint64, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return uint64(val), true
	}
	if val, ok := i.(int8); ok {
		return uint64(val), true
	}
	if val, ok := i.(int16); ok {
		return uint64(val), true
	}
	if val, ok := i.(int32); ok {
		return uint64(val), true
	}
	if val, ok := i.(int64); ok {
		return uint64(val), true
	}
	if val, ok := i.(uint8); ok {
		return uint64(val), true
	}
	if val, ok := i.(uint16); ok {
		return uint64(val), true
	}
	if val, ok := i.(uint32); ok {
		return uint64(val), true
	}
	if val, ok := i.(uint64); ok {
		return val, true
	}
	if val, ok := i.(float32); ok {
		return uint64(val), true
	}
	if val, ok := i.(float64); ok {
		return uint64(val), true
	}

	return 0, false
}
func Float32Val(i interface{}) (float32, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return float32(val), true
	}
	if val, ok := i.(int8); ok {
		return float32(val), true
	}
	if val, ok := i.(int16); ok {
		return float32(val), true
	}
	if val, ok := i.(int32); ok {
		return float32(val), true
	}
	if val, ok := i.(int64); ok {
		return float32(val), true
	}
	if val, ok := i.(uint8); ok {
		return float32(val), true
	}
	if val, ok := i.(uint16); ok {
		return float32(val), true
	}
	if val, ok := i.(uint32); ok {
		return float32(val), true
	}
	if val, ok := i.(uint64); ok {
		return float32(val), true
	}
	if val, ok := i.(float32); ok {
		return val, true
	}
	if val, ok := i.(float64); ok {
		return float32(val), true
	}

	return 0, false
}
func Float64Val(i interface{}) (float64, bool) {
	if i == nil {
		return 0, false
	}

	if val, ok := i.(int); ok {
		return float64(val), true
	}
	if val, ok := i.(int8); ok {
		return float64(val), true
	}
	if val, ok := i.(int16); ok {
		return float64(val), true
	}
	if val, ok := i.(int32); ok {
		return float64(val), true
	}
	if val, ok := i.(int64); ok {
		return float64(val), true
	}
	if val, ok := i.(uint8); ok {
		return float64(val), true
	}
	if val, ok := i.(uint16); ok {
		return float64(val), true
	}
	if val, ok := i.(uint32); ok {
		return float64(val), true
	}
	if val, ok := i.(uint64); ok {
		return float64(val), true
	}
	if val, ok := i.(float32); ok {
		return float64(val), true
	}
	if val, ok := i.(float64); ok {
		return val, true
	}

	return 0, false
}
