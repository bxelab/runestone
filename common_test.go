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
