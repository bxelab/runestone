package main

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type Utxo struct {
	TxHash   Hash
	Index    uint32
	Value    int64
	PkScript []byte
}

func (u *Utxo) OutPoint() wire.OutPoint {
	h, _ := chainhash.NewHash(u.TxHash[:])
	return wire.OutPoint{
		Hash:  *h,
		Index: u.Index,
	}
}
func (u *Utxo) TxOut() *wire.TxOut {
	return wire.NewTxOut(u.Value, u.PkScript)
}

type UtxoList []*Utxo

func (l UtxoList) Add(utxo *Utxo) UtxoList {
	return append(l, utxo)
}
func (l UtxoList) FetchPrevOutput(o wire.OutPoint) *wire.TxOut {
	for _, utxo := range l {
		if bytes.Equal(utxo.TxHash[:], o.Hash[:]) && utxo.Index == o.Index {
			return wire.NewTxOut(utxo.Value, utxo.PkScript)
		}
	}
	return nil
}
func (u *Utxo) String() string {
	return fmt.Sprintf("TxHash: %s, Index: %d, Value: %d, PkScript: %x", u.TxHash, u.Index, u.Value, u.PkScript)
}
