package go_runestone

import (
	"encoding/binary"
	"errors"
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

func (r *Runestone) Decipher(transaction *wire.MsgTx) (*Artifact, error) {
	payload, err := r.payload(transaction)
	if err != nil {
		return nil, err
	}

	integers, err := r.integers(payload.Valid)
	if err != nil {
		flaw := Varint
		return &Artifact{
			Cenotaph: &Cenotaph{
				Flaw: &flaw,
			},
		}, err
	}

	message, err := MessageFromIntegers(transaction, integers)
	flags, err := TagTake(TagFlags, message.Fields,
		func(uint128s []uint128.Uint128) (*uint128.Uint128, error) {
			return &uint128s[0], nil
		})
	var etching *Etching
	if FlagEtching.Take(flags) {
		etching = &Etching{}
		etching.Divisibility, err = TagTake(TagDivisibility, message.Fields,
			func(uint128s []uint128.Uint128) (*uint8, error) {
				divisibility := uint8(uint128s[0].Lo)
				if divisibility > MaxDivisibility {
					return nil, errors.New("divisibility too high")
				}
				return &divisibility, nil
			})
		//      premine: Tag::Premine.take(&mut fields, |[premine]| Some(premine)),
		etching.Premine, err = TagTake(TagPremine, message.Fields,
			func(uint128s []uint128.Uint128) (*uint128.Uint128, error) {
				return &uint128s[0], nil
			})
		// rune: Tag::Rune.take(&mut fields, |[rune]| Some(Rune(rune))),
		etching.Rune, err = TagTake(TagRune, message.Fields,
			func(uint128s []uint128.Uint128) (*Rune, error) {
				return &Rune{Value: uint128s[0]}, nil
			})
		//      spacers: Tag::Spacers.take(&mut fields, |[spacers]| {
		//        let spacers = u32::try_from(spacers).ok()?;
		//        (spacers <= Etching::MAX_SPACERS).then_some(spacers)
		//      }),
		etching.Spacers, err = TagTake(TagSpacers, message.Fields,
			func(uint128s []uint128.Uint128) (*uint32, error) {
				spacers := uint32(uint128s[0].Lo)
				if spacers > MaxSpacers {
					return nil, errors.New("spacers too high")
				}
				return &spacers, nil
			})
		//      symbol: Tag::Symbol.take(&mut fields, |[symbol]| {
		//        char::from_u32(u32::try_from(symbol).ok()?)
		//      }),
		etching.Symbol, err = TagTake(TagSymbol, message.Fields,
			func(uint128s []uint128.Uint128) (*rune, error) {
				symbol := rune(uint32(uint128s[0].Lo))
				return &symbol, nil
			})
		//      terms: Flag::Terms.take(&mut flags).then(|| Terms {
		//        cap: Tag::Cap.take(&mut fields, |[cap]| Some(cap)),
		//        height: (
		//          Tag::HeightStart.take(&mut fields, |[start_height]| {
		//            u64::try_from(start_height).ok()
		//          }),
		//          Tag::HeightEnd.take(&mut fields, |[start_height]| {
		//            u64::try_from(start_height).ok()
		//          }),
		//        ),
		//        amount: Tag::Amount.take(&mut fields, |[amount]| Some(amount)),
		//        offset: (
		//          Tag::OffsetStart.take(&mut fields, |[start_offset]| {
		//            u64::try_from(start_offset).ok()
		//          }),
		//          Tag::OffsetEnd.take(&mut fields, |[end_offset]| u64::try_from(end_offset).ok()),
		//        ),
		//      }),
		if FlagTerms.Take(flags) {
			terms := Terms{}
			terms.Cap, err = TagTake(TagCap, message.Fields,
				func(uint128s []uint128.Uint128) (*uint128.Uint128, error) {
					return &uint128s[0], nil
				})
			terms.Height[0], err = TagTake(TagHeightStart, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				})
			terms.Height[1], err = TagTake(TagHeightEnd, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				})
			terms.Amount, err = TagTake(TagAmount, message.Fields,
				func(uint128s []uint128.Uint128) (*uint128.Uint128, error) {
					return &uint128s[0], nil
				})
			terms.Offset[0], err = TagTake(TagOffsetStart, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				})
			terms.Offset[1], err = TagTake(TagOffsetEnd, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				})
			etching.Terms = &terms
		}
		//      turbo: Flag::Turbo.take(&mut flags),
		etching.Turbo = FlagTurbo.Take(flags)
	}
	// let mint = Tag::Mint.take(&mut fields, |[block, tx]| {
	//      RuneId::new(block.try_into().ok()?, tx.try_into().ok()?)
	//    });
	mint, err := TagTake(TagMint, message.Fields,
		func(uint128s []uint128.Uint128) (*RuneId, error) {
			block := uint64(uint128s[0].Lo)
			tx := uint32(uint128s[1].Lo)
			return &RuneId{block, tx}, nil

		})
	//let pointer = Tag::Pointer.take(&mut fields, |[pointer]| {
	//      let pointer = u32::try_from(pointer).ok()?;
	//      (u64::from(pointer) < u64::try_from(transaction.output.len()).unwrap()).then_some(pointer)
	//    });
	pointer, err := TagTake(TagPointer, message.Fields,
		func(uint128s []uint128.Uint128) (*uint32, error) {
			pointer := uint32(uint128s[0].Lo)
			if uint64(pointer) < uint64(len(transaction.TxOut)) {
				return &pointer, nil
			}
			return nil, errors.New("pointer too high")

		})
	//if etching
	//      .map(|etching| etching.supply().is_none())
	//      .unwrap_or_default()
	//    {
	//      flaw.get_or_insert(Flaw::SupplyOverflow);
	//    }
	if etching != nil && etching.Supply().IsZero() {
		*message.Flaw = SupplyOverflow

	}
	// if flags != 0 {
	//      flaw.get_or_insert(Flaw::UnrecognizedFlag);
	//    }
	if !flags.IsZero() {
		*message.Flaw = UnrecognizedFlag

	}
	//    if fields.keys().any(|tag| tag % 2 == 0) {
	//      flaw.get_or_insert(Flaw::UnrecognizedEvenTag);
	//    }
	for tag := range message.Fields {
		if tag%2 == 0 {
			*message.Flaw = UnrecognizedEvenTag
		}

	}
	//if let Some(flaw) = flaw {
	//      return Some(Artifact::Cenotaph(Cenotaph {
	//        flaw: Some(flaw),
	//        mint,
	//        etching: etching.and_then(|etching| etching.rune),
	//      }));
	//    }
	if message.Flaw != nil {
		return &Artifact{
			Cenotaph: &Cenotaph{
				Flaw:    message.Flaw,
				Mint:    mint,
				Etching: etching.Rune,
			},
		}, nil

	}

	return &Artifact{
		Runestone: &Runestone{
			Edicts:  message.Edicts,
			Etching: etching,
			Mint:    mint,
			Pointer: pointer,
		},
	}, nil
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
			payload = append(payload, EncodeUint128(edict.Amount)...)
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

func runeToBytes(r *rune) []byte {
	return []byte(string(*r))
}
