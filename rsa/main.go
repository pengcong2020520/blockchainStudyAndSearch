package main

import (
	"ConsensusAlgorithm/rsa/cli"
	"ConsensusAlgorithm/rsa/hd"
	"fmt"
	"log"
)

func main() {
	//cli.NewMnemonic()
	mnemonic := cli.NewMnemonic()
	//2. 通过助记词生成钱包     助记词--> 种子 --> 钱包
	wallet, err := hd.NewFromMnemonic(mnemonic, "password")
	if err != nil {
		log.Panic("Failed to NewFromMnemonic", err)
	}
	//3. 通过钱包推导账户
	path, _ := hd.ParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, true)
	if err != nil {
		log.Panic("Failed to Derive", err)
	}
	fmt.Println(account.Address.Hex())
}
