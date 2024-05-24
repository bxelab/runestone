package main

import (
	"fmt"

	"github.com/btcsuite/btcd/txscript"
	"github.com/bxelab/runestone"
)

func CountMintRunes() error {
	btcConnector := NewMempoolConnector(config)

	height, err := btcConnector.GetBlockHeight()
	if err != nil {
		return err
	}
	fmt.Println("Current block height:", height)
	block, err := btcConnector.GetBlockByHeight(height)
	if err != nil {
		return err
	}
	count := map[string]int{}
	for _, tx := range block.Transactions {
		for _, out := range tx.TxOut {
			if out.PkScript[0] == txscript.OP_RETURN && out.PkScript[1] == runestone.MAGIC_NUMBER {
				r := &runestone.Runestone{}
				artifact, err := r.Decipher(tx)
				if err != nil {
					fmt.Println(err)
					return err
				}
				//a, _ := json.Marshal(artifact)
				//fmt.Println(string(a))
				if artifact.Runestone != nil && artifact.Runestone.Mint != nil {
					runeId := artifact.Runestone.Mint.String()
					if count[runeId] == 0 {
						count[runeId] = 1
					} else {
						count[runeId]++

					}
				}
			}
		}
	}
	//print count
	for k, v := range count {
		fmt.Println("RuneId:", k, "Count:", v)
	}

	return nil
}
