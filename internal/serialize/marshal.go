package serialize

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

var (
	errNoTag = errors.New("no tage value")
)

func unmarshalString(data []byte) string {
	for idx := range data {
		if data[idx] == 0 {
			data = data[:idx]
			break
		}
	}
	return string(data)
}

func MarshalU32(data []byte, val uint32) {
	binary.LittleEndian.PutUint32(data, val)
}

func GetTagValueRequired(t reflect.StructField, name string) (int, error) {
	tag, ok := t.Tag.Lookup(name)
	if !ok {
		return 0, fmt.Errorf("field missing required tag: %s", name)
	}
	return strconv.Atoi(tag)
}

func GetTagValue(t reflect.StructField, name string) (int, error) {
	if tag, ok := t.Tag.Lookup(name); ok {
		return strconv.Atoi(tag)
	}

	return 0, errNoTag
}

func MarshalPaddedString(buffer []byte, s string) {
	for idx := range buffer {
		buffer[idx] = ' '
	}

	b := []byte(s)

	sz := len(b)
	if sz > len(buffer) {
		sz = len(buffer)
	}

	for idx := 0; idx < sz; idx++ {
		buffer[idx] = b[idx]
	}
}
