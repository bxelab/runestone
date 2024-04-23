package go_runestone

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_integers(t *testing.T) {
	payload, _ := hex.DecodeString("14f1a39f0114b2071601")
	integers, err := integers(payload)
	assert.NoError(t, err)
	t.Logf("integers: %v", integers)
}
