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

import "lukechampine.com/uint128"

type Flag uint8

const (
	FlagEtching  Flag = 0
	FlagTerms    Flag = 1
	FlagCenotaph Flag = 127
	FlagTurbo    Flag = 2
)

func (f Flag) Mask() uint128.Uint128 {
	return uint128.From64(1).Lsh(uint(f))
}

func (f Flag) Take(flags *uint128.Uint128) bool {
	mask := f.Mask()
	set := flags.And(mask).Cmp(uint128.Zero) != 0
	*flags = flags.Xor(flags.And(mask))
	// *flags &= !Mask;
	return set
}

func (f Flag) Set(flags *uint128.Uint128) {
	*flags = flags.Or(f.Mask())
}
