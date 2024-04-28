package main

import (
	"os"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var lang = getDefaultLanguage()

func init() {
	initString("Please select an option", "请选择一个选项")
	initString("Etching a new rune", "发行新的符文")
	initString("Mint rune", "挖掘已定义的符文")
	initString("Prompt failed %v", "提示错误：%v")
	initString("Fatal error config file: %s", "config文件读取错误：%s")
	initString("Unable to unmarshal config: %s", "config文件解码错误：%s")
	initString("Private key error:", "私钥配置错误：")
	initString("Your address is: ", "您的地址是：")
	initString("Etching rune encipher error:", "发行符文配置有误")
	initString("Etching:%s, data:%x", "符文配置:%s, 编码后数据:%x")
	initString("BuildRuneEtchingTxs error:", "发行符文交易构建错误")
	initString("commit Tx: %x\n", "提交交易: %x\n")
	initString("reveal Tx: %x\n", "揭示交易: %x\n")
	initString("SendTx", "发送交易")
	initString("WriteTxToFile", "写入交易到文件")
	initString("How to process the transaction?", "如何处理交易？")
	initString("SendRawTransaction error:", "发送原始交易错误：")
	initString("committed tx hash:", "已提交，交易哈希：")
	initString("waiting for confirmations..., please don't close the program.", "等待确认中，确认数必须大于6之后才能发送揭示交易。请勿关闭程序。")
	initString("GetTransaction error:", "获取交易错误：")
	initString("commit tx confirmations:", "提交交易确认数：")
	initString("Etch complete, reveal tx hash:", "发行完成，揭示交易哈希：")
	initString("create file tx.txt error:", "创建交易文件tx.txt错误：")
	initString("write to file tx.txt", "写入交易到文件tx.txt")
	initString("Mint Rune[%s] data: 0x%x\n", "挖掘符文[%s] 数据: 0x%x\n")
	initString("BuildMintRuneTx error:", "构建挖掘符文交易错误：")
	initString("mint rune tx: %x\n", "挖掘符文交易: %x\n")
}
func initString(english, chinese string) {
	key := english
	message.SetString(language.English, key, english)
	message.SetString(language.Chinese, key, chinese)
}
func i18n(key string) string {
	str := message.NewPrinter(lang).Sprintf(key)
	if len(str) == 0 {
		return key
	}
	return str
}

func getDefaultLanguage() language.Tag {
	langEnv := os.Getenv("LANG")
	if langEnv == "" {
		langEnv = os.Getenv("LANGUAGE")
	}
	langTag := strings.Split(langEnv, ".")[0]
	tag, err := language.Parse(langTag)
	if err != nil {
		return language.Chinese
	}
	return tag
}
