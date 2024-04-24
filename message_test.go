package go_runestone

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

func TestMessageFromIntegers(t *testing.T) {
	r := &Runestone{}
	payload, _ := hex.DecodeString("14f1a39f0114b2071601")
	integers, _ := r.integers(payload)

	message, err := MessageFromIntegers(&wire.MsgTx{}, integers)
	assert.NoError(t, err)
	t.Logf("message: %v", message)
	for tag, values := range message.Fields {
		t.Logf("tag: %v, values: %v", tag, values)
	}
	assert.Equal(t, 2, len(message.Fields))

}
