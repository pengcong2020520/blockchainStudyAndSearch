package cli

import (
	"fmt"
	"github.com/tyler-smith/go-bip39"
)

type CLI struct {
	DataPath string
	NetWorkUrl string
	TokenFile string
}

func NewCLI(path, url, tokenfile string) *CLI {
	return &CLI{
		DataPath : path,
		NetWorkUrl : url,
		TokenFile : tokenfile,
	}
}

func (c *CLI) CreateWallet(name, pass string) {
	//1. 生成助记词
}




//生成助记词函数
func NewMnemonic() string{
	//1. NewEntropy 后面必须填上32的整数倍的数，并且在128-256之间
	entropy, _ := bip39.NewEntropy(128)
	//2. 助记词
	mnemonic, _ := bip39.NewMnemonic(entropy)
	//得到助记词。。
	fmt.Println(mnemonic)
	return mnemonic
}