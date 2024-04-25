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
	"math"

	"github.com/btcsuite/btcd/wire"
	"lukechampine.com/uint128"
)

type Edict struct {
	ID     RuneId
	Amount uint128.Uint128
	Output uint32
}

func EdictFromIntegers(tx *wire.MsgTx, id RuneId, amount uint128.Uint128, output uint128.Uint128) (*Edict, error) {
	if output.Hi > 0 || output.Lo > math.MaxUint32 {
		return nil, errors.New("output overflow")
	}
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
