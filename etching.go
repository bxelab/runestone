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
	"lukechampine.com/uint128"
)

type Terms struct {
	Amount *uint128.Uint128
	Cap    *uint128.Uint128
	Height [2]*uint64
	Offset [2]*uint64
}

type Etching struct {
	Divisibility *uint8
	Premine      *uint128.Uint128
	Rune         *Rune
	Spacers      *uint32
	Symbol       *rune
	Terms        *Terms
	Turbo        bool
}

const (
	MaxDivisibility = 38
	MaxSpacers      = 0b00000111_11111111_11111111_11111111
)

func (e *Etching) Supply() *uint128.Uint128 {
	//cover panic
	defer func() {
		if r := recover(); r != nil {
			return
		}
	}()
	premine := uint128.Zero
	if e.Premine != nil {
		premine = *e.Premine
	}

	cap := uint128.Zero
	amount := uint128.Zero
	if e.Terms != nil {
		if e.Terms.Cap != nil {
			cap = *e.Terms.Cap
		}
		if e.Terms.Amount != nil {
			amount = *e.Terms.Amount
		}
	}

	supply := cap.Mul(amount)
	supply = supply.Add(premine)

	return &supply
}
