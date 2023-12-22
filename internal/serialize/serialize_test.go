package serialize

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

type TestSubStructure struct {
	U8Field1 uint8 `offset:"0"`
	U8Field2 uint8 `offset:"1"`
}

type TestStructure struct {
	U8Field     uint8               `offset:"0"`
	U16Field    uint16              `offset:"8"`
	U32Field    uint32              `offset:"16"`
	U64Field    uint64              `offset:"24"`
	ByteField   byte                `offset:"32"`
	StringField string              `offset:"64" length:"64"`
	ArrayField  [5]byte             `offset:"128" length:"5"`
	Sub         [2]TestSubStructure `offset:"136" length:"2" step:"2"`
	//ignoreField uint16
}

func TestSerialize(t *testing.T) {
	srcStruct := TestStructure{
		U8Field:     0xE,
		U16Field:    0xABCD,
		U32Field:    0x0A1B2C3D,
		U64Field:    0x0123456789ABCDEF,
		ByteField:   0xD,
		StringField: "This is my string",
		ArrayField:  [5]byte{0xF1, 0xF2, 0xF3, 0xF4, 0xF5},
		Sub: [2]TestSubStructure{
			{0x1, 0x2},
			{0xA, 0xB},
		},
		// FIXME: use a different compare that ignores unexported fields
		//ignoreField: 0xFFFF,
	}

	s := New(make([]byte, 256, 256))

	err := s.Serialize(&srcStruct)
	if err != nil {
		t.Fatalf("failed serialize: %s", err.Error())
	}

	dstStruct := TestStructure{}

	d := NewDeserializer(s.Get())
	err = d.Deserialize(&dstStruct)

	if err != nil {
		t.Fatalf("failed serialize: %s", err.Error())
	}

	fmt.Printf(hex.Dump(s.Get()))

	// FIXME: use a different compare that ignores unexported fields
	if !reflect.DeepEqual(srcStruct, dstStruct) {
		t.Logf("Src: %+v\n", srcStruct)
		t.Logf("Dst: %+v\n", dstStruct)
		t.Fatalf("structures not the same")
	}

	var test [16]byte
	val := uint16(0x1234)

	s = New(test[:])
	err = s.Serialize(&val)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Printf("A: %+v\n", test)
}
