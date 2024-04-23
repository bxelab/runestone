package go_runestone

import (
	"encoding/binary"
	"errors"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"lukechampine.com/uint128"
)

const MAGIC_NUMBER = txscript.OP_13

type Runestone struct {
	Edicts  []Edict
	Etching *Etching
	Mint    *RuneId
	Pointer *uint32
}

func (r *Runestone) Decipher(transaction *wire.MsgTx) *Artifact {
	payload := r.payload(transaction)
	if payload == nil {
		return nil
	}

	integers, err := r.integers(payload)
	if err != nil {
		return &Artifact{
			Cenotaph: &Cenotaph{
				Flaw: FlawVarint,
			},
		}
	}

	message := MessageFromIntegers(transaction, integers)
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

func (r *Runestone) encipher() ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	// Push OP_RETURN
	builder.AddOp(txscript.OP_RETURN)

	// Push MAGIC_NUMBER
	builder.AddInt64(int64(MAGIC_NUMBER))

	if r.Etching != nil {
		// Add etching related data
		// This is a placeholder, you need to replace it with your actual logic
		builder.AddData([]byte("etching data"))
	}

	if r.Mint != nil {
		// Add mint related data
		// This is a placeholder, you need to replace it with your actual logic
		builder.AddData([]byte("mint data"))
	}

	if r.Pointer != nil {
		// Add pointer related data
		// This is a placeholder, you need to replace it with your actual logic
		builder.AddData([]byte("pointer data"))
	}

	for _, edict := range r.Edicts {
		// Add edict related data
		// This is a placeholder, you need to replace it with your actual logic
		builder.AddData([]byte("edict data"))
	}

	return builder.Script()
}

type Payload struct {
	Valid   []byte
	Invalid string
}

func (r *Runestone) payload(transaction *wire.MsgTx) *Payload {
	for _, output := range transaction.TxOut {
		scriptClass, instructions, _, err := txscript.ExtractPkScriptAddrs(output.PkScript, &chaincfg.MainNetParams)
		if err != nil || scriptClass != txscript.NullDataTy {
			continue
		}

		// Check for OP_RETURN
		if len(instructions) < 1 || instructions[0].Opcode.Value != txscript.OP_RETURN {
			continue
		}

		// Check for protocol identifier (Runestone::MAGIC_NUMBER)
		if len(instructions) < 2 || instructions[1].Opcode.Value != MAGIC_NUMBER {
			continue
		}

		// Construct the payload by concatenating remaining data pushes
		var payload []byte
		for _, instruction := range instructions[2:] {
			if instruction.Opcode.Value > txscript.OP_16 {
				return &Payload{Invalid: "Invalid opcode"}
			}
			payload = append(payload, instruction.Data...)
		}

		return &Payload{Valid: payload}
	}

	return nil
}

func (r *Runestone) integers(payload []byte) ([]uint64, error) {
	var integers []uint64
	i := 0

	for i < len(payload) {
		integer, length := binary.Uvarint(payload[i:])
		if length <= 0 {
			return nil, errors.New("invalid varint data")
		}
		integers = append(integers, integer)
		i += length
	}

	return integers, nil
}

type Message struct {
	Flaw   *Flaw
	Edicts []Edict
	Fields map[uint128.Uint128][]uint128.Uint128
}

func MessageFromIntegers(tx *wire.MsgTx, payload []uint128.Uint128) (*Message, error) {
	edicts := []Edict{}
	fields := make(map[uint128.Uint128][]uint128.Uint128)
	var flaw *Flaw

	for i := 0; i < len(payload); i += 2 {
		tag := payload[i]

		if tag == uint128.Uint128(Body) {
			id := RuneId{} // Initialize with default values
			for _, chunk := range payload[i+1:] {
				if len(chunk) != 4 {
					flaw = &Flaw{} // Initialize with appropriate flaw
					break
				}

				next, err := id.Next(chunk[0], chunk[1])
				if err != nil {
					flaw = &Flaw{} // Initialize with appropriate flaw
					break
				}

				edict, err := EdictFromIntegers(tx, next, chunk[2], chunk[3])
				if err != nil {
					flaw = &Flaw{} // Initialize with appropriate flaw
					break
				}

				id = next
				edicts = append(edicts, edict)
			}
			break
		}

		if i+1 >= len(payload) {
			flaw = &Flaw{} // Initialize with appropriate flaw
			break
		}

		value := payload[i+1]
		fields[tag] = append(fields[tag], value)
	}

	if flaw != nil {
		return nil, errors.New("error parsing payload")
	}

	return &Message{
		Flaw:   flaw,
		Edicts: edicts,
		Fields: fields,
	}, nil
}
