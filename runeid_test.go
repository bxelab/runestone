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
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func newRuneId(block uint64, tx uint32) *RuneId {
	return &RuneId{Block: block, Tx: tx}

}
func TestDelta(t *testing.T) {
	expected := []*RuneId{
		newRuneId(3, 1),
		newRuneId(4, 2),
		newRuneId(1, 2),
		newRuneId(1, 1),
		newRuneId(3, 1),
		newRuneId(2, 0),
	}

	sort.Slice(expected, func(i, j int) bool {
		if expected[i].Block == expected[j].Block {
			return expected[i].Tx < expected[j].Tx
		}
		return expected[i].Block < expected[j].Block
	})

	assert.Equal(t, []*RuneId{
		newRuneId(1, 1),
		newRuneId(1, 2),
		newRuneId(2, 0),
		newRuneId(3, 1),
		newRuneId(3, 1),
		newRuneId(4, 2),
	}, expected)

	previous := &RuneId{}
	var deltas [][2]uint64
	for _, id := range expected {
		block, tx, err := previous.Delta(*id)
		assert.NoError(t, err)
		deltas = append(deltas, [2]uint64{block, uint64(tx)})
		previous = id
	}

	assert.Equal(t, [][2]uint64{{1, 1}, {0, 1}, {1, 0}, {1, 1}, {0, 0}, {1, 2}}, deltas)

	previous = &RuneId{}
	var actual []*RuneId
	for _, delta := range deltas {
		block, tx := delta[0], uint32(delta[1])
		next, err := previous.Next(uint128.From64(block), uint128.From64(uint64(tx)))
		assert.NoError(t, err)
		actual = append(actual, next)
		previous = next
	}

	assert.Equal(t, expected, actual)
}
func TestRuneIdDisplay(t *testing.T) {
	r := RuneId{Block: 1, Tx: 2}
	expected := "1:2"
	if r.String() != expected {
		t.Errorf("Expected %s, but got %s", expected, r.String())
	}
}

func TestFromStr(t *testing.T) {
	_, err := RuneIdFromString("123")
	assert.EqualError(t, err, ErrSeparator.Error())

	_, err = RuneIdFromString(":")
	assert.Contains(t, fmt.Sprintf("%v", err), "Block")

	_, err = RuneIdFromString("1:")
	assert.Contains(t, fmt.Sprintf("%v", err), "Transaction")

	_, err = RuneIdFromString(":2")
	assert.Contains(t, fmt.Sprintf("%v", err), "Block")

	_, err = RuneIdFromString("a:2")
	assert.Contains(t, fmt.Sprintf("%v", err), "Block")

	_, err = RuneIdFromString("1:a")
	assert.Contains(t, fmt.Sprintf("%v", err), "Transaction")

	r, err := RuneIdFromString("1:2")
	assert.NoError(t, err)
	assert.Equal(t, newRuneId(1, 2), r)
}
