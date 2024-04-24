package go_runestone

import (
	"fmt"
	"strconv"
	"strings"

	"lukechampine.com/uint128"
)

type RuneId struct {
	Block uint64
	Tx    uint32
}

func NewRuneId(block uint64, tx uint32) *RuneId {
	if block == 0 && tx > 0 {
		return nil
	}
	return &RuneId{Block: block, Tx: tx}
}

func (r RuneId) Delta(next RuneId) (uint64, uint32, error) {
	if next.Block < r.Block {
		return 0, 0, fmt.Errorf("next block is less than current block")
	}
	block := next.Block - r.Block
	var tx uint32
	if block == 0 {
		if next.Tx < r.Tx {
			return 0, 0, fmt.Errorf("next tx is less than current tx")
		}
		tx = next.Tx - r.Tx
	} else {
		tx = next.Tx
	}
	return block, tx, nil
}

func (r RuneId) Next(block uint128.Uint128, tx uint128.Uint128) (*RuneId, error) {
	newBlock := r.Block + block.Lo
	var newTx uint32
	if block.IsZero() {
		newTx = r.Tx + uint32(tx.Lo)
	} else {
		newTx = uint32(tx.Lo)
	}
	return NewRuneId(newBlock, newTx), nil
}

func (r RuneId) String() string {
	return fmt.Sprintf("%d:%d", r.Block, r.Tx)
}

func ParseRuneId(s string) (*RuneId, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid format")
	}
	block, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid block: %v", err)
	}
	tx, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid tx: %v", err)
	}
	return NewRuneId(block, uint32(tx)), nil
}
