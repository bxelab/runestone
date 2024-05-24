package main

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/bxelab/runestone"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
	"golang.org/x/text/message"
)

var config = DefaultConfig()
var p *message.Printer

func main() {
	p = message.NewPrinter(lang)
	loadConfig()
	checkAndPrintConfig()

	// 显示多语言文本
	items := []string{i18n("Etching a new rune"), i18n("Mint rune"), i18n("Count Rune mint")}
	prompt := promptui.Select{
		Label: i18n("Please select an option"),
		Items: items,
	}

	optionIdx, _, err := prompt.Run()

	if err != nil {
		p.Printf("Prompt failed %v", err)
		return
	}
	if optionIdx == 0 { //Etching a new rune
		BuildEtchingTxs()
	}
	if optionIdx == 1 { //Mint rune
		BuildMintTxs()
	}
	if optionIdx == 2 {
		CountMintRunes()
	}
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(p.Sprintf("Fatal error config file: %s", err))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		panic(p.Sprintf("Unable to unmarshal config: %s", err))
	}
}
func checkAndPrintConfig() {
	//check privatekey and print address
	_, addr, err := config.GetPrivateKeyAddr()
	if err != nil {
		p.Println("Private key error:", err.Error())
		return
	}
	p.Println("Your address is: ", addr)

}
func BuildEtchingTxs() {
	etching, err := config.GetEtching()
	if err != nil {
		p.Println("error:", err.Error())
		return
	}
	rs := runestone.Runestone{Etching: etching}
	data, err := rs.Encipher()
	if err != nil {
		p.Println("Etching rune encipher error:", err.Error())
		return
	}
	etchJson, _ := json.Marshal(etching)
	p.Printf("Etching:%s, data:%x", string(etchJson), data)
	commitment := etching.Rune.Commitment()
	btcConnector := NewMempoolConnector(config)
	prvKey, address, _ := config.GetPrivateKeyAddr()
	utxos, err := btcConnector.GetUtxos(address)
	var cTx, rTx []byte
	mime, logoData := config.GetRuneLogo()
	if len(mime) == 0 {
		cTx, rTx, err = BuildRuneEtchingTxs(prvKey, utxos, data, commitment, config.GetFeePerByte(), config.GetUtxoAmount(), config.GetNetwork(), address)
	} else {
		cTx, rTx, err = BuildInscriptionTxs(prvKey, utxos, mime, logoData, config.GetFeePerByte(), config.GetUtxoAmount(), config.GetNetwork(), commitment, data)
	}
	if err != nil {
		p.Println("BuildRuneEtchingTxs error:", err.Error())
		return
	}
	p.Printf("commit Tx: %x\n", cTx)
	p.Printf("reveal Tx: %x\n", rTx)
	items := []string{i18n("SendTx"), i18n("WriteTxToFile")}
	prompt := promptui.Select{
		Label: i18n("How to process the transaction?"),
		Items: items,
	}

	optionIdx, _, err := prompt.Run()

	if err != nil {
		p.Printf("Prompt failed %v", err)
		return
	}
	if optionIdx == 0 { //Direct send
		SendTx(btcConnector, cTx, rTx)
	}
	if optionIdx == 1 { //write to file
		WriteFile(string(etchJson), cTx, rTx)
	}
}

func SendTx(connector *MempoolConnector, ctx []byte, rtx []byte) {
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.Deserialize(bytes.NewReader(ctx))
	ctxHash, err := connector.SendRawTransaction(tx, false)
	if err != nil {
		p.Println("SendRawTransaction error:", err.Error())
		return
	}
	p.Println("committed tx hash:", ctxHash)
	if rtx == nil {

		return
	}
	p.Println("waiting for confirmations..., please don't close the program.")
	//wail ctx tx confirm
	lock.Lock()
	go func(ctxHash *chainhash.Hash) {
		for {
			time.Sleep(30 * time.Second)
			txInfo, err := connector.GetTxByHash(ctxHash.String())
			if err != nil {
				p.Println("GetTransaction error:", err.Error())
				continue
			}
			p.Println("commit tx confirmations:", txInfo.Confirmations)
			if txInfo.Confirmations > runestone.COMMIT_CONFIRMATIONS {
				break
			}
		}
		lock.Unlock()
	}(ctxHash)
	lock.Lock() //wait
	tx = wire.NewMsgTx(wire.TxVersion)
	tx.Deserialize(bytes.NewReader(rtx))
	rtxHash, err := connector.SendRawTransaction(tx, false)
	if err != nil {
		p.Println("SendRawTransaction error:", err.Error())
		return
	}
	p.Println("Etch complete, reveal tx hash:", rtxHash)
}

var lock sync.Mutex

func WriteFile(etching string, tx []byte, tx2 []byte) {
	//write to file
	file, err := os.OpenFile("tx.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		p.Println("create file tx.txt error:", err.Error())
		return
	}
	defer file.Close()
	file.WriteString(time.Now().String())
	file.WriteString("Etching: " + etching)
	file.WriteString("\n")
	file.WriteString("Commit Tx: " + p.Sprintf("%x", tx))
	file.WriteString("\n")
	if tx2 != nil {
		file.WriteString("Reveal Tx: " + p.Sprintf("%x", tx2))
		file.WriteString("\n")
	}
	p.Println("write to file tx.txt")
}

func BuildMintTxs() {
	runeId, err := config.GetMint()
	if err != nil {
		p.Println(err.Error())
		return

	}
	r := runestone.Runestone{Mint: runeId}
	runeData, err := r.Encipher()
	if err != nil {
		p.Println(err)
	}
	p.Printf("Mint Rune[%s] data: 0x%x\n", config.Mint.RuneId, runeData)
	//dataString, _ := txscript.DisasmString(data)
	//p.Printf("Mint Script: %s\n", dataString)
	btcConnector := NewMempoolConnector(config)
	prvKey, address, _ := config.GetPrivateKeyAddr()
	utxos, err := btcConnector.GetUtxos(address)
	tx, err := BuildTransferBTCTx(prvKey, utxos, address, config.GetUtxoAmount(), config.GetFeePerByte(), config.GetNetwork(), runeData)
	if err != nil {
		p.Println("BuildMintRuneTx error:", err.Error())
		return
	}
	p.Printf("mint rune tx: %x\n", tx)
	items := []string{i18n("SendTx"), i18n("WriteTxToFile")}
	prompt := promptui.Select{
		Label: i18n("How to process the transaction?"),
		Items: items,
	}

	optionIdx, _, err := prompt.Run()

	if err != nil {
		p.Printf("Prompt failed %v", err)
		return
	}
	if optionIdx == 0 { //Direct send
		SendTx(btcConnector, tx, nil)
	}
	if optionIdx == 1 { //write to file
		WriteFile(p.Sprintf("Mint rune[%s]", runeId.String()), tx, nil)
	}
}
