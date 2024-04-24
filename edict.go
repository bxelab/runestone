package go_runestone

import (
	"errors"

	"github.com/btcsuite/btcd/wire"
	"lukechampine.com/uint128"
)

type Edict struct {
	ID     RuneId
	Amount uint128.Uint128
	Output uint32
}

func EdictFromIntegers(tx *wire.MsgTx, id RuneId, amount uint128.Uint128, output uint128.Uint128) (*Edict, error) {
	output32 := uint32(output.Lo)
	if output32 > uint32(len(tx.TxOut)) {
		return nil, errors.New("output is greater than transaction output count")
	}

	return &Edict{
		ID:     id,
		Amount: amount,
		Output: output32,
	}, nil
}
