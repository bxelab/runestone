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

import "errors"

type Flaw int

const (
	EdictOutput Flaw = iota
	EdictRuneId
	InvalidScript
	Opcode
	SupplyOverflow
	TrailingIntegers
	TruncatedField
	UnrecognizedEvenTag
	UnrecognizedFlag
	Varint
)

func FlawP(f Flaw) *Flaw {
	return &f
}

var flawToString = map[Flaw]string{
	EdictOutput:         "edict output greater than transaction output count",
	EdictRuneId:         "invalid Rune ID in edict",
	InvalidScript:       "invalid script in OP_RETURN",
	Opcode:              "non-pushdata opcode in OP_RETURN",
	SupplyOverflow:      "supply overflows u128",
	TrailingIntegers:    "trailing integers in body",
	TruncatedField:      "field with missing Value",
	UnrecognizedEvenTag: "unrecognized even tag",
	UnrecognizedFlag:    "unrecognized field",
	Varint:              "invalid varint",
}

func (f Flaw) String() string {
	return flawToString[f]
}
func (f Flaw) Error() error {
	return errors.New(f.String())
}
func NewFlaw(s string) Flaw {
	for k, v := range flawToString {
		if v == s {
			return k
		}
	}
	return -1
}
