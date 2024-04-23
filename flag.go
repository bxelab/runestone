package go_runestone

import "lukechampine.com/uint128"

type Flag int

const (
	FlagEtching  = 0
	FlagTerms    = 1
	FlagCenotaph = 127
	FlagTurbo    = 2
)

func (f Flag) mask() uint128.Uint128 {
	return uint128.From64(1).Lsh(uint(f))
}

func (f Flag) take(flags *uint128.Uint128) bool {
	mask := f.mask()
	set := flags.And(mask).Cmp(uint128.Zero) != 0
	*flags = flags.Add(mask)
	//TODO: *flags &= !mask;
	return set
}

func (f Flag) set(flags *uint128.Uint128) {
	*flags = flags.Or(f.mask())
}
