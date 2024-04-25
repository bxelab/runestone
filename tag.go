// Copyright 2024 The BxELab studyzy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runestone

import (
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
func (tag Tag) String() string {
	return fmt.Sprintf("Tag(%d)", tag)
}

func TagTake[T any](t Tag, fields map[Tag][]uint128.Uint128, with func([]uint128.Uint128) (*T, error), n ...int) (*T, error) {
	field, ok := fields[t]
	if !ok {
		return nil, errors.New("tag not found in fields")
	}
	N := len(field)
	if N == 0 {
		return nil, errors.New("no values for tag in fields")
	}
	if len(n) > 0 {
		N = n[0]
	}
	//check filed length
	if len(field) < N {
		return nil, errors.New("field length is less than N")
	}
	values := make([]uint128.Uint128, N)
	//for i := 0; i < N; i++ {
	//	if len(field)>i {
	//		values[i] = field[i]
	//	}
	//}
	copy(values, field)

	value, err := with(values)
	if err != nil {
		return nil, err
	}

	if len(field) >= N {
		field = field[N:]
	} else {
		field = []uint128.Uint128{}
	}
	if len(field) == 0 {
		delete(fields, t)
	} else {
		fields[t] = field
	}

	return value, nil
}

func (t Tag) Encode(values []uint128.Uint128, payload *[]byte) {
	for _, value := range values {
		*payload = append(*payload, t.Byte())
		*payload = append(*payload, EncodeUint128(value)...)
	}
}
