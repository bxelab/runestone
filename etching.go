package go_runestone

import "math/big"

type Terms struct {
	amount *big.Int
	cap    *big.Int
	height [2]*big.Int
	offset [2]*big.Int
}

type Etching struct {
	divisibility *uint8
	premine      *big.Int
	rune         *Rune
	spacers      *uint32
	symbol       *rune
	terms        *Terms
	turbo        bool
}

const (
	MaxDivisibility = 38
	MaxSpacers      = 0b00000111_11111111_11111111_11111111
)

func (e *Etching) Supply() *big.Int {
	premine := big.NewInt(0)
	if e.premine != nil {
		premine = e.premine
	}

	cap := big.NewInt(0)
	amount := big.NewInt(0)
	if e.terms != nil {
		if e.terms.cap != nil {
			cap = e.terms.cap
		}
		if e.terms.amount != nil {
			amount = e.terms.amount
		}
	}

	supply := new(big.Int).Mul(cap, amount)
	supply.Add(supply, premine)

	return supply
}
