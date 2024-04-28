package main

import (
	"encoding/hex"
	"errors"
	"unicode/utf8"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/bxelab/runestone"
	"lukechampine.com/uint128"
)

type Config struct {
	PrivateKey string
	FeePerByte int64
	UtxoAmount int64
	Network    string
	RpcUrl     string
	Etching    *struct {
		Rune              string
		Symbol            *string
		Premine           *uint64
		Amount            *uint64
		Cap               *uint64
		Divisibility      *int
		HeightStart       *int
		HeightEnd         *int
		HeightOffsetStart *int
		HeightOffsetEnd   *int
	}
	Mint *struct {
		RuneId string
	}
}

func DefaultConfig() Config {
	return Config{
		FeePerByte: 5,
		UtxoAmount: 1000,
		Network:    "mainnet",
		RpcUrl:     "https://mempool.space/api",
	}

}
func (c Config) GetFeePerByte() int64 {
	if c.FeePerByte == 0 {
		return 5
	}
	return c.FeePerByte
}
func (c Config) GetUtxoAmount() int64 {
	if c.UtxoAmount == 0 {
		return 666
	}
	return c.UtxoAmount
}

func (c Config) GetEtching() (*runestone.Etching, error) {
	if c.Etching == nil {
		return nil, errors.New("Etching config is required")
	}
	if c.Etching.Rune == "" {
		return nil, errors.New("Rune is required")
	}
	if c.Etching.Symbol != nil {
		runeCount := utf8.RuneCountInString(*c.Etching.Symbol)
		if runeCount != 1 {
			return nil, errors.New("Symbol must be a single character")
		}
	}
	etching := &runestone.Etching{}
	r, err := runestone.SpacedRuneFromString(c.Etching.Rune)
	if err != nil {
		return nil, err
	}
	etching.Rune = &r.Rune
	etching.Spacers = &r.Spacers
	if c.Etching.Symbol != nil {
		symbolStr := *c.Etching.Symbol
		symbol := rune(symbolStr[0])
		etching.Symbol = &symbol
	}
	if c.Etching.Premine != nil {
		premine := uint128.From64(*c.Etching.Premine)
		etching.Premine = &premine
	}
	if c.Etching.Amount != nil {
		amount := uint128.From64(*c.Etching.Amount)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Amount = &amount
	}
	if c.Etching.Cap != nil {
		cap := uint128.From64(*c.Etching.Cap)
		etching.Terms.Cap = &cap
	}
	if c.Etching.Divisibility != nil {
		d := uint8(*c.Etching.Divisibility)
		etching.Divisibility = &d
	}
	if c.Etching.HeightStart != nil {
		h := uint64(*c.Etching.HeightStart)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Height[0] = &h
	}
	if c.Etching.HeightEnd != nil {
		h := uint64(*c.Etching.HeightEnd)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Height[1] = &h
	}
	if c.Etching.HeightOffsetStart != nil {
		h := uint64(*c.Etching.HeightOffsetStart)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Offset[0] = &h
	}
	if c.Etching.HeightOffsetEnd != nil {
		h := uint64(*c.Etching.HeightOffsetEnd)
		if etching.Terms == nil {
			etching.Terms = &runestone.Terms{}
		}
		etching.Terms.Offset[1] = &h
	}
	return etching, nil
}
func (c Config) GetMint() (*runestone.RuneId, error) {
	if c.Mint == nil {
		return nil, errors.New("Mint config is required")
	}
	if c.Mint.RuneId == "" {
		return nil, errors.New("RuneId is required")
	}
	runeId, err := runestone.RuneIdFromString(c.Mint.RuneId)
	if err != nil {
		return nil, err
	}
	return runeId, nil
}
func (c Config) GetNetwork() *chaincfg.Params {
	if c.Network == "mainnet" {
		return &chaincfg.MainNetParams
	}
	if c.Network == "testnet" {
		return &chaincfg.TestNet3Params
	}
	if c.Network == "regtest" {
		return &chaincfg.RegressionNetParams
	}
	if c.Network == "signet" {
		return &chaincfg.SigNetParams
	}
	panic("unknown network")
}

func (c Config) GetPrivateKeyAddr() (*btcec.PrivateKey, string, error) {
	if c.PrivateKey == "" {
		return nil, "", errors.New("PrivateKey is required")
	}
	pkBytes, err := hex.DecodeString(c.PrivateKey)
	if err != nil {
		return nil, "", err
	}
	privKey, pubKey := btcec.PrivKeyFromBytes(pkBytes)
	if err != nil {
		return nil, "", err
	}
	tapKey := txscript.ComputeTaprootKeyNoScript(pubKey)
	addr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(tapKey), c.GetNetwork(),
	)
	if err != nil {
		return nil, "", err
	}
	address := addr.EncodeAddress()
	return privKey, address, nil
}
