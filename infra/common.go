package infra

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func WrapErr(err error, msg string) error {
	if err == nil {
		return err
	}
	return errors.Wrap(err, msg)
}

// TryUnmarshal 如果src是[]byte 或 string类型，则json.Unmarshal(src, value)
func TryUnmarshal(src, value interface{}) error {
	if src == nil {
		return nil
	}
	var bytes []byte
	switch v := src.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("try unmarshal failed, unknown src type %T", src)
	}
	if len(bytes) > 0 {
		return json.Unmarshal([]byte(bytes), value)
	}
	return nil
}
