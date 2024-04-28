package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/pkg/errors"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type MempoolConnector struct {
	baseUrl string
	network *chaincfg.Params
}

func NewMempoolConnector(config Config) *MempoolConnector {
	baseURL := config.RpcUrl
	//net := config.Network
	//if net == "mainet" {
	//	baseURL = "https://mempool.space/api"
	//} else if net == "testnet" {
	//	//baseURL = "https://mempool.space/testnet/api"
	//	baseURL = "https://blockstream.info/testnet/api"
	//} else if net == "signet" {
	//	baseURL = "https://mempool.space/signet/api"
	//} else {
	//	log.Fatal("mempool don't support other netParams")
	//}
	connector := &MempoolConnector{
		baseUrl: baseURL,
		network: config.GetNetwork(),
	}
	return connector
}

func (m MempoolConnector) GetBlockHeight() (uint64, error) {
	res, err := m.request(http.MethodGet, "/blocks/tip/height", nil)
	if err != nil {
		return 0, err
	}
	var height uint64
	err = json.Unmarshal(res, &height)
	if err != nil {
		return 0, err
	}
	log.Printf("found latest block height %d", height)
	return height, nil

}

func (m MempoolConnector) GetBlockHashByHeight(height uint64) ([]byte, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/block-height/%d", height), nil)
	if err != nil {
		return nil, err
	}
	hashString := string(res)
	hash, err := hex.DecodeString(hashString)
	log.Printf("found block hash %x by height:%d", hash, height)
	return hash, nil
}

func (m MempoolConnector) GetBlockByHash(blockHash Hash) (*wire.MsgBlock, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/block/%s/raw", blockHash), nil)
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	block := &wire.MsgBlock{}
	if err := block.Deserialize(bytes.NewReader(res)); err != nil {
		return nil, err
	}
	log.Printf("found block %s", blockHash)
	return block, nil
}

func (m MempoolConnector) GetHeaderByHash(h Hash) (*wire.BlockHeader, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/block/%s/header", h), nil)
	if err != nil {
		return nil, err
	}
	headerBytes, err := hex.DecodeString(string(res))
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	header := &wire.BlockHeader{}
	if err := header.Deserialize(bytes.NewReader(headerBytes)); err != nil {
		return nil, err
	}
	log.Printf("found header %s", h)
	return header, nil
}

func (m MempoolConnector) GetBlockByHeight(height uint64) (*wire.MsgBlock, error) {
	hash, err := m.GetBlockHashByHeight(height)
	if err != nil {
		return nil, err
	}
	return m.GetBlockByHash(Hash(hash))
}

func (m MempoolConnector) GetBlockTxIDS(bh Hash) ([]Hash, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/block/%s/txids", bh), nil)
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	var txids []string
	err = json.Unmarshal(res, &txids)
	if err != nil {
		return nil, err
	}
	hashes := make([]Hash, len(txids))
	for i, txid := range txids {
		hashes[i] = HexToHash(txid)
	}
	log.Printf("found %d txids for block %s", len(hashes), bh)
	return hashes, nil
}
func (m MempoolConnector) SendRawTransaction(tx *wire.MsgTx, allowHighFees bool) (*chainhash.Hash, error) {
	log.Printf("send tx %s to bitcoin network", tx.TxHash())
	//th := tx.TxHash()
	//return &th, nil

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return nil, err
	}

	res, err := m.request(http.MethodPost, "/tx", strings.NewReader(hex.EncodeToString(buf.Bytes())))
	if err != nil {
		return nil, err
	}

	txHash, err := chainhash.NewHashFromStr(string(res))
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to parse tx hash, %s", string(res)))
	}
	return txHash, nil
}

func (m MempoolConnector) GetUtxos(address string) ([]*Utxo, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/address/%s/utxo", address), nil)
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	var mutxos []mempoolUTXO
	err = json.Unmarshal(res, &mutxos)
	if err != nil {
		return nil, err
	}
	addr, err := btcutil.DecodeAddress(address, m.network)
	if err != nil {
		return nil, err
	}
	pkScript, _ := txscript.PayToAddrScript(addr)
	utxos := make([]*Utxo, len(mutxos))
	for i, mutxo := range mutxos {
		txHash, err := chainhash.NewHashFromStr(mutxo.Txid)
		if err != nil {
			return nil, err
		}
		utxos[i] = &Utxo{
			TxHash:   BytesToHash(txHash.CloneBytes()),
			Index:    uint32(mutxo.Vout),
			Value:    mutxo.Value,
			PkScript: pkScript,
		}
	}
	log.Printf("found %d unspent outputs for address %s", len(utxos), address)
	return utxos, nil
}

type txStatus struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight uint64 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   int64  `json:"block_time"`
}
type txResponse struct {
	Txid     string   `json:"txid"`
	Version  int      `json:"version"`
	Locktime int      `json:"locktime"`
	Size     int      `json:"size"`
	Fee      int      `json:"fee"`
	Status   txStatus `json:"status"`
}

func (m MempoolConnector) GetTxByHash(hash string) (*BtcTxInfo, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/tx/%s", hash), nil)
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	var resp txResponse
	err = json.Unmarshal(res, &resp)
	if err != nil {
		return nil, err
	}
	txInfo := &BtcTxInfo{
		Tx:            nil,
		BlockHeight:   resp.Status.BlockHeight,
		BlockHash:     HexToHash(resp.Status.BlockHash),
		BlockTime:     uint64(resp.Status.BlockTime),
		Confirmations: 0,
		TxIndex:       0,
	}
	tx, err := m.GetRawTxByHash(hash)
	if err != nil {
		return nil, err
	}
	txInfo.Tx = tx
	return txInfo, nil
}

func (m MempoolConnector) GetRawTxByHash(hash string) (*wire.MsgTx, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/tx/%s/raw", hash), nil)
	if err != nil {
		return nil, err
	}
	//unmarshal the response
	tx := &wire.MsgTx{}
	if err := tx.Deserialize(bytes.NewReader(res)); err != nil {
		return nil, err
	}
	log.Printf("found tx %s", hash)
	return tx, nil
}

type mempoolUTXO struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Status struct {
		Confirmed   bool   `json:"confirmed"`
		BlockHeight int    `json:"block_height"`
		BlockHash   string `json:"block_hash"`
		BlockTime   int64  `json:"block_time"`
	} `json:"status"`
	Value int64 `json:"value"`
}

func (m MempoolConnector) request(method, subPath string, requestBody io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s", m.baseUrl, subPath)

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	return body, nil
}

func (m MempoolConnector) GetBalance(address string) (uint64, error) {
	res, err := m.request(http.MethodGet, fmt.Sprintf("/address/%s", address), nil)
	if err != nil {
		return 0, err
	}
	//unmarshal the response
	var balance struct {
		ChainStats struct {
			FundedTxoCount uint64 `json:"funded_txo_count"`
			FundedTxoSum   uint64 `json:"funded_txo_sum"`
			SpentTxoCount  uint64 `json:"spent_txo_count"`
			SpentTxoSum    uint64 `json:"spent_txo_sum"`
		} `json:"chain_stats"`
	}
	err = json.Unmarshal(res, &balance)
	if err != nil {
		return 0, err
	}
	log.Printf("found balance %d for address %s", balance.ChainStats.FundedTxoSum-balance.ChainStats.SpentTxoSum, address)
	return balance.ChainStats.FundedTxoSum - balance.ChainStats.SpentTxoSum, nil
}

type BtcTxInfo struct {
	Tx            *wire.MsgTx
	BlockHeight   uint64
	BlockHash     Hash
	BlockTime     uint64
	Confirmations uint64
	TxIndex       uint64
}
