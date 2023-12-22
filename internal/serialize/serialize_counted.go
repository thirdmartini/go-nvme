package serialize

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

func NewSerializer(data []byte) *SerializerCounted {
	return &SerializerCounted{
		data:   data,
		offset: 0,
	}
}

type SerializerCounted struct {
	offset int
	data   []byte
}

func (w *SerializerCounted) Length() int {
	return len(w.data)
}

func (w *SerializerCounted) marshalString(offset, length int, s string) {
	b := []byte(s)

	if length != 0 && len(b) > length {
		b = b[:length]
	}

	for idx := range b {
		w.data[offset+idx] = b[idx]
	}
}

func (w *SerializerCounted) serialize(offset, length, step int, data interface{}) (error, int) {
	value := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		typ = typ.Elem()
	}

	//fmt.Printf("Kind: %v | %v\n", value.Kind(), typ.Name())
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() == 0 {
			return nil, 0
		}

		if step == 0 {
			t := reflect.TypeOf(value.Index(0).Addr().Interface())
			e := t.Elem()
			step = int(e.Size())
		}

		for i := 0; i < value.Len(); i += 1 {
			err, _ := w.serialize(offset+(step*i), 0, 0, value.Index(i).Addr().Interface())
			if err != nil {
				return err, 0
			}
		}
		return nil, step * value.Len()

	case reflect.Struct:
		if !value.CanSet() {
			return nil, 0
		}

		ws := NewSerializer(w.data[offset:])
		return ws.Serialize(data)

	default:
		switch v := data.(type) {
		case *uint8:
			w.data[offset] = *v
			return nil, 1

		case *uint16:
			binary.LittleEndian.PutUint16(w.data[offset:], *v)
			return nil, 2

		case *uint32:
			binary.LittleEndian.PutUint32(w.data[offset:], *v)
			return nil, 4

		case *uint64:
			binary.LittleEndian.PutUint64(w.data[offset:], *v)
			return nil, 8

		case *string:
			if length == 0 {
				return fmt.Errorf("need a limit for string types"), 0
			}
			w.marshalString(offset, length, *v)
			return nil, length

		default:
			return fmt.Errorf("desirialize error: unsupported type: %s", reflect.TypeOf(data).String()), 0
		}

	}
}

func (w *SerializerCounted) Serialize(s interface{}) (error, int) {
	parentValue := reflect.ValueOf(s)
	parentType := reflect.TypeOf(s)

	if parentValue.Kind() == reflect.Ptr {
		parentValue = parentValue.Elem()
		parentType = parentType.Elem()
	} else {
		panic("must be ptr")
	}

	switch parentValue.Kind() {
	case reflect.Struct:
		sCount := 0
		for i := 0; i < parentValue.NumField(); i += 1 {
			// ignore unexported fields
			if !parentValue.Field(i).CanSet() {
				continue
			}

			t := parentType.Field(i)

			offset, err := GetTagValue(t, "offset")
			if err != nil {
				return err, w.offset
			}

			if offset >= len(w.data) {
				return fmt.Errorf("write past end of buffer"), w.offset
			}

			length, _ := GetTagValue(t, "length")
			if offset+length > len(w.data) {
				return fmt.Errorf("write past end of buffer"), w.offset
			}

			step, _ := GetTagValue(t, "step")

			err, count := w.serialize(w.offset+offset, length, step, parentValue.Field(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("serialize %s: failed:%s\n", t.Name, err), w.offset
			}

			sCount = offset + count
		}
		w.offset += sCount
	default:
		return w.serialize(w.offset, 0, 0, s)
	}
	return nil, w.offset
}

func (w *SerializerCounted) Get() []byte {
	return w.data
}
