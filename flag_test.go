package go_runestone

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func TestFlagMask(t *testing.T) {
	m := FlagEtching.Mask()
	assert.Equal(t, uint128.From64(1), m)
	assert.Equal(t, uint128.From64(1).Lsh(1), FlagTerms.Mask())

	assert.Equal(t, uint128.From64(1).Lsh(2), FlagTurbo.Mask())
	assert.Equal(t, uint128.From64(1).Lsh(127), FlagCenotaph.Mask())
}

func TestFlagTake(t *testing.T) {
	flags := uint128.From64(1)
	assert.True(t, FlagEtching.Take(&flags))
	assert.Equal(t, 0, flags.Cmp(uint128.Zero))

	flags = uint128.From64(0)
	assert.False(t, FlagEtching.Take(&flags))
	assert.Equal(t, 0, flags.Cmp(uint128.Zero))
}

func TestFlagSet(t *testing.T) {
	flags := uint128.Zero
	FlagEtching.Set(&flags)
	assert.Equal(t, uint64(1), flags.Lo)

}
