package go_runestone

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/wire"
	"lukechampine.com/uint128"
)

type Rune struct {
	Value uint128.Uint128
}

func Uint128FromString(s string) uint128.Uint128 {
	i, _ := uint128.FromString(s)
	return i
}

var RESERVED = Uint128FromString("6402364363415443603228541259936211926")
var STEPS = []uint128.Uint128{
	Uint128FromString("0"),
	Uint128FromString("26"),
	Uint128FromString("702"),
	Uint128FromString("18278"),
	Uint128FromString("475254"),
	Uint128FromString("12356630"),
	Uint128FromString("321272406"),
	Uint128FromString("8353082582"),
	Uint128FromString("217180147158"),
	Uint128FromString("5646683826134"),
	Uint128FromString("146813779479510"),
	Uint128FromString("3817158266467286"),
	Uint128FromString("99246114928149462"),
	Uint128FromString("2580398988131886038"),
	Uint128FromString("67090373691429037014"),
	Uint128FromString("1744349715977154962390"),
	Uint128FromString("45353092615406029022166"),
	Uint128FromString("1179180408000556754576342"),
	Uint128FromString("30658690608014475618984918"),
	Uint128FromString("797125955808376366093607894"),
	Uint128FromString("20725274851017785518433805270"),
	Uint128FromString("538857146126462423479278937046"),
	Uint128FromString("14010285799288023010461252363222"),
	Uint128FromString("364267430781488598271992561443798"),
	Uint128FromString("9470953200318703555071806597538774"),
	Uint128FromString("246244783208286292431866971536008150"),
	Uint128FromString("6402364363415443603228541259936211926"),
	Uint128FromString("166461473448801533683942072758341510102"),
}

func NewRune(value uint128.Uint128) Rune {
	return Rune{Value: value}
}

func (r Rune) N() uint128.Uint128 {
	return r.Value
}

const SUBSIDY_HALVING_INTERVAL uint32 = 210_000

func FirstRuneHeight(network wire.BitcoinNet) uint32 {
	var multiplier uint32
	switch network {
	case wire.MainNet:
		multiplier = 4
	case wire.TestNet, wire.SimNet:
		multiplier = 0
	case wire.TestNet3:
		multiplier = 12
	default:
		multiplier = 0
	}
	return SUBSIDY_HALVING_INTERVAL * multiplier
}
func MinimumAtHeight(chain wire.BitcoinNet, height uint64) Rune {
	offset := height + 1
	const interval uint32 = SUBSIDY_HALVING_INTERVAL / 12
	start := FirstRuneHeight(chain)
	end := start + SUBSIDY_HALVING_INTERVAL
	if offset < uint64(start) {
		return Rune{STEPS[12]}
	}
	if offset >= uint64(end) {
		return Rune{}
	}
	progress := offset - uint64(start)
	length := 12 - progress/uint64(interval)
	endStep := STEPS[length-1]
	startStep := STEPS[length]
	remainder := progress % uint64(interval)

	//val := startStep - ((startStep - endStep) * remainder / uint64(interval))
	val := startStep.Sub(startStep.Sub(endStep).Mul(uint128.From64(remainder)).Div(uint128.From64(uint64(interval))))
	return Rune{val}

}

func (r Rune) IsReserved() bool {
	return r.Value.Cmp(RESERVED) >= 0
}

func Reserved(block uint64, tx uint32) Rune {
	v := RESERVED.Add(uint128.From64(block).Lsh(32).Or(uint128.From64(uint64(tx))))
	return Rune{
		Value: v,
	}
}

func (r Rune) Commitment() []byte {
	bytes := r.Value.Big().Bytes()

	// Reverse bytes to get little-endian representation
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}

	end := len(bytes)
	for end > 0 && bytes[end-1] == 0 {
		end--
	}

	return bytes[:end]
}
func (r Rune) String() string {
	n := r.Value
	if n.Cmp(uint128.Max) == 0 {
		return "BCGDENLQRQWDSLRUGSNLBTMFIJAV"
	}

	n = n.Add64(1)
	var symbol strings.Builder
	for n.Cmp(uint128.Zero) > 0 {
		index := n.Sub64(1).Mod64(26)

		symbol.WriteByte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"[index])
		n = n.Sub64(1).Div64(26)
	}

	// Reverse the string
	runes := []rune(symbol.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func RuneFromString(s string) (*Rune, error) {
	x := big.NewInt(0)
	tmp := big.NewInt(0)
	for i, c := range s {
		if i > 0 {
			x.Add(x, tmp.SetInt64(1))
		}
		x.Mul(x, tmp.SetInt64(26))
		if x.BitLen() > 128 {
			return nil, errors.New("overflow")
		}
		if c >= 'A' && c <= 'Z' {
			x.Add(x, tmp.SetInt64(int64(c-'A')))
			if x.BitLen() > 128 {
				return nil, errors.New("overflow")
			}
		} else {
			return nil, fmt.Errorf("invalid character `%c`", c)
		}
	}
	u := uint128.FromBig(x)

	return &Rune{Value: u}, nil
}

type Error struct {
	Character rune
	Range     bool
}

func (e Error) Error() string {
	if e.Range {
		return "name out of range"
	}
	return fmt.Sprintf("invalid character `%c`", e.Character)
}
