package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

//  defines derivation paths to be of the form
//
//  m / purpose' / coin_type' / account' / change / address_index
//  BIP-44 purpose == 44  SLIP-44  coin_type == 60
//  The root path for Ethereum is m/44'/60'/0'/0

func GenerateRSA(length int) {
	// 1. 生成RSA私钥
	//GenerateKey generates an RSA keypair of the given bit size using the random source random (for example, crypto/rand.Reader).
	privateKey, err := rsa.GenerateKey(rand.Reader, length)
	//生成一个给定长度length的私钥，rand.Reader为加密包的随机数生成器的全局共享实例
	if err != nil {
		fmt.Println("Failed to generate privateKey !", err)
		return
	}
	//2. 对所产生的私钥进行x509序列化并定制PEM的Block
	x509privateKey := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:    "private key",
		Headers: nil,
		Bytes:   x509privateKey,
	}
	//3. 将私钥pem序列化并写入文件中
	filename := "/ConsensusAlgorithm/rsa/key/RSAprivatekey.txt"
	privatekeyfile, err := os.Create(filename)
	if err != nil {
		fmt.Println("Failed to Create RSAprivatekey file ! ", err)
		return
	}
	defer privatekeyfile.Close()
	err = pem.Encode(privatekeyfile, &privateKeyBlock)  //将block中的信息pem序列化并写入file中
	if err != nil {
		fmt.Println("Failed to pem encode privateKeyBlock")
		return
	}

	// 4. 生成公钥(公钥可由私钥产生)  其他步骤与上述相同
	publicKey := privateKey.PublicKey
	x509publicKey := x509.MarshalPKCS1PublicKey(&publicKey)
	publicKeyBlock := pem.Block{
		Type:    "public key",
		Headers: nil,
		Bytes:   x509publicKey,
	}
	path := "/ConsensusAlgorithm/rsa/key/RSApublickey.txt"
	publicKeyfile, err := os.Create(path)
	if err != nil {
		fmt.Println("Failed to Create RSApublickey file ! ", err)
		return
	}
	defer publicKeyfile.Close()
	err = pem.Encode(publicKeyfile, &publicKeyBlock)
	if err != nil {
		fmt.Println("Failed to pem encode publicKeyBlock")
		return
	}
}

//对数据进行签名
func SignatureRSA(data string) string{
	//1. 获取私钥(打开keystore文件 )
	privateKeyPath := "/ConsensusAlgorithm/rsa/key/RSAprivatekey.txt"
	file, err := os.Open(privateKeyPath)// 得到*File类型
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	// 获取文件长度
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	//定义一个放数据的切片
	buf := make([]byte, fileSize)
	//将file内容读取到buf缓存中
	file.Read(buf)
	//2. 将得到的私钥内容进行pem与x509解码得到私钥明文
	privateKeyBlock, _ := pem.Decode(buf)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		fmt.Println("Failed to parsed private key !")
		return ""
	}
	//3. 用私钥进行签名
	hash256 := sha256.New()
	//将数据进行 json 序列化
	datajson, _ := json.Marshal(data)
	//将json数据进行hash
	hash256.Write(datajson)
	//用私钥进行签名
	signData, _ := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash256.Sum(nil))
	if err != nil {
		fmt.Println("Failed to sign data!", err)
		return ""
	}
	return string(signData)
}

///验证签名
func VerifyRSA(data string, signData string) bool {
	// 1. 获取公钥文件
	publicKeyPath := "/ConsensusAlgorithm/rsa/key/RSApublickey.txt"
	file, err := os.Open(publicKeyPath)// 得到*File类型
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	// 获取文件长度
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	//定义一个放数据的切片
	buf := make([]byte, fileSize)
	//将file内容读取到buf缓存中
	file.Read(buf)
	//2. 将得到的公钥内容进行pem与x509解码得到公钥明文
	publicKeyBlock, _ := pem.Decode(buf)
	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBlock.Bytes)
	if err != nil {
		fmt.Println("Failed to parsed public key !")
		return false
	}
	hash256 := sha256.New()
	jsonData, _ := json.Marshal(data)
	hash256.Write(jsonData)
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hash256.Sum(nil), []byte(signData))
	if err != nil {
		return false
	}
	return true
}
