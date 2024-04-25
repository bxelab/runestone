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
	"errors"
	"math/big"
	"sort"
	"unicode/utf8"

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
		if payload != nil {
			return &Artifact{
				Cenotaph: &Cenotaph{
					Flaw: &payload.Invalid,
				}}, nil
		}

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
		}, 1)
	if flags == nil {
		//unwrap_or_default
		flags = &uint128.Zero
	}
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
			}, 1)
		//      premine: Tag::Premine.take(&mut fields, |[premine]| Some(premine)),
		etching.Premine, err = TagTake(TagPremine, message.Fields,
			func(uint128s []uint128.Uint128) (*uint128.Uint128, error) {
				return &uint128s[0], nil
			}, 1)
		// rune: Tag::Rune.take(&mut fields, |[rune]| Some(Rune(rune))),
		etching.Rune, err = TagTake(TagRune, message.Fields,
			func(uint128s []uint128.Uint128) (*Rune, error) {
				return &Rune{Value: uint128s[0]}, nil
			}, 1)
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
			}, 1)
		//      symbol: Tag::Symbol.take(&mut fields, |[symbol]| {
		//        char::from_u32(u32::try_from(symbol).ok()?)
		//      }),
		etching.Symbol, err = TagTake(TagSymbol, message.Fields,
			func(uint128s []uint128.Uint128) (*rune, error) {

				symbol := rune(uint32(uint128s[0].Lo))
				if symbol > utf8.MaxRune {
					return nil, errors.New("symbol too high")
				}
				return &symbol, nil
			}, 1)
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
				}, 1)
			terms.Height[0], err = TagTake(TagHeightStart, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				}, 1)
			terms.Height[1], err = TagTake(TagHeightEnd, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				}, 1)
			terms.Amount, err = TagTake(TagAmount, message.Fields,
				func(uint128s []uint128.Uint128) (*uint128.Uint128, error) {
					return &uint128s[0], nil
				}, 1)
			terms.Offset[0], err = TagTake(TagOffsetStart, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				}, 1)
			terms.Offset[1], err = TagTake(TagOffsetEnd, message.Fields,
				func(uint128s []uint128.Uint128) (*uint64, error) {
					h := uint128s[0].Lo
					return &h, nil
				}, 1)
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
			return NewRuneId(block, tx)
		}, 2)
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

		}, 1)
	//if etching
	//      .map(|etching| etching.supply().is_none())
	//      .unwrap_or_default()
	//    {
	//      flaw.get_or_insert(Flaw::SupplyOverflow);
	//    }
	if etching != nil && etching.Supply() == nil {
		message.Flaw = FlawP(SupplyOverflow)

	}
	// if flags != 0 {
	//      flaw.get_or_insert(Flaw::UnrecognizedFlag);
	//    }
	if !flags.IsZero() {
		message.Flaw = FlawP(UnrecognizedFlag)

	}
	//    if fields.keys().any(|tag| tag % 2 == 0) {
	//      flaw.get_or_insert(Flaw::UnrecognizedEvenTag);
	//    }
	for tag := range message.Fields {
		if tag%2 == 0 {
			message.Flaw = FlawP(UnrecognizedEvenTag)
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
		a := &Artifact{
			Cenotaph: &Cenotaph{
				Flaw: message.Flaw,
				Mint: mint,
			},
		}
		if etching != nil {
			a.Cenotaph.Etching = etching.Rune
		}
		return a, nil

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
		if r.Etching.Turbo {
			FlagTurbo.Set(&flags)
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
			payload = append(payload, EncodeChar(*r.Etching.Symbol)...)
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
	if len(r.Edicts) != 0 {
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
	builder.AddOp(MAGIC_NUMBER)
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
		tokenizer := txscript.MakeScriptTokenizer(0, output.PkScript)
		if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != txscript.OP_RETURN {
			// Check for OP_RETURN
			continue
		}
		if !tokenizer.Next() || tokenizer.Err() != nil || tokenizer.Opcode() != MAGIC_NUMBER {
			// Check for protocol identifier (Runestone::MAGIC_NUMBER)
			continue
		}

		// Construct the payload by concatenating remaining data pushes
		var payload []byte
		for tokenizer.Next() {
			//is PushBytes
			if isPushBytes(tokenizer.Opcode()) {
				payload = append(payload, tokenizer.Data()...)
				continue
			} else {
				return &Payload{Invalid: Opcode}, Opcode.Error()
			}

		}
		//Err(_) => {
		//            return Some(Payload::Invalid(Flaw::InvalidScript));
		//          }
		if tokenizer.Err() != nil {
			return &Payload{Invalid: InvalidScript}, InvalidScript.Error()
		}

		return &Payload{Valid: payload}, nil
	}

	return nil, errors.New("no OP_RETURN output found")
}
func isPushBytes(opCode byte) bool {
	return opCode >= txscript.OP_0 && opCode <= txscript.OP_PUSHDATA4
}

func (r *Runestone) integers(payload []byte) ([]uint128.Uint128, error) {
	integers := make([]uint128.Uint128, 0)
	i := 0

	for i < len(payload) {
		integer, length, err := uvarint128(payload[i:])
		if err != nil {
			return nil, err
		}
		integers = append(integers, integer)
		i += length
	}

	return integers, nil
}
func uvarint128(buf []byte) (uint128.Uint128, int, error) {
	n := big.NewInt(0)
	for i, tick := range buf {
		if i > 18 {
			return uint128.Zero, 0, errors.New("varint too long")
		}
		value := uint64(tick) & 0b0111_1111
		if i == 18 && value&0b0111_1100 != 0 {
			return uint128.Zero, 0, errors.New("varint too large")
		}
		temp := new(big.Int).SetUint64(value)
		n.Or(n, temp.Lsh(temp, uint(7*i)))
		if tick&0b1000_0000 == 0 {
			return uint128.FromBig(n), i + 1, nil
		}
	}
	return uint128.Zero, 0, errors.New("varint too short")
}
