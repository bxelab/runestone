package go_runestone

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
		Rune:    r,
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
