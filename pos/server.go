package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	_"github.com/joho/godotenv"
)

// 我们将有一个中心化的TCP服务器节点，其他的节点可以连接到这个服务器上
// 最新的区块链的状态将定期的广播到每个节点上
// 每个节点都可以公平的提议建立新的区块
// 基于每个节点的“资产”的数量，选举出一个获胜者，并将获胜者（区块）添加到区块链当中去

type Block struct {
	BlockNumber  int			// 高度
	Timestamp string		// 时间戳
	Info int				// 交易信息
	prevHash string     // 上一个哈希值
	HashCode string		// 当前的哈希值
	Validator string    // 区块验证者  其中POW为difficulty
}

var Blockchain []Block  //定义一个区块链

var tempBlocks []Block // 临时存储单元 在区块被选出来之前 ，临时存储在这里 该单元最多包含多少个区块

var candidateBlocks = make(chan Block) // 临时候选通道， 任何一个节点在提出一个新块时 都会把他先发到这个通道里
			//最终会从该候选通道中选出一个区块

var announcements = make(chan string) // 也是一个通道 主GO TCP服务器将向所有节点广播最新的区块链

var mutex = &sync.Mutex{}	//防止同一时间产生多个区块

var validators = make(map[string]int)  //节点的临时存储map 同时也会保存每个节点质押的token数

//生成区块函数
	//由旧区块 、 新的BPM结构 、 验证者  产生新的区块
func generateBlock(oldBlock Block, BPM int, address string) (Block) {
	newBlock := &Block {
		Index : oldBlock.Index + 1,
		Timestamp : time.Now().Format("20060102150405"),
		BPM	: BPM,
		prevHash : oldBlock.HashCode,
		Validator : address,
	}
	newBlockHash := newBlock.calculateBlockHash()
	newBlock.HashCode = newBlockHash
	return *newBlock
}

	//对一个Block进行hash  将一个block的所有字段连接到一起后 再转换为一个hash值
func (block Block) calculateBlockHash() string {
	return calculateHash(fmt.Sprintf("%v%v%v%v",
		block.Index, block.Timestamp, block.BPM, block.prevHash, block.Validator))
}

func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 验证区块内容
func isBlockValid(newBlock, oldBlock Block) bool {
	if newBlock.Index == oldBlock.Index + 1 && newBlock.prevHash == oldBlock.HashCode && newBlock.calculateBlockHash() == newBlock.HashCode {
		return true
	}
	return false
}

//验证者
func handleConn(conn net.Conn) {
	defer conn.Close()
	go func(){
		for {
			msg := <- announcements
			io.WriteString(conn, msg)
		}
	}()

	var address string

	//验证者输入所需要质押的tokens， tokens越大 越容易获得新区块的记账权
	//conn是个接口  他的write从连接中读取数据
	io.WriteString(conn, "Enter token balance:")

	//允许验证者输入他持有的令牌数量，然后，该验证者被分配一个 SHA256地址，
	// 随后该验证者地址和验证者的令牌数被添加到验证者列表validators 中
	scanBalance := bufio.NewScanner(conn)
	//提供文件的token
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Panicf("%v not a numbe : %v", scanBalance.Text(), err)
			return
		}
		address = calculateHash(time.Now().String())
		validators[address] = balance
		fmt.Println((validators))
		break
	}


	io.WriteString(conn, "\n Enter a new BPM:")
	scanBPM := bufio.NewScanner(conn)
	go func(){
		for {
			for scanBPM.Scan() {
				bpm, err := strconv.Atoi(scanBPM.Text())
				if err != nil {
					log.Printf("%v not a number : %v", scanBPM.Text(), err)
					delete(validators, address)
					conn.Close()
				}
				mutex.Lock()
				oldLastIndex := Blockchain[len(Blockchain)-1]  //取最后一个区块
				mutex.Unlock()
				newBlock := generateBlock(oldLastIndex, bpm, address)
				if isBlockValid(newBlock, oldLastIndex) {
					candidateBlocks<- newBlock
				}
				io.WriteString(conn, "\n Enter a new BPM:")
			}
		}
	}()

	for {
		time.Sleep(3 * time.Minute)
		mutex.Lock()
		output, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output) + "\n")
	}
}

//选择获取记账权节点
func pickWinner() {
	time.Sleep(time.Minute)
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()
	lotteryPool := []string{}
	OUTER:
	if len(temp) > 0 {
			for _, block := range temp {
				//如果验证者验证过即不可再验证
				for _, node := range lotteryPool {
					if block.Validator == node {
						goto OUTER
					}
				}
				mutex.Lock()
				setValidators := validators  //validators 每个账户地址对应的token数
				mutex.Unlock()
				k, ok := setValidators[block.Validator] //k为validators的token数
				if ok {
					for i := 0; i < k; i++ {
						lotteryPool = append(lotteryPool, block.Validator)
					}
				}
			}
			//随机选取矿工
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s)
			lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]
			for _, block := range temp {
				if block.Validator == lotteryWinner {
					mutex.Lock()
					Blockchain = append(Blockchain, block)
					mutex.Unlock()
					for _ = range validators {
						announcements<- "\n winning validator: " + lotteryWinner + "\n"
					}
					break
				}
			}
	}
	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}

func main() {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal(err)
	//}
	// 创建初始区块
	genesisBlock := &Block{
		Index :	0,
		Timestamp :	time.Now().Format("20060102150405"),
		BPM :	0,
		prevHash :	"",
		Validator :	"",
	}
	genesisBlock.HashCode = genesisBlock.calculateBlockHash()

	spew.Dump(*genesisBlock)
	Blockchain = append(Blockchain, *genesisBlock)
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP Server Listening on port : 8080")
	defer server.Close()

	go func(){
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner()
		}
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}





