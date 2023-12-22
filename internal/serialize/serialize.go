package serialize

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

func New(data []byte) *Serializer {
	return &Serializer{
		data: data,
	}
}

type Serializer struct {
	data []byte
}

func (w *Serializer) Length() int {
	return len(w.data)
}

func (w *Serializer) marshalString(offset, length int, s string) {
	b := []byte(s)

	if length != 0 && len(b) > length {
		b = b[:length]
	}

	for idx := range b {
		w.data[offset+idx] = b[idx]
	}
}

func (w *Serializer) serialize(offset, length, step int, data interface{}) error {
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
			return nil
		}

		if step == 0 {
			t := reflect.TypeOf(value.Index(0).Addr().Interface())
			e := t.Elem()
			step = int(e.Size())
		}

		for i := 0; i < value.Len(); i += 1 {
			err := w.serialize(offset+(step*i), 0, 0, value.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		if !value.CanSet() {
			return nil
		}

		ws := New(w.data[offset:])
		return ws.Serialize(data)

	default:
		switch v := data.(type) {
		case *uint8:
			w.data[offset] = *v

		case *uint16:
			binary.LittleEndian.PutUint16(w.data[offset:], *v)
			//marshalU16(w.data[offset:], *v)

		case *uint32:
			binary.LittleEndian.PutUint32(w.data[offset:], *v)
			//marshalU32(w.data[offset:], *v)

		case *uint64:
			binary.LittleEndian.PutUint64(w.data[offset:], *v)
			//marshalU64(w.data[offset:], *v)

		case *string:
			if length == 0 {
				return fmt.Errorf("need a limit for string types")
			}
			w.marshalString(offset, length, *v)

		default:
			return fmt.Errorf("desirialize error: unsupported type: %s", reflect.TypeOf(data).String())
		}

	}
	return nil
}

func (w *Serializer) Serialize(s interface{}) error {
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
		for i := 0; i < parentValue.NumField(); i += 1 {
			// ignore unexported fields
			if !parentValue.Field(i).CanSet() {
				return nil
			}

			t := parentType.Field(i)

			offset, err := GetTagValueRequired(t, "offset")
			if err != nil {
				return err
			}

			if offset >= len(w.data) {
				return fmt.Errorf("write past end of buffer")
			}

			length, _ := GetTagValue(t, "length")
			if offset+length > len(w.data) {
				return fmt.Errorf("write past end of buffer")
			}

			step, _ := GetTagValue(t, "step")

			err = w.serialize(offset, length, step, parentValue.Field(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("serialize %s: failed:%s\n", t.Name, err)
			}
		}
	default:
		return w.serialize(0, 0, 0, s)
	}
	return nil
}

func (w *Serializer) Get() []byte {
	return w.data
}
