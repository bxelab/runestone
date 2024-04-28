package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	defaultSequenceNum    = wire.MaxTxInSequenceNum - 10
	defaultRevealOutValue = int64(330) // 500 sat, ord default 10000

	MaxStandardTxWeight = blockchain.MaxBlockWeight / 10
)

func BuildInscriptionTxs(privateKey *btcec.PrivateKey, utxo []*Utxo, mime string, content []byte, feeRate int64, revealValue int64, net *chaincfg.Params) ([]byte, []byte, error) {
	//build 2 tx, 1 transfer BTC to taproot address, 2 inscription transfer taproot address to another address
	pubKey := privateKey.PubKey()
	receiver, err := getP2TRAddress(pubKey, net)
	if err != nil {
		return nil, nil, err
	}
	// 1. build inscription script
	inscriptionScript, err := CreateInscriptionScript(pubKey, mime, content)
	if err != nil {
		return nil, nil, err
	}
	inscriptionAddress, err := GetTapScriptAddress(pubKey, inscriptionScript, net)
	if err != nil {
		return nil, nil, err
	}
	inscriptionPkScript, _ := txscript.PayToAddrScript(inscriptionAddress)
	// 2. build reveal tx
	revealTx, totalPrevOutput, err := buildEmptyRevealTx(receiver, inscriptionScript, revealValue, feeRate, nil)
	if err != nil {
		return nil, nil, err
	}
	// 3. build commit tx
	out := &wire.TxOut{
		Value:    totalPrevOutput,
		PkScript: inscriptionPkScript,
	}
	commitTx, err := buildCommitTx(utxo, out, feeRate, nil, true)
	if err != nil {
		return nil, nil, err
	}
	// 4. completeRevealTx
	revealTx, err = completeRevealTx(privateKey, commitTx, revealTx, inscriptionScript)
	if err != nil {
		return nil, nil, err
	}
	// 5. sign commit tx
	commitTx, err = signCommitTx(privateKey, utxo, commitTx)
	if err != nil {
		return nil, nil, err
	}
	// 6. serialize
	commitTxBytes, err := serializeTx(commitTx)
	if err != nil {
		return nil, nil, err
	}
	revealTxBytes, err := serializeTx(revealTx)
	if err != nil {
		return nil, nil, err
	}
	return commitTxBytes, revealTxBytes, nil
}
func BuildRuneEtchingTxs(privateKey *btcec.PrivateKey, utxo []*Utxo, runeOpReturnData []byte, runeCommitment []byte,
	feeRate int64, revealValue int64, net *chaincfg.Params, toAddr string) ([]byte, []byte, error) {
	//build 2 tx, 1 transfer BTC to taproot address, 2 inscription transfer taproot address to another address
	pubKey := privateKey.PubKey()
	receiver, err := btcutil.DecodeAddress(toAddr, net)
	if err != nil {
		return nil, nil, err
	}
	// 1. build inscription script
	inscriptionScript, err := CreateCommitmentScript(pubKey, runeCommitment)
	if err != nil {
		return nil, nil, err
	}
	inscriptionAddress, err := GetTapScriptAddress(pubKey, inscriptionScript, net)
	if err != nil {
		return nil, nil, err
	}
	inscriptionPkScript, _ := txscript.PayToAddrScript(inscriptionAddress)
	// 2. build reveal tx
	revealTx, totalPrevOutput, err := buildEmptyRevealTx(receiver, inscriptionScript, revealValue, feeRate, runeOpReturnData)
	if err != nil {
		return nil, nil, err
	}
	// 3. build commit tx
	out := &wire.TxOut{
		Value:    totalPrevOutput,
		PkScript: inscriptionPkScript,
	}
	commitTx, err := buildCommitTx(utxo, out, feeRate, nil, true)
	if err != nil {
		return nil, nil, err
	}
	// 4. completeRevealTx
	revealTx, err = completeRevealTx(privateKey, commitTx, revealTx, inscriptionScript)
	if err != nil {
		return nil, nil, err
	}
	// 5. sign commit tx
	commitTx, err = signCommitTx(privateKey, utxo, commitTx)
	if err != nil {
		return nil, nil, err
	}
	// 6. serialize
	commitTxBytes, err := serializeTx(commitTx)
	if err != nil {
		return nil, nil, err
	}
	revealTxBytes, err := serializeTx(revealTx)
	if err != nil {
		return nil, nil, err
	}
	return commitTxBytes, revealTxBytes, nil
}
func BuildTransferBTCTx(privateKey *btcec.PrivateKey, utxo []*Utxo, toAddr string, toAmount, feeRate int64, net *chaincfg.Params, runeData []byte) ([]byte, error) {
	address, err := btcutil.DecodeAddress(toAddr, net)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		return nil, err
	}
	// 1. build tx
	transferTx, err := buildCommitTx(utxo, wire.NewTxOut(toAmount, pkScript), feeRate, runeData, true)
	if err != nil {
		return nil, err
	}
	// 2.sign tx
	transferTx, err = signCommitTx(privateKey, utxo, transferTx)
	if err != nil {
		return nil, err
	}
	// 3. serialize
	commitTxBytes, err := serializeTx(transferTx)
	if err != nil {
		return nil, err
	}
	return commitTxBytes, nil
}

func VerifyTx(rawTx string, prevTxOutScript []byte, prevTxOutValue int64) error {
	txBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		fmt.Println("Error decoding transaction:", err)
		return err
	}

	var tx wire.MsgTx
	if err := tx.Deserialize(bytes.NewReader(txBytes)); err != nil {
		fmt.Println("Error deserializing transaction:", err)
		return err
	}

	for i, _ := range tx.TxIn {

		outputFetcher := txscript.NewCannedPrevOutputFetcher(prevTxOutScript, prevTxOutValue)
		sigHashes := txscript.NewTxSigHashes(&tx, outputFetcher)
		vm, err := txscript.NewEngine(prevTxOutScript, &tx, i, txscript.StandardVerifyFlags, nil, sigHashes, prevTxOutValue, outputFetcher)
		if err != nil {
			fmt.Printf("Error creating script engine for input %d: %v\n", i, err)
			return err
		}

		if err := vm.Execute(); err != nil {
			fmt.Printf("Invalid signature for input %d: %v\n", i, err)
			return err
		} else {
			fmt.Printf("Valid signature for input %d\n", i)
		}
	}
	fmt.Println("Transaction successfully verified")
	return nil
}

func buildEmptyRevealTx(receiver btcutil.Address, inscriptionScript []byte, revealOutValue, feeRate int64, opReturnData []byte) (
	*wire.MsgTx, int64, error) {
	totalPrevOutput := int64(0)
	tx := wire.NewMsgTx(wire.TxVersion)
	// add 1 txin
	in := wire.NewTxIn(&wire.OutPoint{Index: uint32(0)}, nil, nil)
	in.Sequence = defaultSequenceNum
	tx.AddTxIn(in)
	if len(opReturnData) > 0 {
		tx.AddTxOut(wire.NewTxOut(0, opReturnData))
	}
	// add 1 txout
	scriptPubKey, err := txscript.PayToAddrScript(receiver)
	if err != nil {
		return nil, 0, err
	}
	out := wire.NewTxOut(revealOutValue, scriptPubKey)
	tx.AddTxOut(out)
	// calculate total prev output
	revealBaseTxFee := int64(tx.SerializeSize()) * feeRate
	totalPrevOutput += revealOutValue + revealBaseTxFee
	// add witness
	emptySignature := make([]byte, 64)
	emptyControlBlockWitness := make([]byte, 33)
	// calculate total prev output
	fee := (int64(wire.TxWitness{emptySignature, inscriptionScript, emptyControlBlockWitness}.SerializeSize()+2+3) / 4) * feeRate
	totalPrevOutput += fee

	return tx, totalPrevOutput, nil
}
func findBestUtxo(commitTxOutPointList []*Utxo, totalRevealPrevOutput, commitFeeRate int64) []*Utxo {
	sort.Slice(commitTxOutPointList, func(i, j int) bool {
		return commitTxOutPointList[i].Value > commitTxOutPointList[j].Value
	})
	best := make([]*Utxo, 0)
	total := int64(0)
	for _, utxo := range commitTxOutPointList {
		if total >= totalRevealPrevOutput+commitFeeRate {
			break
		}
		best = append(best, utxo)
		total += utxo.Value
	}
	return best
}

func buildCommitTx(commitTxOutPointList []*Utxo, revealTxPrevOutput *wire.TxOut, commitFeeRate int64, runeData []byte, splitChangeOutput bool) (*wire.MsgTx, error) {
	totalSenderAmount := btcutil.Amount(0)
	totalRevealPrevOutput := revealTxPrevOutput.Value
	tx := wire.NewMsgTx(wire.TxVersion)
	var changePkScript *[]byte
	bestUtxo := findBestUtxo(commitTxOutPointList, totalRevealPrevOutput, commitFeeRate)
	for _, utxo := range bestUtxo {
		txOut := utxo.TxOut()
		outPoint := utxo.OutPoint()
		if changePkScript == nil { // first sender as change address
			changePkScript = &txOut.PkScript
		}
		in := wire.NewTxIn(&outPoint, nil, nil)
		in.Sequence = defaultSequenceNum
		tx.AddTxIn(in)
		totalSenderAmount += btcutil.Amount(txOut.Value)
	}
	if len(runeData) > 0 {

		tx.AddTxOut(wire.NewTxOut(0, runeData))

	}
	// add reveal tx output
	tx.AddTxOut(revealTxPrevOutput)
	if splitChangeOutput || !bytes.Equal(*changePkScript, revealTxPrevOutput.PkScript) {
		// add change output
		tx.AddTxOut(wire.NewTxOut(0, *changePkScript))
	}
	//mock witness to calculate fee
	emptySignature := make([]byte, 64)
	for _, in := range tx.TxIn {
		in.Witness = wire.TxWitness{emptySignature}
	}
	fee := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(tx))) * btcutil.Amount(commitFeeRate)
	changeAmount := totalSenderAmount - btcutil.Amount(totalRevealPrevOutput) - fee
	if changeAmount > 0 {
		tx.TxOut[len(tx.TxOut)-1].Value += int64(changeAmount)
	} else {
		tx.TxOut = tx.TxOut[:len(tx.TxOut)-1]
		if changeAmount < 0 {
			feeWithoutChange := btcutil.Amount(mempool.GetTxVirtualSize(btcutil.NewTx(tx))) * btcutil.Amount(commitFeeRate)
			if totalSenderAmount-btcutil.Amount(totalRevealPrevOutput)-feeWithoutChange < 0 {
				return nil, errors.New("insufficient balance")
			}
		}
	}
	//clear mock witness
	for _, in := range tx.TxIn {
		in.Witness = nil
	}
	return tx, nil
}

func completeRevealTx(privateKey *btcec.PrivateKey, commitTx *wire.MsgTx, revealTx *wire.MsgTx, inscriptionScript []byte) (*wire.MsgTx, error) {
	//set commit tx hash to reveal tx input
	revealTx.TxIn[0].PreviousOutPoint.Hash = commitTx.TxHash()
	// witness[0]. sign commit tx
	revealTxPrevOutputFetcher := txscript.NewCannedPrevOutputFetcher(commitTx.TxOut[0].PkScript, commitTx.TxOut[0].Value)
	tsHash, err := txscript.CalcTapscriptSignaturehash(txscript.NewTxSigHashes(revealTx, revealTxPrevOutputFetcher),
		txscript.SigHashDefault, revealTx, 0, revealTxPrevOutputFetcher, txscript.NewBaseTapLeaf(inscriptionScript))
	if err != nil {
		return nil, err
	}
	signature, err := schnorr.Sign(privateKey, tsHash)
	if err != nil {
		return nil, err
	}
	//witness[2]. build control block
	leafNode := txscript.NewBaseTapLeaf(inscriptionScript)
	proof := &txscript.TapscriptProof{
		TapLeaf:  leafNode,
		RootNode: leafNode,
	}
	controlBlock := proof.ToControlBlock(privateKey.PubKey())
	controlBlockWitness, err := controlBlock.ToBytes()
	if err != nil {
		return nil, err
	}
	// 3. set full witness
	revealTx.TxIn[0].Witness = wire.TxWitness{signature.Serialize(), inscriptionScript, controlBlockWitness}

	// check tx max tx weight

	revealWeight := blockchain.GetTransactionWeight(btcutil.NewTx(revealTx))
	if revealWeight > MaxStandardTxWeight {
		return nil, errors.New(fmt.Sprintf("reveal(index %d) transaction weight greater than %d (MAX_STANDARD_TX_WEIGHT): %d", 0, MaxStandardTxWeight, revealWeight))
	}

	return revealTx, nil
}

func signCommitTx(prvKey *btcec.PrivateKey, utxos []*Utxo, commitTx *wire.MsgTx) (*wire.MsgTx, error) {
	// build utxoList for FetchPrevOutput
	utxoList := UtxoList(utxos)
	for i, txIn := range commitTx.TxIn {
		txOut := utxoList.FetchPrevOutput(commitTx.TxIn[i].PreviousOutPoint)
		witness, err := txscript.TaprootWitnessSignature(commitTx, txscript.NewTxSigHashes(commitTx, utxoList),
			i, txOut.Value, txOut.PkScript, txscript.SigHashDefault, prvKey)
		if err != nil {
			return nil, err
		}
		txIn.Witness = witness
	}

	return commitTx, nil
}
func serializeTx(tx *wire.MsgTx) ([]byte, error) {
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
