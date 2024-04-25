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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func TestMaxSpacers(t *testing.T) {
	var r strings.Builder

	maxRune := Rune{Value: uint128.Max}

	for i, c := range maxRune.String() {
		if i > 0 {
			r.WriteRune('•')
		}
		r.WriteRune(c)
	}
	t.Logf("rune: %s", r.String())
	spacedRune, err := SpacedRuneFromString(r.String())
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, MaxSpacers, spacedRune.Spacers)
}

func TestSupply(t *testing.T) {
	caseFunc := func(premine *uint128.Uint128, terms *Terms, expected *uint128.Uint128) {
		e := &Etching{
			Premine: premine,
			Terms:   terms,
		}
		assert.Equal(t, expected, e.Supply())
	}

	caseFunc(nil, nil, uint128From(0))
	caseFunc(uint128From(0), nil, uint128From(0))
	caseFunc(uint128From(1), nil, uint128From(1))
	caseFunc(uint128From(1), &Terms{Amount: nil, Cap: nil}, uint128From(1))
	caseFunc(nil, &Terms{Amount: nil, Cap: nil}, uint128From(0))
	half := uint128.Max.Div64(2)
	half1 := uint128.Max.Div64(2).Add64(1)
	caseFunc(&half1, &Terms{Amount: &half, Cap: uint128From(1)}, &uint128.Max)
	caseFunc(uint128From(1000), &Terms{Amount: uint128From(10), Cap: uint128From(100)}, uint128From(2000))
	caseFunc(&uint128.Max, &Terms{Amount: uint128From(1), Cap: uint128From(1)}, nil)
	caseFunc(uint128From(0), &Terms{Amount: uint128From(1), Cap: &uint128.Max}, &uint128.Max)
}
func uint128From(i int) *uint128.Uint128 {
	u := uint128.From64(uint64(i))
	return &u
}
