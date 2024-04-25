package go_runestone

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeBig(t *testing.T) {
	zero := big.NewInt(0)
	result := Encode(zero)
	assert.Equal(t, 1, len(result))
}
func TestEncodeChar(t *testing.T) {
	result := EncodeChar(rune(0x10FFFF))
	t.Logf("%v,%x", result, result)
	assert.EqualValues(t, []byte{0xff, 0xff, 0x43}, result)
}
