package go_runestone

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"lukechampine.com/uint128"
)

type Tag uint8

const (
	TagBody         Tag = 0
	TagFlags        Tag = 2
	TagRune         Tag = 4
	TagPremine      Tag = 6
	TagCap          Tag = 8
	TagAmount       Tag = 10
	TagHeightStart  Tag = 12
	TagHeightEnd    Tag = 14
	TagOffsetStart  Tag = 16
	TagOffsetEnd    Tag = 18
	TagMint         Tag = 20
	TagPointer      Tag = 22
	TagCenotaph     Tag = 126
	TagDivisibility Tag = 1
	TagSpacers      Tag = 3
	TagSymbol       Tag = 5
	TagNop          Tag = 127
)

func NewTag(u uint128.Uint128) Tag {
	return Tag(u.Lo)

}
func (tag Tag) Byte() byte {
	return byte(tag)
}

type HashMap map[Tag][]uint128.Uint128

func TagTake[T any](t Tag, fields map[Tag][]uint128.Uint128, with func([]uint128.Uint128) (*T, error)) (*T, error) {
	field, ok := fields[t]
	if !ok {
		return nil, errors.New("field not found")
	}

	if len(field) == 0 {
		return nil, errors.New("field is empty")
	}

	value, err := with(field)
	if err != nil {
		return nil, err
	}

	delete(fields, t)

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
