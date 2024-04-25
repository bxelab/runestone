module github.com/bxelab/runestone/cmd/runestonecli

go 1.22.2

require github.com/bxelab/runestone v0.0.0-20240425113004-bea3419a6a3e

require (
	github.com/btcsuite/btcd v0.24.0 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.1.3 // indirect
	github.com/btcsuite/btcd/btcutil v1.1.5 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
	lukechampine.com/uint128 v1.3.0 // indirect
)
replace (
	github.com/bxelab/runestone => ../../
)