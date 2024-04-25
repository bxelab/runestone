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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"testing"
	"unicode/utf8"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func Uint8P(i uint8) *uint8 {
	return &i
}
func Uint32P(i uint32) *uint32 {
	return &i
}
func Uint64P(i uint64) *uint64 {
	return &i
}
func Uint128P(i uint128.Uint128) *uint128.Uint128 {
	return &i
}
func Uint128PFrom64(i uint64) *uint128.Uint128 {
	u := uint128.From64(i)
	return &u
}
func CharP(c rune) *rune {
	return &c
}

func RuneP64(i uint64) *Rune {
	return &Rune{uint128.From64(i)}
}
func assertArtifactSame(t *testing.T, expected, actual *Artifact) {
	assertJsonEqual(t, expected, actual)
}
func assertJsonEqual(t *testing.T, expected, actual interface{}) {
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
	return i.Mask()
}

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
func runeIdP(tx uint32) *RuneId {
	return &RuneId{Block: 1, Tx: tx}
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
	builder.AddOp(txscript.OP_DATA_4)
	scriptPubKey, err := builder.Script()
	assert.NoError(t, err)

	tx := wire.NewMsgTx(2)
	tx.AddTxOut(&wire.TxOut{
		PkScript: scriptPubKey,
		Value:    0,
	})
	runestone := &Runestone{}
	payload, err := runestone.payload(tx)
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
			Flaw: FlawP(UnrecognizedFlag),
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
				Flaw: FlawP(UnrecognizedEvenTag),
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
			Flaw: FlawP(Varint),
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
			Flaw:    FlawP(UnrecognizedEvenTag),
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
			Flaw: FlawP(UnrecognizedEvenTag),
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
			Flaw: FlawP(UnrecognizedFlag),
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
			Flaw: FlawP(EdictRuneId),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestRunestoneWithOverflowingEdictIdDeltaIsCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(0),
		uint128.From64(0),
		uint128.From64(0),
		uint128.Max,
		uint128.From64(0),
		uint128.From64(0),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(EdictRuneId),
		},
	}

	assertArtifactSame(t, expected, actual)

	actual, err = decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(0),
		uint128.From64(0),
		uint128.From64(0),
		uint128.Max,
		uint128.From64(0),
		uint128.From64(0),
	})

	assert.NoError(t, err)
	assertArtifactSame(t, expected, actual)
}

func TestRunestoneWithOutputOverMaxIsCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(2),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(EdictOutput),
		},
	}

	assertArtifactSame(t, expected, actual)
}

func TestTagWithNoValueIsCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		uint128.From64(1),
		itag(TagFlags),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(TruncatedField),
		},
	}

	assertArtifactSame(t, expected, actual)
}

func TestTrailingIntegersInBodyIsCenotaph(t *testing.T) {
	integers := []uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	}

	for i := 0; i < 4; i++ {
		actual, err := decipher(integers)
		assert.NoError(t, err)

		var expected *Artifact
		if i == 0 {
			expected = &Artifact{
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
		} else {
			expected = &Artifact{
				Cenotaph: &Cenotaph{
					Flaw: FlawP(TrailingIntegers),
				},
			}
		}

		assertArtifactSame(t, expected, actual)

		integers = append(integers, uint128.From64(0))
	}
}
func TestDecipherEtchingWithDivisibility(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
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
	u5 := uint8(5)
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
			Etching: &Etching{
				Rune:         &r,
				Divisibility: &u5,
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestDivisibilityAboveMaxIsIgnored(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		uint128.From64(4),
		itag(TagDivisibility),
		uint128.From64(MaxDivisibility + 1),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
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
			Etching: &Etching{
				Rune: &r,
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestSymbolAboveMaxIsIgnored(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagSymbol),
		uint128.From64(uint64(utf8.MaxRune) + 1),
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
			Etching: &Etching{},
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestDecipherEtchingWithSymbol(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		uint128.From64(4),
		itag(TagSymbol),
		uint128.From64('a'),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	r := NewRune(uint128.From64(4))
	a := 'a'
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
				Rune:   &r,
				Symbol: &a,
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestDecipherEtchingWithAllEtchingTags(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()).Or(FlagTurbo.Mask()),
		itag(TagRune),
		uint128.From64(4),
		itag(TagDivisibility),
		uint128.From64(1),
		itag(TagSpacers),
		uint128.From64(5),
		itag(TagSymbol),
		uint128.From64('a'),
		itag(TagOffsetEnd),
		uint128.From64(2),
		itag(TagAmount),
		uint128.From64(3),
		itag(TagPremine),
		uint128.From64(8),
		itag(TagCap),
		uint128.From64(9),
		itag(TagPointer),
		uint128.From64(0),
		itag(TagMint),
		uint128.From64(1),
		itag(TagMint),
		uint128.From64(1),
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
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
			Etching: &Etching{
				Divisibility: Uint8P(1),
				Premine:      Uint128PFrom64(8),
				Rune:         &r,
				Spacers:      Uint32P(5),
				Symbol:       CharP('a'),
				Terms: &Terms{
					Cap:    Uint128PFrom64(9),
					Offset: [2]*uint64{nil, Uint64P(2)},
					Amount: Uint128PFrom64(3),
				},
				Turbo: true,
			},
			Pointer: Uint32P(0),
			Mint:    runeIdP(1),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestRecognizedEvenEtchingFieldsProduceCenotaphIfEtchingFlagIsNotSet(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagRune),
		uint128.From64(4),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestDecipherEtchingWithDivisibilityAndSymbol(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		uint128.From64(4),
		itag(TagDivisibility),
		uint128.From64(1),
		itag(TagSymbol),
		uint128.From64('a'),
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
			Etching: &Etching{
				Rune:         RuneP64(4),
				Divisibility: Uint8P(1),
				Symbol:       CharP('a'),
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestTagValuesAreNotParsedAsTags(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagDivisibility),
		itag(TagBody),
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
			Etching: &Etching{
				Divisibility: Uint8P(0),
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestRunestoneMayContainMultipleEdicts(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
		uint128.From64(0),
		uint128.From64(3),
		uint128.From64(5),
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
				{
					ID:     runeId(4),
					Amount: uint128.From64(5),
					Output: 0,
				},
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestRunestonesWithInvalidRuneIDBlocksAreCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
		uint128.Max,
		uint128.From64(1),
		uint128.From64(0),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(EdictRuneId),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestRunestonesWithInvalidRuneIdTxsAreCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(2),
		uint128.From64(0),
		uint128.From64(1),
		uint128.Max,
		uint128.From64(0),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(EdictRuneId),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestPayloadPushesAreConcatenated(t *testing.T) {

	s := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).
		AddOp(MAGIC_NUMBER)
	data := []byte{}
	data = append(data, EncodeUint64(uint64(TagFlags))...)
	data = append(data, EncodeUint128(FlagEtching.Mask())...)
	data = append(data, EncodeUint64(uint64(TagDivisibility))...)
	data = append(data, EncodeUint64(5)...)
	data = append(data, EncodeUint64(uint64(TagBody))...)
	data = append(data, EncodeUint64(1)...)
	data = append(data, EncodeUint64(1)...)
	data = append(data, EncodeUint64(2)...)
	data = append(data, EncodeUint64(0)...)
	pkScript, _ := s.AddData(data).Script()

	tx := &wire.MsgTx{
		Version:  2,
		TxIn:     []*wire.TxIn{},
		LockTime: 0,
		TxOut: []*wire.TxOut{{
			Value:    0,
			PkScript: pkScript,
		}},
	}
	r := &Runestone{}
	artifact, err := r.Decipher(tx)
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{
			Edicts: []Edict{{
				ID:     runeId(1),
				Amount: uint128.From64(2),
				Output: 0,
			}},
			Etching: &Etching{
				Divisibility: Uint8P(5),
			},
		},
	}
	assertArtifactSame(t, expected, artifact)
}

func TestRunestoneMayBeInSecondOutput(t *testing.T) {
	payload := []byte{0, 1, 1, 2, 0}
	pkScript, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).AddOp(MAGIC_NUMBER).AddData(payload).Script()
	tx := wire.NewMsgTx(2)
	tx.AddTxOut(wire.NewTxOut(0, pkScript))
	r := &Runestone{}
	actual, err := r.Decipher(tx)
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
func TestRunestoneMayBeAfterNonMatchingOpReturn(t *testing.T) {
	payload := []byte{0, 1, 1, 2, 0}
	foo, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).AddData([]byte("FOO")).Script()
	pkScript, _ := txscript.NewScriptBuilder().
		AddOp(txscript.OP_RETURN).AddOp(MAGIC_NUMBER).AddData(payload).Script()

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

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxOut(wire.NewTxOut(0, foo))
	tx.AddTxOut(wire.NewTxOut(0, pkScript))
	r := &Runestone{}
	actual, err := r.Decipher(tx)
	assert.NoError(t, err)
	assertArtifactSame(t, expected, actual)
}
func TestRunestoneSize(t *testing.T) {
	type testCase struct {
		edicts   []Edict
		etching  *Etching
		expected int
	}

	cases := []testCase{
		{
			edicts:   []Edict{},
			etching:  nil,
			expected: 2,
		},
		{
			edicts: []Edict{},
			etching: &Etching{
				Rune: RuneP64(0),
			},
			expected: 7,
		},
		{
			edicts: []Edict{},
			etching: &Etching{
				Divisibility: Uint8P(MaxDivisibility),
				Rune:         RuneP64(0),
			},
			expected: 9,
		},
		{
			edicts: []Edict{},
			etching: &Etching{
				Divisibility: Uint8P(MaxDivisibility),
				Terms: &Terms{
					Cap:    Uint128PFrom64(math.MaxUint32),
					Amount: Uint128PFrom64(math.MaxUint64),
					Offset: [2]*uint64{Uint64P(math.MaxUint32), Uint64P(math.MaxUint32)},
					Height: [2]*uint64{Uint64P(math.MaxUint32), Uint64P(math.MaxUint32)},
				},
				Turbo:   true,
				Premine: Uint128PFrom64(math.MaxUint64),
				Rune:    &Rune{Value: uint128.Max},
				Symbol:  CharP(rune(0x10FFFF)),
				Spacers: Uint32P(MaxSpacers),
			},
			expected: 89,
		},
		{
			edicts: []Edict{},
			etching: &Etching{
				Rune: &Rune{Value: uint128.Max},
			},
			expected: 25,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.Zero,
					ID:     RuneId{Block: 0, Tx: 0},
					Output: 0,
				},
			},
			etching: &Etching{
				Divisibility: Uint8P(MaxDivisibility),
				Rune:         &Rune{Value: uint128.Max},
			},
			expected: 32,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 0, Tx: 0},
					Output: 0,
				},
			},
			etching: &Etching{
				Divisibility: Uint8P(MaxDivisibility),
				Rune:         &Rune{Value: uint128.Max},
			},
			expected: 50,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.Zero,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 14,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 32,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 54,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.Max,
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 76,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 62,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 75,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 0, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 0, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 0, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 0, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128.From64(math.MaxUint64),
					ID:     RuneId{Block: 0, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 73,
		},
		{
			edicts: []Edict{
				{
					Amount: uint128FromString("1000000000000000000"),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128FromString("1000000000000000000"),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128FromString("1000000000000000000"),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128FromString("1000000000000000000"),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
				{
					Amount: uint128FromString("1000000000000000000"),
					ID:     RuneId{Block: 1000000, Tx: math.MaxUint32},
					Output: 0,
				},
			},
			etching:  nil,
			expected: 70,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case[%d]", i), func(t *testing.T) {
			runestone := &Runestone{
				Edicts:  c.edicts,
				Etching: c.etching,
			}
			code, err := runestone.Encipher()
			assert.NoError(t, err)
			t.Logf("code: %x", code)
			actual := len(code)
			assert.Equal(t, c.expected, actual)
		})
	}
}
func uint128FromString(str string) uint128.Uint128 {
	v, _ := uint128.FromString(str)
	return v
}
func TestEtchingWithTermGreaterThanMaximumIsStillAnEtching(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagOffsetEnd),
		uint128.From64(math.MaxUint64).Add(uint128.From64(1)),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestEncipher(t *testing.T) {
	caseFunc := func(runestone Runestone, expected []uint128.Uint128) {
		scriptPubkey, _ := runestone.Encipher()

		transaction := &wire.MsgTx{
			TxIn: []*wire.TxIn{},
			TxOut: []*wire.TxOut{
				{
					PkScript: scriptPubkey,
					Value:    0,
				},
			},
			LockTime: 0,
			Version:  2,
		}

		payload, err := runestone.payload(transaction)
		if err != nil {
			t.Fatalf("invalid payload: %v", err)
		}

		actual, err := runestone.integers(payload.Valid)
		if err != nil {
			t.Fatalf("failed to get integers: %v", err)
		}

		assertJsonEqual(t, expected, actual)

		sort.Slice(runestone.Edicts, func(i, j int) bool {
			return runestone.Edicts[i].ID.Cmp(runestone.Edicts[j].ID) < 0
		})

		artifact, err := runestone.Decipher(transaction)
		if err != nil {
			t.Fatalf("failed to decipher: %v", err)
		}

		assertArtifactSame(t, &Artifact{Runestone: &runestone}, artifact)
	}
	//    case(Runestone::default(), &[]);
	caseFunc(Runestone{}, []uint128.Uint128{})

	runeId1, _ := NewRuneId(2, 3)
	runeId2, _ := NewRuneId(5, 6)
	runeId3, _ := NewRuneId(17, 18)
	caseFunc(
		Runestone{
			Edicts: []Edict{
				{
					ID:     *runeId1,
					Amount: uint128.From64(1),
					Output: 0,
				},
				{
					ID:     *runeId2,
					Amount: uint128.From64(4),
					Output: 1,
				},
			},
			Etching: &Etching{
				Divisibility: Uint8P(7),
				Premine:      Uint128PFrom64(8),
				Rune:         RuneP64(9),
				Spacers:      Uint32P(10),
				Symbol:       CharP('@'),
				Terms: &Terms{
					Cap:    Uint128PFrom64(11),
					Height: [2]*uint64{Uint64P(12), Uint64P(13)},
					Amount: Uint128PFrom64(14),
					Offset: [2]*uint64{Uint64P(15), Uint64P(16)},
				},
				Turbo: true,
			},
			Mint:    runeId3,
			Pointer: Uint32P(0),
		},
		[]uint128.Uint128{
			itag(TagFlags),
			FlagEtching.Mask().Or(FlagTerms.Mask()).Or(FlagTurbo.Mask()),
			itag(TagRune),
			uint128.From64(9),
			itag(TagDivisibility),
			uint128.From64(7),
			itag(TagSpacers),
			uint128.From64(10),
			itag(TagSymbol),
			uint128.From64('@'),
			itag(TagPremine),
			uint128.From64(8),
			itag(TagAmount),
			uint128.From64(14),
			itag(TagCap),
			uint128.From64(11),
			itag(TagHeightStart),
			uint128.From64(12),
			itag(TagHeightEnd),
			uint128.From64(13),
			itag(TagOffsetStart),
			uint128.From64(15),
			itag(TagOffsetEnd),
			uint128.From64(16),
			itag(TagMint),
			uint128.From64(17),
			itag(TagMint),
			uint128.From64(18),
			itag(TagPointer),
			uint128.From64(0),
			itag(TagBody),
			uint128.From64(2),
			uint128.From64(3),
			uint128.From64(1),
			uint128.From64(0),
			uint128.From64(3),
			uint128.From64(6),
			uint128.From64(4),
			uint128.From64(1),
		},
	)

	caseFunc(
		Runestone{
			Etching: &Etching{
				Rune: RuneP64(3),
			},
		},
		[]uint128.Uint128{
			itag(TagFlags),
			FlagEtching.Mask(),
			itag(TagRune),
			uint128.From64(3),
		},
	)

	caseFunc(
		Runestone{
			Etching: &Etching{},
		},
		[]uint128.Uint128{
			itag(TagFlags),
			FlagEtching.Mask(),
		},
	)
}
func scriptInstructionsCount(script []byte) int {
	tokenizer := txscript.MakeScriptTokenizer(0, script)
	count := 0
	for tokenizer.Next() {
		count++
	}
	return count
}

func TestRunestonePayloadIsChunked(t *testing.T) {
	script := &Runestone{
		Edicts: make([]Edict, 129),
	}
	data, err := script.Encipher()
	assert.NoError(t, err)
	assert.Equal(t, 3, scriptInstructionsCount(data))

	script = &Runestone{
		Edicts: make([]Edict, 130),
	}
	data, err = script.Encipher()
	assert.NoError(t, err)

	assert.Equal(t, 4, scriptInstructionsCount(data))
}
func TestEdictOutputGreaterThan32MaxProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagBody),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(1),
		uint128.From64(math.MaxUint32).Add(uint128.From64(1)),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(EdictOutput),
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestPartialMintProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagMint),
		uint128.From64(1),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestInvalidMintProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagMint),
		uint128.Zero,
		itag(TagMint),
		uint128.From64(1),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestInvalidDeadlineProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagOffsetEnd),
		uint128.Max,
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestInvalidDefaultOutputProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagPointer),
		uint128.From64(1),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)

	actual, err = decipher([]uint128.Uint128{
		itag(TagPointer),
		uint128.Max,
	})
	assert.NoError(t, err)
	assertArtifactSame(t, expected, actual)
}

func TestInvalidDivisibilityDoesNotProduceCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagDivisibility),
		uint128.Max,
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{},
	}
	assertArtifactSame(t, expected, actual)
}
func TestMinAndMaxRunesAreNotCenotaphs(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		uint128.From64(0),
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{
			Etching: &Etching{
				Rune: RuneP64(0),
			},
		},
	}
	assertArtifactSame(t, expected, actual)

	actual, err = decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask(),
		itag(TagRune),
		uint128.Max,
	})
	assert.NoError(t, err)

	expected = &Artifact{
		Runestone: &Runestone{
			Etching: &Etching{
				Rune: &Rune{Value: uint128.Max},
			},
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestInvalidSpacersDoesNotProduceCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagSpacers),
		uint128.Max,
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{},
	}
	assertArtifactSame(t, expected, actual)
}

func TestInvalidSymbolDoesNotProduceCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagSymbol),
		uint128.Max,
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Runestone: &Runestone{},
	}
	assertArtifactSame(t, expected, actual)
}

func TestInvalidTermProducesCenotaph(t *testing.T) {
	actual, err := decipher([]uint128.Uint128{
		itag(TagOffsetEnd),
		uint128.Max,
	})
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(UnrecognizedEvenTag),
		},
	}
	assertArtifactSame(t, expected, actual)
}
func TestInvalidSupplyProducesCenotaph(t *testing.T) {
	actual1, err1 := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()),
		itag(TagCap),
		uint128.From64(1),
		itag(TagAmount),
		uint128.Max,
	})
	assert.NoError(t, err1)

	expected1 := &Artifact{
		Runestone: &Runestone{
			Etching: &Etching{
				Terms: &Terms{
					Cap:    Uint128PFrom64(1),
					Amount: Uint128P(uint128.Max),
					Height: [2]*uint64{nil, nil},
					Offset: [2]*uint64{nil, nil},
				},
			},
		},
	}
	assertArtifactSame(t, expected1, actual1)

	actual2, err2 := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()),
		itag(TagCap),
		uint128.From64(2),
		itag(TagAmount),
		uint128.Max,
	})
	assert.NoError(t, err2)

	expected2 := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(SupplyOverflow),
		},
	}
	assertArtifactSame(t, expected2, actual2)

	actual3, err3 := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()),
		itag(TagCap),
		uint128.From64(2),
		itag(TagAmount),
		uint128.Max.Div(uint128.From64(2)).Add(uint128.From64(1)),
	})
	assert.NoError(t, err3)

	expected3 := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(SupplyOverflow),
		},
	}
	assertArtifactSame(t, expected3, actual3)

	actual4, err4 := decipher([]uint128.Uint128{
		itag(TagFlags),
		FlagEtching.Mask().Or(FlagTerms.Mask()),
		itag(TagPremine),
		uint128.From64(1),
		itag(TagCap),
		uint128.From64(1),
		itag(TagAmount),
		uint128.Max,
	})
	assert.NoError(t, err4)

	expected4 := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(SupplyOverflow),
		},
	}
	assertArtifactSame(t, expected4, actual4)
}
func TestInvalidScriptsInOpReturnsWithoutMagicNumberAreIgnored(t *testing.T) {
	tx1 := &wire.MsgTx{
		Version: 2,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{},
				SignatureScript:  []byte{},
				Sequence:         wire.MaxTxInSequenceNum,
			},
		},
		TxOut: []*wire.TxOut{
			{
				PkScript: []byte{
					txscript.OP_RETURN,
					txscript.OP_DATA_4,
				},
				Value: 0,
			},
		},
	}
	r := &Runestone{}
	actual1, err := r.Decipher(tx1)

	assert.Nil(t, actual1)

	pkScript, _ := r.Encipher()
	tx2 := &wire.MsgTx{
		Version: 2,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{},
				SignatureScript:  []byte{},
				Sequence:         wire.MaxTxInSequenceNum,
			},
		},
		TxOut: []*wire.TxOut{
			{
				PkScript: []byte{
					txscript.OP_RETURN,
					txscript.OP_DATA_4,
				},
				Value: 0,
			},
			{
				PkScript: pkScript,
				Value:    0,
			},
		},
	}
	actual2, err := r.Decipher(tx2)
	assert.NoError(t, err)

	expected2 := &Artifact{
		Runestone: r,
	}
	assertArtifactSame(t, expected2, actual2)
}
func TestInvalidScriptsInOpReturnsWithMagicNumberProduceCenotaph(t *testing.T) {
	tx := &wire.MsgTx{
		Version: 2,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: wire.OutPoint{},
				SignatureScript:  []byte{},
				Sequence:         wire.MaxTxInSequenceNum,
			},
		},
		TxOut: []*wire.TxOut{
			{
				PkScript: []byte{
					txscript.OP_RETURN,
					byte(MAGIC_NUMBER),
					txscript.OP_DATA_4,
				},
				Value: 0,
			},
		},
	}
	r := &Runestone{}
	actual, err := r.Decipher(tx)
	assert.NoError(t, err)

	expected := &Artifact{
		Cenotaph: &Cenotaph{
			Flaw: FlawP(InvalidScript),
		},
	}
	assertArtifactSame(t, expected, actual)
}

func TestAllPushdataOpcodesAreValid(t *testing.T) {
	for i := 0; i < 79; i++ {
		t.Run(fmt.Sprintf("case[%d]", i), func(t *testing.T) {
			var scriptPubkey []byte
			scriptPubkey = append(scriptPubkey, uint8(txscript.OP_RETURN))
			scriptPubkey = append(scriptPubkey, uint8(MAGIC_NUMBER))
			scriptPubkey = append(scriptPubkey, byte(i))

			switch {
			case i >= 0 && i <= 75:
				for j := 0; j < i; j++ {
					if j%2 == 0 {
						scriptPubkey = append(scriptPubkey, 1)
					} else {
						scriptPubkey = append(scriptPubkey, 0)
					}
				}

				if i%2 == 1 {
					scriptPubkey = append(scriptPubkey, 1)
					scriptPubkey = append(scriptPubkey, 1)
				}
			case i == 76:
				scriptPubkey = append(scriptPubkey, 0)
			case i == 77:
				scriptPubkey = append(scriptPubkey, 0)
				scriptPubkey = append(scriptPubkey, 0)
			case i == 78:
				scriptPubkey = append(scriptPubkey, 0)
				scriptPubkey = append(scriptPubkey, 0)
				scriptPubkey = append(scriptPubkey, 0)
				scriptPubkey = append(scriptPubkey, 0)
			default:
				assert.Fail(t, "unreachable")
			}

			tx := &wire.MsgTx{
				Version: 2,

				TxOut: []*wire.TxOut{
					{
						PkScript: scriptPubkey,
						Value:    0,
					},
				},
			}
			r := &Runestone{}
			artifact, err := r.Decipher(tx)
			assert.NoError(t, err)
			assertArtifactSame(t, &Artifact{Runestone: &Runestone{}}, artifact)
		})
	}
}
func TestAllNonPushdataOpcodesAreInvalid(t *testing.T) {
	for i := 79; i <= math.MaxUint8; i++ {
		pkScript, _ := txscript.NewScriptBuilder().AddOp(txscript.OP_RETURN).AddOp(byte(MAGIC_NUMBER)).AddOp(byte(i)).Script()

		tx := &wire.MsgTx{
			Version: 2,

			TxOut: []*wire.TxOut{
				{
					PkScript: pkScript,
					Value:    0,
				},
			},
		}
		r := &Runestone{}
		artifact, err := r.Decipher(tx)
		assert.NoError(t, err)

		expected := &Artifact{
			Cenotaph: &Cenotaph{
				Flaw: FlawP(Opcode),
			},
		}
		assertArtifactSame(t, expected, artifact)
	}
}
