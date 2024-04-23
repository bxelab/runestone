package go_runestone

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"lukechampine.com/uint128"
)

type Tag int

const (
	TagBody         = 0
	TagFlags        = 2
	TagRune         = 4
	TagPremine      = 6
	TagCap          = 8
	TagAmount       = 10
	TagHeightStart  = 12
	TagHeightEnd    = 14
	TagOffsetStart  = 16
	TagOffsetEnd    = 18
	TagMint         = 20
	TagPointer      = 22
	TagCenotaph     = 126
	TagDivisibility = 1
	TagSpacers      = 3
	TagSymbol       = 5
	TagNop          = 127
)

type HashMap map[Tag][]uint128.Uint128

func (t Tag) Take(fields *HashMap, with func([]uint128.Uint128) (uint128.Uint128, error)) (uint128.Uint128, error) {
	field, ok := (*fields)[t]
	if !ok {
		return uint128.Zero, errors.New("field not found")
	}

	if len(field) == 0 {
		return uint128.Zero, errors.New("field is empty")
	}

	value, err := with(field)
	if err != nil {
		return uint128.Zero, err
	}

	(*fields)[t] = (*fields)[t][1:]

	if len((*fields)[t]) == 0 {
		delete(*fields, t)
	}

	return value, nil
}

func (t Tag) Encode(values []uint128.Uint128, payload *[]byte) {
	for _, value := range values {
		t.encodeToSlice(payload)
		encodeToSlice(value, payload)
	}
}

func (t Tag) encodeToSlice(payload *[]byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, t)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	*payload = append(*payload, buf.Bytes()...)
}

func encodeToSlice(value uint128.Uint128, payload *[]byte) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	*payload = append(*payload, buf.Bytes()...)
}
