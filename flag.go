package go_runestone

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
	*flags = flags.Add(mask)
	//TODO: *flags &= !Mask;
	return set
}

func (f Flag) Set(flags *uint128.Uint128) {
	*flags = flags.Or(f.Mask())
}
