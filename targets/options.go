package targets

import (
	"fmt"
	"strconv"
)

type Options map[string]string

func (o Options) String(key string) string {
	s, _ := o[key]
	return s
}

func (o Options) Uint64(key string, def uint64) uint64 {
	if s, ok := o[key]; ok {
		v, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return uint64(v)
		}
	}
	return def
}

func (o Options) Int(key string, def int) int {
	if s, ok := o[key]; ok {
		v, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return int(v)
		}
	}
	return def
}

func (o Options) With(key string, data interface{}) Options {
	//fmt.Printf("Kind: %v | %v\n", value.Kind(), typ.Name())
	switch v := data.(type) {
	case uint, uint8, uint16, uint32, uint64, int, int16, int32, int64:
		o[key] = fmt.Sprintf("%d", v)
	case string:
		o[key] = v
	}

	return o
}
