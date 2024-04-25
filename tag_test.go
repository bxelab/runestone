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
	"testing"

	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func TestFromUint128(t *testing.T) {
	assert.Equal(t, TagBody, NewTag(uint128.From64(uint64(TagBody))))
	assert.Equal(t, TagFlags, NewTag(uint128.From64(uint64(TagFlags))))
}

//func TestPartialEq(t *testing.T) {
//	assert.Equal(t, TagBody, uint128.From64(TagBody).Lo)
//	assert.Equal(t, TagFlags, uint128.From64(TagFlags).Lo)
//}

func TestTake(t *testing.T) {
	fields := make(map[Tag][]uint128.Uint128)
	fields[TagFlags] = []uint128.Uint128{uint128.From64(3)}

	n, _ := TagTake(TagFlags, fields, func(_ []uint128.Uint128) (*uint128.Uint128, error) {
		return nil, ErrNone
	})
	assert.Nil(t, n)
	//not empty
	assert.NotEqual(t, 0, len(fields))

	value, err := TagTake(TagFlags, fields, func(values []uint128.Uint128) (*uint128.Uint128, error) {
		return &values[0], nil
	})
	assert.NoError(t, err)
	assert.Equal(t, uint128.From64(3), *value)

	assert.Empty(t, fields)

	_, err = TagTake(TagFlags, fields, func(values []uint128.Uint128) (*uint128.Uint128, error) {
		return &values[0], nil
	})
	assert.Error(t, err)
}

func TestTakeLeavesUnconsumedValues(t *testing.T) {
	fields := make(map[Tag][]uint128.Uint128)
	fields[TagFlags] = []uint128.Uint128{uint128.From64(1), uint128.From64(2), uint128.From64(3)}

	assert.Equal(t, 3, len(fields[TagFlags]))

	_, err := TagTake(TagFlags, fields, func(_ []uint128.Uint128) (*uint128.Uint128, error) {
		return nil, ErrNone
	})
	assert.Error(t, err)

	assert.Equal(t, 3, len(fields[TagFlags]))

	value, err := TagTake(TagFlags, fields, func(values []uint128.Uint128) (*[2]uint128.Uint128, error) {
		return &[2]uint128.Uint128{values[0], values[1]}, nil
	}, 2)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint128.Uint128{uint128.From64(1), uint128.From64(2)}, *value)

	assert.Equal(t, 1, len(fields[TagFlags]))

	valueSingle, err := TagTake(TagFlags, fields, func(values []uint128.Uint128) (*uint128.Uint128, error) {
		return &values[0], nil
	})
	assert.NoError(t, err)
	assert.Equal(t, uint128.From64(3), *valueSingle)

	_, exists := fields[TagFlags]
	assert.False(t, exists)
}

func TestEncode(t *testing.T) {
	payload := make([]byte, 0)

	TagFlags.Encode([]uint128.Uint128{uint128.From64(3)}, &payload)
	assert.Equal(t, []byte{2, 3}, payload)

	TagRune.Encode([]uint128.Uint128{uint128.From64(5)}, &payload)
	assert.Equal(t, []byte{2, 3, 4, 5}, payload)

	TagRune.Encode([]uint128.Uint128{uint128.From64(5), uint128.From64(6)}, &payload)
	assert.Equal(t, []byte{2, 3, 4, 5, 4, 5, 4, 6}, payload)
}

func TestBurnAndNopAreOneByte(t *testing.T) {
	payload := make([]byte, 0)
	TagCenotaph.Encode([]uint128.Uint128{uint128.From64(0)}, &payload)
	assert.Equal(t, 2, len(payload))

	payload = make([]byte, 0)
	TagNop.Encode([]uint128.Uint128{uint128.From64(0)}, &payload)
	assert.Equal(t, 2, len(payload))
}
