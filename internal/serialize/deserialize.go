package serialize

import (
	"encoding/binary"
	"fmt"
	"log"
	"reflect"
)

func NewDeserializer(data []byte) *Deserializer {
	return &Deserializer{
		data: data,
	}
}

type Deserializer struct {
	data []byte
}

func (w *Deserializer) deserialize(offset, length, step int, data interface{}) error {
	value := reflect.ValueOf(data)
	typ := reflect.TypeOf(data)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		typ = typ.Elem()
	}
	if !value.CanSet() {
		return nil
	}

	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		if value.Len() == 0 {
			return nil
		}

		// if step was not passed in from parent
		// step by the native size
		if step == 0 {
			t := reflect.TypeOf(value.Index(0).Addr().Interface())
			e := t.Elem()
			step = int(e.Size())
		}

		for i := 0; i < value.Len(); i += 1 {
			err := w.deserialize(offset+(step*i), 0, 0, value.Index(i).Addr().Interface())
			if err != nil {
				return err
			}
		}
	case reflect.Struct:
		// ignore unexported fields
		if !value.CanSet() {
			return nil
		}

		return NewDeserializer(w.data[offset:]).Deserialize(data)

	default:
		switch v := data.(type) {
		case *uint8:
			*v = w.data[offset]

		case *uint16:
			//*v = unmarshalU16(w.data[offset:])
			*v = binary.LittleEndian.Uint16(w.data[offset:])

		case *uint32:
			//*v = unmarshalU32(w.data[offset:])
			*v = binary.LittleEndian.Uint32(w.data[offset:])

		case *uint64:
			//*v = unmarshalU64(w.data[offset:])
			*v = binary.LittleEndian.Uint64(w.data[offset:])

		case *string:
			if length == 0 {
				return fmt.Errorf("need a limit for string types")
			}
			*v = unmarshalString(w.data[offset : offset+length])

		default:
			return fmt.Errorf("desirialize error: unsupported type: %s", reflect.TypeOf(data).String())
		}

	}

	return nil
}

func (w *Deserializer) Deserialize(s interface{}) error {
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

			offset, err := GetTagValue(t, "offset")
			if err != nil {
				return err
			}

			if offset >= len(w.data) {
				return fmt.Errorf("read past end of buffer %d:%d", offset, len(w.data))
			}

			length, _ := GetTagValue(t, "length")
			if offset+length > len(w.data) {
				return fmt.Errorf("read past end of buffer %d:%d:%d", offset, length, len(w.data))
			}

			step, _ := GetTagValue(t, "step")

			err = w.deserialize(offset, length, step, parentValue.Field(i).Addr().Interface())
			if err != nil {
				log.Panicf("deserialize %s: failed:%s\n", t.Name, err)
			}
		}
	default:
		return w.deserialize(0, 0, 0, s)

	}
	return nil
}

func (w *Deserializer) Get() []byte {
	return w.data
}
