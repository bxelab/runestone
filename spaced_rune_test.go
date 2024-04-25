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
)

func TestDisplay(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"A.B", "A•B"},
		{"A.B.C", "A•B•C"},
		// TODO: Add a test case for SpacedRune with rune: Rune(0), spacers: 1
	}

	for _, tc := range testCases {
		sr, err := SpacedRuneFromString(tc.input)
		if err != nil {
			t.Fatalf("ParseSpacedRune(%q) returned error: %v", tc.input, err)
		}

		got := sr.String()
		if got != tc.expected {
			t.Errorf("SpacedRune.String() = %q; want %q", got, tc.expected)
		}
	}
}
func TestFromString(t *testing.T) {
	caseFn := func(s, runeStr string, spacers uint32) {
		sr, err := SpacedRuneFromString(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		r, err := RuneFromString(runeStr)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := NewSpacedRune(*r, spacers)
		if *sr != *expected {
			t.Fatalf("expected: %v, got: %v", expected, sr)
		}
	}

	_, err := SpacedRuneFromString(".A")
	assert.Equal(t, ErrLeadingSpacer, err)

	_, err = SpacedRuneFromString("A..B")
	assert.Equal(t, ErrDoubleSpacer, err)

	_, err = SpacedRuneFromString("A.")
	assert.Equal(t, ErrTrailingSpacer, err)

	_, err = SpacedRuneFromString("Ax")
	if err == nil || err.Error() != "invalid character `x`" {
		t.Fatalf("expected error: invalid character `x`, got: %v", err)
	}

	caseFn("A.B", "AB", 0b1)
	caseFn("A.B.C", "ABC", 0b11)
	caseFn("A•B", "AB", 0b1)
	caseFn("A•B•C", "ABC", 0b11)
	caseFn("A•BC", "ABC", 0b1)
}

//func TestSerDe(t *testing.T) {
//	spacedRune := NewSpacedRune(NewRune(uint128.From64(26)), 1)
//	json := "\"A•A\""
//
//	// Note: Go doesn't have built-in serialization for custom types
//	// like serde_json in Rust, so we'll test the String method instead.
//	str := spacedRune.String()
//	if str != json {
//		t.Fatalf("expected: %s, got: %s", json, str)
//	}
//
//	// Deserialize test is covered by TestFromString
//}
