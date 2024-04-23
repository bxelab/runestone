package go_runestone

import (
	"encoding/binary"
	"errors"
	"math/big"
	"sort"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"lukechampine.com/uint128"
)

const (
	MAGIC_NUMBER         = txscript.OP_13
	COMMIT_CONFIRMATIONS = 6
)

type Runestone struct {
	Edicts  []Edict
	Etching *Etching
	Mint    *RuneId
	Pointer *uint32
}

func (r *Runestone) Decipher(transaction *wire.MsgTx) *Artifact {
	payload, err := r.payload(transaction)
	if err == nil {
		return &Artifact{
			Cenotaph: &Cenotaph{
				Flaw: &payload.Invalid,
			},
		}
	}

	integers, err := r.integers(payload.Valid)
	if err != nil {
		flaw := Varint
		return &Artifact{
			Cenotaph: &Cenotaph{
				Flaw: &flaw,
			},
		}
	}

	message, err := MessageFromIntegers(transaction, integers)
	flags := message.takeFlags()
	etching := message.takeEtching(flags)
	mint := message.takeMint()
	pointer := message.takePointer()

	if message.hasFlaw() {
		return &Artifact{
			Cenotaph: &Cenotaph{
				Flaw:    message.Flaw,
				Mint:    mint,
				Etching: etching,
			},
		}
	}

	return &Artifact{
		Runestone: &Runestone{
			Edicts:  message.Edicts,
			Etching: etching,
			Mint:    mint,
			Pointer: pointer,
		},
	}
}

func (r *Runestone) Encipher() ([]byte, error) {

	var payload []byte
	//Etching
	if r.Etching != nil {
		flags := uint128.Zero
		FlagEtching.Set(&flags)
		if r.Etching.Terms != nil {
			FlagTerms.Set(&flags)
		}
		payload = append(payload, TagFlags.Byte())
		payload = append(payload, EncodeUint128(flags)...)
		if r.Etching.Rune != nil {
			payload = append(payload, TagRune.Byte())
			payload = append(payload, EncodeUint128(r.Etching.Rune.Value)...)
		}
		if r.Etching.Divisibility != nil {
			payload = append(payload, TagDivisibility.Byte())
			payload = append(payload, EncodeUint8(*r.Etching.Divisibility)...)
		}
		if r.Etching.Spacers != nil {
			payload = append(payload, TagSpacers.Byte())
			payload = append(payload, EncodeUint32(*r.Etching.Spacers)...)
		}
		if r.Etching.Symbol != nil {
			payload = append(payload, TagSymbol.Byte())
			payload = append(payload, runeToBytes(r.Etching.Symbol)...)
		}
		if r.Etching.Premine != nil {
			payload = append(payload, TagPremine.Byte())
			payload = append(payload, EncodeUint128(*r.Etching.Premine)...)
		}
		if r.Etching.Terms != nil {
			payload = append(payload, TagAmount.Byte())
			payload = append(payload, EncodeUint128(*r.Etching.Terms.Amount)...)
			payload = append(payload, TagCap.Byte())
			payload = append(payload, EncodeUint128(*r.Etching.Terms.Cap)...)
			if r.Etching.Terms.Height[0] != nil {
				payload = append(payload, TagHeightStart.Byte())
				payload = append(payload, EncodeUint64(*r.Etching.Terms.Height[0])...)
			}
			if r.Etching.Terms.Height[1] != nil {
				payload = append(payload, TagHeightEnd.Byte())
				payload = append(payload, EncodeUint64(*r.Etching.Terms.Height[1])...)
			}
			if r.Etching.Terms.Offset[0] != nil {
				payload = append(payload, TagOffsetStart.Byte())
				payload = append(payload, EncodeUint64(*r.Etching.Terms.Offset[0])...)
			}
			if r.Etching.Terms.Offset[1] != nil {
				payload = append(payload, TagOffsetEnd.Byte())
				payload = append(payload, EncodeUint64(*r.Etching.Terms.Offset[1])...)
			}
		}
	}
	//Mint
	if r.Mint != nil {
		payload = append(payload, TagMint.Byte())
		payload = append(payload, EncodeUint64(r.Mint.Block)...)
		payload = append(payload, TagMint.Byte())
		payload = append(payload, EncodeUint32(r.Mint.Tx)...)
	}
	//Pointer
	if r.Pointer != nil {
		payload = append(payload, TagPointer.Byte())
		payload = append(payload, EncodeUint32(*r.Pointer)...)
	}
	//Edicts
	if r.Edicts != nil {
		payload = append(payload, TagBody.Byte())
		edicts := r.Edicts
		sort.Slice(edicts, func(i, j int) bool {
			if edicts[i].ID.Block < (edicts[j].ID.Block) {
				return true
			}
			if edicts[i].ID.Block == edicts[j].ID.Block && edicts[i].ID.Block < edicts[j].ID.Block {
				return true
			}
			return false
		})

		var previous = RuneId{0, 0}
		for _, edict := range edicts {
			temp := RuneId{edict.ID.Block, edict.ID.Tx}
			block, tx, _ := previous.Delta(edict.ID)
			payload = append(payload, EncodeUint64(block)...)
			payload = append(payload, EncodeUint32(tx)...)
			payload = append(payload, EncodeUint64(edict.Amount)...)
			payload = append(payload, EncodeUint32(edict.Output)...)
			previous = temp
		}
	}

	//build op_return script
	builder := txscript.NewScriptBuilder()
	// Push OP_RETURN
	builder.AddOp(txscript.OP_RETURN)
	// Push MAGIC_NUMBER
	builder.AddInt64(int64(MAGIC_NUMBER))
	for len(payload) > 0 {
		chunkSize := txscript.MaxScriptElementSize
		if len(payload) < chunkSize {
			chunkSize = len(payload)
		}
		chunk := payload[:chunkSize]
		builder.AddData(chunk)
		payload = payload[chunkSize:]
	}
	return builder.Script()
}

type Payload struct {
	Valid   []byte
	Invalid Flaw
}

func (r *Runestone) payload(transaction *wire.MsgTx) (*Payload, error) {
	for _, output := range transaction.TxOut {
		script, err := txscript.ParsePkScript(output.PkScript)
		scriptClass := script.Class()
		if err != nil || scriptClass != txscript.NullDataTy {
			continue
		}
		instructions := output.PkScript
		// Check for OP_RETURN
		if len(instructions) < 1 || instructions[0] != txscript.OP_RETURN {
			continue
		}

		// Check for protocol identifier (Runestone::MAGIC_NUMBER)
		if len(instructions) < 2 || instructions[1] != MAGIC_NUMBER {
			continue
		}

		// Construct the payload by concatenating remaining data pushes
		var payload []byte
		if instructions[2] > txscript.OP_16 {
			return &Payload{Invalid: InvalidScript}, InvalidScript.Error()
		}
		payload = append(payload, instructions[3:]...)
		//for _, instruction := range instructions[2:] {
		//	if instruction > txscript.OP_16 {
		//		return &Payload{Invalid: "Invalid opcode"}
		//	}
		//	payload = append(payload, instruction.Data...)
		//}

		return &Payload{Valid: payload}, nil
	}

	return nil, errors.New("no OP_RETURN output found")
}

func (r *Runestone) integers(payload []byte) ([]uint128.Uint128, error) {
	var integers []uint128.Uint128
	i := 0

	for i < len(payload) {
		integer, length := binary.Uvarint(payload[i:])
		if length <= 0 {
			return nil, errors.New("invalid varint data")
		}
		integers = append(integers, uint128.From64(integer))
		i += length
	}

	return integers, nil
}
func Encode(n *big.Int) []byte {
	var result []byte
	for n.Cmp(big.NewInt(128)) > 0 {
		temp := new(big.Int).Set(n)
		last := temp.And(n, new(big.Int).SetUint64(0b0111_1111))
		result = append(result, last.Or(last, new(big.Int).SetUint64(0b1000_0000)).Bytes()...)
		n.Rsh(n, 7)
	}
	result = append(result, n.Bytes()...)
	return result
}
func EncodeUint64(n uint64) []byte {
	var result []byte
	for n >= 128 {
		result = append(result, byte(n&0x7F|0x80))
		n >>= 7
	}
	result = append(result, byte(n))
	return result
}
func EncodeUint32(n uint32) []byte {
	var result []byte
	for n >= 128 {
		result = append(result, byte(n&0x7F|0x80))
		n >>= 7
	}
	result = append(result, byte(n))
	return result
}
func EncodeUint8(n uint8) []byte {
	var result []byte
	for n >= 128 {
		result = append(result, byte(n&0x7F|0x80))
		n >>= 7
	}
	result = append(result, byte(n))
	return result
}
func EncodeUint128(n uint128.Uint128) []byte {
	return Encode(n.Big())
}

func runeToBytes(r *rune) []byte {
	return []byte(string(*r))
}
