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
	"math/bits"
	"strings"
	"unicode"
)

type SpacedRune struct {
	Rune    Rune
	Spacers uint32
}

func NewSpacedRune(r Rune, spacers uint32) *SpacedRune {
	return &SpacedRune{Rune: r, Spacers: spacers}
}

func (sr *SpacedRune) String() string {
	var b strings.Builder
	runeStr := sr.Rune.String()

	for i, c := range runeStr {
		b.WriteRune(c)
		if i < len(runeStr)-1 && sr.Spacers&(1<<i) != 0 {
			b.WriteRune('•')
		}
	}

	return b.String()
}

func SpacedRuneFromString(s string) (*SpacedRune, error) {
	var runeStr string
	var spacers uint32

	for _, c := range s {
		switch {
		case unicode.IsUpper(c):
			runeStr += string(c)
		case c == '.' || c == '•':

			//let flag = 1 << rune.len().checked_sub(1).ok_or(Error::LeadingSpacer)?;
			if len(runeStr) == 0 {
				return nil, ErrLeadingSpacer
			}
			flag := uint32(1) << (len(runeStr) - 1)
			if spacers&flag != 0 {
				return nil, ErrDoubleSpacer
			}
			spacers |= flag
		default:
			return nil, ErrCharacter(c)
		}
	}

	if 32-bits.LeadingZeros32(spacers) >= len(runeStr) {
		return nil, ErrTrailingSpacer
	}

	r, err := RuneFromString(runeStr)
	if err != nil {
		return nil, fmt.Errorf("rune error: %v", err)
	}

	return &SpacedRune{
		Rune:    *r,
		Spacers: spacers,
	}, nil
}

var (
	ErrCharacter = func(c rune) error {
		return fmt.Errorf("invalid character `%c`", c)
	}
	ErrDoubleSpacer   = errors.New("double spacer")
	ErrLeadingSpacer  = errors.New("leading spacer")
	ErrTrailingSpacer = errors.New("trailing spacer")
)
