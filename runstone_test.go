package go_runestone

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func Test_integers(t *testing.T) {
	r := &Runestone{}
	payload, _ := hex.DecodeString("14f1a39f0114b2071601")
	integers, err := r.integers(payload)
	assert.NoError(t, err)
	t.Logf("integers: %v", integers)
}
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
	artifact, _ := runestone.Decipher(&tx)
	t.Logf("artifact: %v", artifact)
	//TODO:
}
func TestDecipheringTransactionWithNoOutputsReturnsNone(t *testing.T) {
	var tx wire.MsgTx

	runestone := &Runestone{}
	artifact, err := runestone.Decipher(&tx)

	if artifact != nil || err == nil {
		t.Errorf("Expected decipher to return nil and an error, but got: artifact=%v, err=%v", artifact, err)
	}
}

func TestDecipheringTransactionWithNonOpReturnOutputReturnsNone(t *testing.T) {
	tx := wire.NewMsgTx(wire.TxVersion)
	txOut := wire.NewTxOut(0, []byte{})
	tx.AddTxOut(txOut)
	runestone := &Runestone{}
	a, _ := runestone.Decipher(tx)
	assert.Nil(t, a)
}

func TestDecipheringTransactionWithBareOpReturnReturnsNone(t *testing.T) {
	tx := wire.NewMsgTx(wire.TxVersion)
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	script, _ := builder.Script()
	txOut := wire.NewTxOut(0, script)
	tx.AddTxOut(txOut)
	runestone := &Runestone{}
	a, _ := runestone.Decipher(tx)
	assert.Nil(t, a)
}
func TestDecipheringTransactionWithNonMatchingOpReturnReturnsNone(t *testing.T) {
	tx := wire.NewMsgTx(wire.TxVersion)
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	builder.AddData([]byte("FOOO"))
	script, _ := builder.Script()
	txOut := wire.NewTxOut(0, script)
	tx.AddTxOut(txOut)
	runestone := &Runestone{}
	a, _ := runestone.Decipher(tx)
	assert.Nil(t, a)
}

func TestDecipheringValidRunestoneWithInvalidScriptPostfixReturnsInvalidPayload(t *testing.T) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	builder.AddOp(MAGIC_NUMBER)

	scriptPubKey, err := builder.Script()
	assert.NoError(t, err)

	scriptPubKey = append(scriptPubKey, txscript.OP_PUSHDATA4)

	tx := wire.NewMsgTx(2)
	tx.AddTxOut(&wire.TxOut{
		PkScript: scriptPubKey,
		Value:    0,
	})
	runestone := &Runestone{}
	payload, err := runestone.payload(tx)
	assert.Error(t, err)
	assert.NotNil(t, payload)
	assert.Equal(t, InvalidScript, payload.Invalid)
}

func TestDecipheringRunestoneWithTruncatedVarintSucceeds(t *testing.T) {
	tx := wire.NewMsgTx(2)

	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN).AddOp(MAGIC_NUMBER).AddData([]byte{128})
	script, err := builder.Script()
	assert.NoError(t, err)

	txOut := &wire.TxOut{
		PkScript: script,
		Value:    0,
	}
	tx.AddTxOut(txOut)
	runestone := &Runestone{}
	a, err := runestone.Decipher(tx)
	assert.NotNil(t, a)
}

func TestOutputsNonPushDataOpcodesAreCenotaph(t *testing.T) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN).AddOp((MAGIC_NUMBER)).AddOp(txscript.OP_VERIFY).
		AddData([]byte{0}).AddData(EncodeUint64(uint64(1))).
		AddData(EncodeUint64(uint64(1))).AddData([]byte{2, 0})
	script1, err := builder.Script()
	assert.NoError(t, err)

	builder = txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN).AddOp((MAGIC_NUMBER)).AddData([]byte{0}).
		AddData(EncodeUint64(uint64(1))).AddData(EncodeUint64(uint64(2))).AddData([]byte{3, 0})
	script2, err := builder.Script()
	assert.NoError(t, err)

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxOut(wire.NewTxOut(0, script1))
	tx.AddTxOut(wire.NewTxOut(0, script2))
	runestone := &Runestone{}
	a, err := runestone.Decipher(tx)
	t.Logf("err:%v", err)
	assert.NotNil(t, a)
	t.Log(a.Cenotaph.Flaw)
	assert.Equal(t, Opcode.String(), a.Cenotaph.Flaw.String())
}
func TestPushNumOpcodesInRunestoneProduceCenotaph(t *testing.T) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN).AddOp((MAGIC_NUMBER)).AddOp(txscript.OP_1)
	script1, err := builder.Script()
	assert.NoError(t, err)
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxOut(wire.NewTxOut(0, script1))
	runestone := &Runestone{}
	a, err := runestone.Decipher(tx)
	t.Logf("err:%v", err)
	assert.NotNil(t, a)
	t.Log(a.Cenotaph.Flaw)
	assert.Equal(t, Opcode.String(), a.Cenotaph.Flaw.String())
}

func TestDecipheringEmptyRunestoneIsSuccessful(t *testing.T) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	builder.AddOp(MAGIC_NUMBER)
	scriptPubKey, _ := builder.Script()

	tx := &wire.MsgTx{
		Version: 2,
		TxOut: []*wire.TxOut{
			{
				PkScript: scriptPubKey,
				Value:    0,
			},
		},
		LockTime: 0,
	}

	runestone := &Runestone{}
	artifact, _ := runestone.Decipher(tx)
	assert.NotNil(t, artifact)
}
func TestDecipherEtching(t *testing.T) {

	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagBody),
		i128(1),
		i128(1),
		i128(2),
		i128(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{
				{
					ID:     runeId(1),
					Amount: uint128.From64(2),
					Output: 0,
				},
			},
			Etching: &Etching{},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func assertArtifactSame(t *testing.T, expected, actual *Artifact) {
	expect, err := json.Marshal(expected)
	assert.NoError(t, err)
	act, err := json.Marshal(actual)
	assert.NoError(t, err)
	assert.Equal(t, string(expect), string(act))
}

func i128(i int) uint128.Uint128 {
	return uint128.From64(uint64(i))
}
func itag(i Tag) uint128.Uint128 {
	return uint128.From64(uint64(i))
}
func iflag(i Flag) uint128.Uint128 {
	return uint128.From64(uint64(i))
}
func TestDecipherEtchingWithRune(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		i128(4),
		itag(TagBody),
		i128(1),
		i128(1),
		i128(2),
		i128(0),
	})
	assert.NoError(t, err)
	r := NewRune(uint128.From64(4))
	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{
				{
					ID:     runeId(1),
					Amount: uint128.From64(2),
					Output: 0,
				},
			},
			Etching: &Etching{Rune: &r},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestTermsFlagWithoutEtchingFlagProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagTerms.Mask(),
		itag(TagBody),
		i128(0),
		i128(0),
		i128(0),
		i128(0),
	})
	assert.NoError(t, err)
	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: NewFlawP(UnrecognizedFlag),
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestRecognizedFieldsWithoutFlagProducesCenotaph(t *testing.T) {
	caseFunc := func(integers []uint128.Uint128) {
		result, err := decipher(integers)
		assert.NoError(t, err)
		expected := Artifact{
			Cenotaph: &Cenotaph{
				Flaw: NewFlawP(UnrecognizedEvenTag),
			},
		}
		assertArtifactSame(t, &expected, result)
	}

	caseFunc([]uint128.Uint128{itag(TagPremine), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagRune), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagCap), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagAmount), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagOffsetStart), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagOffsetEnd), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagHeightStart), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagHeightEnd), uint128.Zero})

	caseFunc([]uint128.Uint128{itag(TagFlags), iflag(FlagEtching), itag(TagCap), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagFlags), iflag(FlagEtching), itag(TagAmount), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagFlags), iflag(FlagEtching), itag(TagOffsetStart), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagFlags), iflag(FlagEtching), itag(TagOffsetEnd), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagFlags), iflag(FlagEtching), itag(TagHeightStart), uint128.Zero})
	caseFunc([]uint128.Uint128{itag(TagFlags), iflag(FlagEtching), itag(TagHeightEnd), uint128.Zero})
}
func TestDecipherEtchingWithTerm(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()),
		itag(TagOffsetEnd),
		uint128.From64(4),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)
	u4 := uint64(4)
	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{
				{
					ID:     runeId(1),
					Amount: uint128.From64(2),
					Output: 0,
				},
			},
			Etching: &Etching{
				Terms: &Terms{
					Offset: [2]*uint64{nil, &u4},
				},
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestDecipherEtchingWithAmount(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()),
		itag(TagAmount),
		uint128.From64(4),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)
	u4 := uint128.From64(4)
	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{
				{
					ID:     runeId(1),
					Amount: uint128.From64(2),
					Output: 0,
				},
			},
			Etching: &Etching{
				Terms: &Terms{
					Amount: &u4,
				},
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestInvalidVarintProducesCenotaph(t *testing.T) {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_RETURN)
	builder.AddOp(MAGIC_NUMBER)
	builder.AddData([]byte{128})

	script, err := builder.Script()
	assert.NoError(t, err)

	tx := wire.NewMsgTx(2)
	tx.AddTxOut(&wire.TxOut{
		PkScript: script,
		Value:    0,
	})
	r := &Runestone{}
	actual, err := r.Decipher(tx)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: NewFlawP(Varint),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestDuplicateEvenTagsProduceCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		uint128.From64(4),
		itag(TagRune),
		uint128.From64(5),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw:    NewFlawP(UnrecognizedEvenTag),
			Etching: &Rune{uint128.From64(4)},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestDuplicateOddTagsAreIgnored(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagDivisibility),
		uint128.From64(4),
		itag(TagDivisibility),
		uint128.From64(5),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)
	u4 := uint8(4)
	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{
				{
					ID:     runeId(1),
					Amount: uint128.From64(2),
					Output: 0,
				},
			},
			Etching: &Etching{
				Rune:         nil,
				Divisibility: &u4,
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestUnrecognizedOddTagIsIgnored(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagNop),
		uint128.From64(100),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{
				{
					ID:     runeId(1),
					Amount: uint128.From64(2),
					Output: 0,
				},
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestRunestoneWithUnrecognizedEvenTagIsCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagCenotaph),
		uint128.From64(0),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: NewFlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestRunestoneWithUnrecognizedFlagIsCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagCenotaph.Mask(),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: NewFlawP(UnrecognizedFlag),
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestRunestoneWithEdictIdWithZeroBlockAndNonzeroTxIsCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(0),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: NewFlawP(EdictRuneId),
		},
	}
	assertArtifactSame(t, expected, actual)
}
