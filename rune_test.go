package go_runestone

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
	"lukechampine.com/uint128"
)

func TestRoundTrip(t *testing.T) {
	testCase := func(n uint128.Uint128, s string) {
		r := Rune{Value: n}
		if r.String() != s {
			t.Errorf("Rune(%v).String() = %v, want %v", n, r.String(), s)
		}
		parsedRune, err := RuneFromString(s)
		if err != nil {
			t.Errorf("RuneFromString(%v) returned error: %v", s, err)
		}
		if parsedRune != r {
			t.Errorf("RuneFromString(%v) = %v, want %v", s, parsedRune, r)
		}
	}

	testCase(uint128.From64(0), "A")
	testCase(uint128.From64(1), "B")
	testCase(uint128.From64(2), "C")
	testCase(uint128.From64(3), "D")
	testCase(uint128.From64(4), "E")
	testCase(uint128.From64(5), "F")
	testCase(uint128.From64(6), "G")
	testCase(uint128.From64(7), "H")
	testCase(uint128.From64(8), "I")
	testCase(uint128.From64(9), "J")
	testCase(uint128.From64(10), "K")
	testCase(uint128.From64(11), "L")
	testCase(uint128.From64(12), "M")
	testCase(uint128.From64(13), "N")
	testCase(uint128.From64(14), "O")
	testCase(uint128.From64(15), "P")
	testCase(uint128.From64(16), "Q")
	testCase(uint128.From64(17), "R")
	testCase(uint128.From64(18), "S")
	testCase(uint128.From64(19), "T")
	testCase(uint128.From64(20), "U")
	testCase(uint128.From64(21), "V")
	testCase(uint128.From64(22), "W")
	testCase(uint128.From64(23), "X")
	testCase(uint128.From64(24), "Y")
	testCase(uint128.From64(25), "Z")
	testCase(uint128.From64(26), "AA")
	testCase(uint128.From64(27), "AB")
	testCase(uint128.From64(51), "AZ")
	testCase(uint128.From64(52), "BA")

	testCase(uint128.Max.Sub64(2), "BCGDENLQRQWDSLRUGSNLBTMFIJAT")
	testCase(uint128.Max.Sub64(1), "BCGDENLQRQWDSLRUGSNLBTMFIJAU")
	testCase(uint128.Max, "BCGDENLQRQWDSLRUGSNLBTMFIJAV")
}
func TestFromStrError(t *testing.T) {
	_, err := RuneFromString("BCGDENLQRQWDSLRUGSNLBTMFIJAW")
	assert.Error(t, err)

	_, err = RuneFromString("BCGDENLQRQWDSLRUGSNLBTMFIJAVX")
	assert.Error(t, err)

	_, err = RuneFromString("x")
	assert.Error(t, err)
}
func TestMainnetMinimumAtHeight(t *testing.T) {
	testCase := func(height uint32, expected string) {
		t.Helper()
		r := MinimumAtHeight(wire.MainNet, uint64(height))
		if r.String() != expected {
			t.Errorf("minimumAtHeight(%d) = %s, want %s", height, r.String(), expected)
		}
	}

	start := SUBSIDY_HALVING_INTERVAL * 4
	end := start + SUBSIDY_HALVING_INTERVAL
	interval := SUBSIDY_HALVING_INTERVAL / 12

	testCase(0, "AAAAAAAAAAAAA")
	testCase(start/2, "AAAAAAAAAAAAA")
	testCase(start, "ZZYZXBRKWXVA")
	testCase(start+1, "ZZXZUDIVTVQA")
	testCase(end-1, "A")
	testCase(end, "A")
	testCase(end+1, "A")
	testCase(^uint32(0), "A")

	testCase(start+interval*0-1, "AAAAAAAAAAAAA")
	testCase(start+interval*0+0, "ZZYZXBRKWXVA")
	testCase(start+interval*0+1, "ZZXZUDIVTVQA")

	testCase(start+interval*1-1, "AAAAAAAAAAAA")
	testCase(start+interval*1+0, "ZZYZXBRKWXV")
	testCase(start+interval*1+1, "ZZXZUDIVTVQ")

	testCase(start+interval*2-1, "AAAAAAAAAAA")
	testCase(start+interval*2+0, "ZZYZXBRKWY")
	testCase(start+interval*2+1, "ZZXZUDIVTW")

	testCase(start+interval*3-1, "AAAAAAAAAA")
	testCase(start+interval*3+0, "ZZYZXBRKX")
	testCase(start+interval*3+1, "ZZXZUDIVU")

	testCase(start+interval*4-1, "AAAAAAAAA")
	testCase(start+interval*4+0, "ZZYZXBRL")
	testCase(start+interval*4+1, "ZZXZUDIW")

	testCase(start+interval*5-1, "AAAAAAAA")
	testCase(start+interval*5+0, "ZZYZXBS")
	testCase(start+interval*5+1, "ZZXZUDJ")

	testCase(start+interval*6-1, "AAAAAAA")
	testCase(start+interval*6+0, "ZZYZXC")
	testCase(start+interval*6+1, "ZZXZUE")

	testCase(start+interval*7-1, "AAAAAA")
	testCase(start+interval*7+0, "ZZYZY")
	testCase(start+interval*7+1, "ZZXZV")

	testCase(start+interval*8-1, "AAAAA")
	testCase(start+interval*8+0, "ZZZA")
	testCase(start+interval*8+1, "ZZYA")

	testCase(start+interval*9-1, "AAAA")
	testCase(start+interval*9+0, "ZZZ")
	testCase(start+interval*9+1, "ZZY")

	testCase(start+interval*10-2, "AAC")
	testCase(start+interval*10-1, "AAA")
	testCase(start+interval*10+0, "AAA")
	testCase(start+interval*10+1, "AAA")
	testCase(start+interval*10+interval/2, "NA")

	testCase(start+interval*11-2, "AB")
	testCase(start+interval*11-1, "AA")
	testCase(start+interval*11+0, "AA")
	testCase(start+interval*11+1, "AA")
	testCase(start+interval*11+interval/2, "N")

	testCase(start+interval*12-2, "B")
	testCase(start+interval*12-1, "A")
	testCase(start+interval*12+0, "A")
	testCase(start+interval*12+1, "A")
}
func TestMinimumAtHeight(t *testing.T) {
	testCase := func(network wire.BitcoinNet, height uint32, minimum string) {
		t.Helper()
		result := MinimumAtHeight(network, uint64(height)).String()
		if result != minimum {
			t.Errorf("MinimumAtHeight(%v, %v) = %v, want %v", network, height, result, minimum)
		}
	}

	testCase(wire.TestNet3, 0, "AAAAAAAAAAAAA")
	testCase(wire.TestNet3, SUBSIDY_HALVING_INTERVAL*12-1, "AAAAAAAAAAAAA")
	testCase(wire.TestNet3, SUBSIDY_HALVING_INTERVAL*12, "ZZYZXBRKWXVA")
	testCase(wire.TestNet3, SUBSIDY_HALVING_INTERVAL*12+1, "ZZXZUDIVTVQA")

	// Assuming Signet and Regtest are defined as custom Bitcoin networks in your code
	testCase(wire.SimNet, 0, "ZZYZXBRKWXVA")
	testCase(wire.SimNet, 1, "ZZXZUDIVTVQA")

	testCase(wire.TestNet, 0, "ZZYZXBRKWXVA")
	testCase(wire.TestNet, 1, "ZZXZUDIVTVQA")
}

//TODO:serde

func TestReserved(t *testing.T) {
	assert := assert.New(t)

	// Assuming you have defined Rune::RESERVED as a constant
	const A = "AAAAAAAAAAAAAAAAAAAAAAAAAAA"
	r, err := RuneFromString(A)
	assert.NoError(err)
	assert.Equal(r, Rune{RESERVED}, "Rune value should be RESERVED")

	assert.Equal(Reserved(0, 0), Rune{RESERVED}, "Rune.Reserved(0, 0) should return RESERVED")

	assert.Equal(Reserved(0, 1), Rune{RESERVED.Add64(1)}, "Rune.Reserved(0, 1) should return RESERVED + 1")

	assert.Equal(Reserved(1, 0), Rune{RESERVED.Add64(1 << 32)}, "Rune.Reserved(1, 0) should return RESERVED + (1 << 32)")

	assert.Equal(Reserved(1, 1), Rune{RESERVED.Add64(1 << 32).Add64(1)}, "Rune.Reserved(1, 1) should return RESERVED + (1 << 32) + 1")
	// Rune::reserved(u64::MAX, u32::MAX),
	r1 := Reserved(math.MaxUint64, math.MaxUint32)
	//      Rune(Rune::RESERVED + (u128::from(u64::MAX) << 32 | u128::from(u32::MAX))),
	r2 := Rune{RESERVED.Add(uint128.From64(math.MaxUint64).Lsh(32).Or(uint128.From64(math.MaxUint32)))}
	assert.Equal(r1, r2, "Rune.Reserved(u64::MAX, u32::MAX) should return RESERVED + (u128::from(u64::MAX) << 32 | u128::from(u32::MAX))")
}
func TestIsReserved(t *testing.T) {
	assert := assert.New(t)

	r, _ := RuneFromString("A")
	assert.Equal(r.IsReserved(), false, "Rune 'A' should not be reserved")

	r, _ = RuneFromString("ZZZZZZZZZZZZZZZZZZZZZZZZZZ")
	assert.Equal(r.IsReserved(), false, "Rune 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ' should not be reserved")

	r, _ = RuneFromString("AAAAAAAAAAAAAAAAAAAAAAAAAAA")
	assert.Equal(r.IsReserved(), true, "Rune 'AAAAAAAAAAAAAAAAAAAAAAAAAAA' should be reserved")

	r, _ = RuneFromString("AAAAAAAAAAAAAAAAAAAAAAAAAAB")
	assert.Equal(r.IsReserved(), true, "Rune 'AAAAAAAAAAAAAAAAAAAAAAAAAAB' should be reserved")

	r, _ = RuneFromString("BCGDENLQRQWDSLRUGSNLBTMFIJAV")
	assert.Equal(r.IsReserved(), true, "Rune 'BCGDENLQRQWDSLRUGSNLBTMFIJAV' should be reserved")
}
func TestSteps(t *testing.T) {
	assert := assert.New(t)

	// Assuming you have defined Rune::STEPS as a constant array
	for i := 0; ; i++ {
		repeatedA := strings.Repeat("A", i+1)
		r, err := RuneFromString(repeatedA)

		if err != nil {
			assert.Equal(len(STEPS), i, "Rune::STEPS length should be equal to i")
			break
		} else {
			assert.Equal(Rune{STEPS[i]}, r, fmt.Sprintf("Rune(STEPS[%d]) should be equal to Rune with %s", i, repeatedA))
		}
	}
}
func TestCommitment(t *testing.T) {
	assert := assert.New(t)

	// Anonymous function testCase
	testCase := func(r Rune, expected []byte) {
		assert.Equal(r.Commitment(), expected, "Rune commitment should be equal to expected")
	}

	r := Rune{uint128.From64(0)}
	testCase(r, []byte{})

	r = Rune{uint128.From64(1)}
	testCase(r, []byte{1})

	r = Rune{uint128.From64(255)}
	testCase(r, []byte{255})

	r = Rune{uint128.From64(256)}
	testCase(r, []byte{0, 1})

	r = Rune{uint128.From64(65535)}
	testCase(r, []byte{255, 255})

	r = Rune{uint128.From64(65536)}
	testCase(r, []byte{0, 0, 1})

	r = Rune{uint128.Max}
	expected := make([]byte, 16)
	for i := range expected {
		expected[i] = 255
	}
	testCase(r, expected)
}
