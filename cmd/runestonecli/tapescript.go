package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

func CreateInscriptionScript(pk *btcec.PublicKey, contentType string, fileBytes []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	//push pubkey
	pk32 := schnorr.SerializePubKey(pk)
	log.Printf("put pubkey:%x to tapScript", pk32)
	builder.AddData(pk32)
	builder.AddOp(txscript.OP_CHECKSIG)
	//Ordinals script
	builder.AddOp(txscript.OP_FALSE)
	builder.AddOp(txscript.OP_IF)
	builder.AddData([]byte("ord"))
	builder.AddOp(txscript.OP_DATA_1) //??
	builder.AddOp(txscript.OP_DATA_1) //??
	builder.AddData([]byte(contentType))
	builder.AddOp(txscript.OP_0)
	data, err := builder.Script()
	if err != nil {
		return nil, err
	}
	splitLen := 520
	point := 0
	for {
		builder = txscript.NewScriptBuilder()
		if point+splitLen > len(fileBytes) {
			builder.AddData(fileBytes[point:])
			data1, err1 := builder.Script()
			if err1 != nil {
				return nil, err1
			}
			data = append(data, data1...)
			break
		}
		builder.AddData(fileBytes[point : point+splitLen])
		data1, err1 := builder.Script()
		if err1 != nil {
			return nil, err1
		}
		data = append(data, data1...)
		point += splitLen
	}
	data = append(data, txscript.OP_ENDIF)
	return data, err
}

func CreateCommitmentScript(pk *btcec.PublicKey, commitment []byte) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	//push pubkey
	pk32 := schnorr.SerializePubKey(pk)

	builder.AddData(pk32)
	builder.AddOp(txscript.OP_CHECKSIG)
	//Commitment script
	builder.AddOp(txscript.OP_FALSE)
	builder.AddOp(txscript.OP_IF)
	builder.AddData(commitment)
	builder.AddOp(txscript.OP_ENDIF)
	return builder.Script()
}
func IsTapScript(witness wire.TxWitness) bool {
	if len(witness) != 3 {
		return false
	}
	witness2 := witness[2]
	if len(witness2) == 33 && (witness2[0] == 0xc0 || witness2[0] == 0xc1) {
		return true
	}
	return false
}

func GetOrdinalsContent(tapScript []byte) (mime string, content []byte, err error) {
	scriptStr, err := txscript.DisasmString(tapScript)
	if err != nil {
		return "", nil, err
	}
	start := false
	scriptStrArray := strings.Split(scriptStr, " ")
	contentHex := ""
	for i := 0; i < len(scriptStrArray); i++ {
		if scriptStrArray[i] == "6f7264" { // 6f7264 ==ord
			start = true
			mimeBytes, _ := hex.DecodeString(scriptStrArray[i+2])
			mime = string(mimeBytes)
			i = i + 4
		}
		if i < len(scriptStrArray) {
			if start {
				contentHex = contentHex + scriptStrArray[i]
			}
			if scriptStrArray[i] == "OP_ENDIF" {
				break
			}
		}
	}
	contentBytes, _ := hex.DecodeString(contentHex)
	return mime, contentBytes, nil
}

var ordiBytes []byte

func init() {
	builder := txscript.NewScriptBuilder()
	builder.AddOp(txscript.OP_FALSE)
	builder.AddOp(txscript.OP_IF)
	builder.AddData([]byte("ord"))
	builder.AddOp(txscript.OP_DATA_1)
	ordiBytes, _ = builder.Script()
}
func IsOrdinalsScript(script []byte) bool {
	if bytes.Contains(script, ordiBytes) && script[len(script)-1] == txscript.OP_ENDIF {
		return true
	}
	return false
}

func GetInscriptionContent(tx *wire.MsgTx) (contentType string, content []byte, err error) {
	for _, txIn := range tx.TxIn {
		if IsTapScript(txIn.Witness) {
			if IsOrdinalsScript(txIn.Witness[1]) {
				contentType, data, err := GetOrdinalsContent(txIn.Witness[1])
				if err != nil {
					return "", nil, err
				}
				return contentType, data, nil
			}
		}
	}
	return "", nil, errors.New("no ordinals script found")
}
