package go_runestone

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func runeId(tx uint32) RuneId {
	return RuneId{Block: 1, Tx: tx}
}

func decipher(integers []uint128.Uint128) (*Artifact, error) {
	var tx wire.MsgTx

	p := payload(integers)

	builder := txscript.NewScriptBuilder()

	// Push opcode OP_RETURN
	builder.AddOp(txscript.OP_RETURN)

	// Push MAGIC_NUMBER
	builder.AddOp(MAGIC_NUMBER)

	// Push payload
	builder.AddData(p)
	pkScript, _ := builder.Script()
	txOut := wire.NewTxOut(0, pkScript)
	tx.AddTxOut(txOut)
	r := &Runestone{}
	artifact, err := r.Decipher(&tx)

	return artifact, err
}

func payload(integers []uint128.Uint128) []byte {
	var payload []byte

	for _, integer := range integers {
		payload = append(payload, EncodeUint128(integer)...)
	}

	return payload
}

func TestDecipherReturnsNoneIfFirstOpcodeIsMalformed(t *testing.T) {
	var tx wire.MsgTx

	txOut := wire.NewTxOut(0, []byte{txscript.OP_PUSHDATA4})
	tx.AddTxOut(txOut)

	runestone := &Runestone{}
	artifact, err := runestone.Decipher(&tx)
	assert.Error(t, err)
	t.Logf("artifact: %v", artifact)
	t.Logf("err: %v", err)
}
func TestDecipheringTransactionWithNoOutputsReturnsNone(t *testing.T) {
	var tx wire.MsgTx

	runestone := &Runestone{}
	artifact, err := runestone.Decipher(&tx)

	if artifact != nil || err == nil {
		t.Errorf("Expected decipher to return nil and an error, but got: artifact=%v, err=%v", artifact, err)
	}
}
func Test_integers(t *testing.T) {
	r := &Runestone{}
	payload, _ := hex.DecodeString("14f1a39f0114b2071601")
	integers, err := r.integers(payload)
	assert.NoError(t, err)
	t.Logf("integers: %v", integers)
}
