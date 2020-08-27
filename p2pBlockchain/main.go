package p2pBlockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	log2 "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/multiformats/go-multiaddr"
	"github.com/whyrusleeping/go-logging"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)


type Block struct {
	BlockNumber  int			// 高度
	Timestamp string		// 时间戳
	Info int				// 交易信息
	PrevHash string     // 上一个哈希值
	HashCode string		// 当前的哈希值
}

var Blockchain []Block  //定义一个区块链

var mutex = &sync.Mutex{} //防止同一时间产生多个区块

var announcements = make(chan string) //出块结束后向所有节点进行广播

func generateBlock(oldBlock Block, info int, address string) (Block) {
	newBlock := &Block {
		BlockNumber : oldBlock.BlockNumber + 1,
		Timestamp : time.Now().Format("20060102150405"),
		Info	: info,
		PrevHash : oldBlock.HashCode,
		Difficulty : DIFFICULTY,
	}
	newBlockHash := calculateBlockHash(*newBlock)
	//产生Nonce
	for i := 0; ;i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(newBlockHash, (*newBlock).Difficulty) {
			fmt.Println(newBlockHash, "do work!")
		} else {
			fmt.Println(newBlockHash, "work done")
			(*newBlock).HashCode = newBlockHash
			break
		}

	}
	return *newBlock
}

//对一个Block进行hash  将一个block的所有字段连接到一起后 再转换为一个hash值
func  calculateBlockHash(block Block) string {
	return calculateHash(fmt.Sprintf("%v%v%v%v",
		block.Index, block.Timestamp, block.BPM, block.PrevHash, block.Nonce))
}

func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 验证hash是否有效
//新区块中的preHash存储上一个区块的Hash
//for 循环通过循环改变Nonce,然后选出符合相应难度系数的Nonce
//isHashValid 判断hash是否满足当前的难度系数，如果难度系数为2，则当前hash的前缀有2个0
func isHashValid(hash string, difficulty int) bool {
	var b = len(hash)
	var end int
	for i := 0; i < b; i++ {
		if hash[i] != '0' {
			end = i
			break
		}
	}
	return  end >= difficulty
}

// 验证区块内容
func isBlockValid(newBlock, oldBlock Block) bool {
	if newBlock.Index == oldBlock.Index + 1 && newBlock.PrevHash == oldBlock.HashCode && calculateBlockHash(newBlock) == newBlock.HashCode {
		return true
	}
	return false
}






// secio  是否对数据流加密；
// privatekey 保证host的安全
// options 构造我们的host地址，以便其他节点链接
func makeBasicHost(secio bool, listenPort int) (host.Host, error) {
	//生成privatekey
	privatekey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		fmt.Println("Failed to generateKey")
		return nil, err
	}

	//创建函数处理集
	options := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(privatekey),
	}

	if !secio {
		options = append(options, libp2p.NoSecurity)
	}

	basicHost, err := libp2p.New(context.Background(), options...)
	if err != nil {
		return nil, err
	}
	//创建主机多地址
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	//通过封装两个地址来建立一个多地址到达主机
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Println("I'm ", fullAddr)
	return basicHost, nil
}

func streamHandle(s network.Stream) {
	log.Println("Got a new stream")
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)
}

func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')
		if str == "" {
			return
		}
		if str != "\n" {
			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "")
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()
		}
	}()
	stReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">")
		sendData, err := stReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		sendData = strings.Replace(sendData, "\n", "", -1)
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)

		if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
			mutex.Lock()
			Blockchain = append(Blockchain, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(Blockchain)
		if err != nil {
			log.Println(err)
		}

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}
}

func main() {
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), ""}

	Blockchain = append(Blockchain, genesisBlock)

	log2.SetAllLoggers(logging.INFO)
	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections") //打开指定的接口
	target := flag.String("d", "", "target peer to dial") //指定想要连接的地址

	secio := flag.Bool("secio", false, "enable secio")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	host, err := makeBasicHost(*secio, *listenF)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		host.SetStreamHandler("/p2p/1.0.0", streamHandle)
		select {} //阻塞程序
	} else {
		host.SetStreamHandler("/p2p/1.0.0", streamHandle)
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}
		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

		log.Println("opening stream")

		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}

		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		go writeData(rw)
		go readData(rw)

		select {} // hang forever

	}
}
