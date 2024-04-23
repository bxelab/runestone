package go_runestone

import (
	"errors"

	"github.com/btcsuite/btcd/wire"
)

type Edict struct {
	ID     RuneId
	Amount uint64
	Output uint32
}

func EdictFromIntegers(tx *wire.MsgTx, id RuneId, amount uint64, output uint64) (*Edict, error) {
	if output > uint64(len(tx.TxOut)) {
		return nil, errors.New("output is greater than transaction output count")
	}

	return &Edict{
		ID:     id,
		Amount: amount,
		Output: uint32(output),
	}, nil
}
