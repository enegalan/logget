package flags

import (
	"fmt"
	"strconv"
	"strings"
)

type SimpleFlag[T string | int | int64] struct {
	Value    T
	TypeName string
}

func (f *SimpleFlag[T]) String() string {
	var v any = f.Value
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	}
	return ""
}

func (f *SimpleFlag[T]) Set(value string) error {
	var v any = f.Value
	switch v.(type) {
	case string:
		f.Value = any(value).(T)
		return nil
	case int:
		val, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid %s value: %v", f.TypeName, err)
		}
		f.Value = any(int(val)).(T)
		return nil
	case int64:
		val, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid %s value: %v", f.TypeName, err)
		}
		f.Value = any(val).(T)
		return nil
	}
	return fmt.Errorf("unsupported type")
}

func (f *SimpleFlag[T]) Type() string {
	return f.TypeName
}

func (f *SimpleFlag[T]) Get() T {
	return f.Value
}

func (f *SimpleFlag[T]) Empty() bool {
	var v any = f.Value
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case int:
		return val == 0
	case int64:
		return val == 0
	}
	return true
}
