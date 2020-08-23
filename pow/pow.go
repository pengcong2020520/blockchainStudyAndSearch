package pow

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

const DIFFICULTY  = 1000000000


type Block struct {
	Index  int			// 高度
	Timestamp string		// 时间戳
	BPM int				// 交易信息
	PrevHash string     // 上一个哈希值
	HashCode string		// 当前的哈希值
	Difficulty int    // 区块验证者  其中POW为difficulty
	Nonce string    //防止双花
}

var Blockchain []Block  //定义一个区块链

var mutex = &sync.Mutex{}

var announcements = make(chan string)

//生成区块函数
//由旧区块 、 新的BPM结构 、 验证者  产生新的区块
func generateBlock(oldBlock Block, BPM int, address string) (Block) {
	newBlock := &Block {
		Index : oldBlock.Index + 1,
		Timestamp : time.Now().Format("20060102150405"),
		BPM	: BPM,
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

//处理conn
func handleConn(conn net.Conn) {
	defer conn.Close()
	go func(){
		for {
			msg := <- announcements
			io.WriteString(conn, msg)
		}
	}()

	var address string

	io.WriteString(conn, "\n Enter a new BPM:")
	scanBPM := bufio.NewScanner(conn)
	go func(){
		for {
			bpm, err := strconv.Atoi(scanBPM.Text())
			if err != nil {
				log.Printf("%v not a number : %v", scanBPM.Text(), err)
				conn.Close()
			}
			mutex.Lock()
			oldLastIndex := Blockchain[len(Blockchain)-1]  //取最后一个区块
			mutex.Unlock()

			newBlock := generateBlock(oldLastIndex, bpm, address)
			if isBlockValid(newBlock, oldLastIndex) {
				Blockchain = append(Blockchain, newBlock)
			}
			io.WriteString(conn, "\n Enter a new BPM:")
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
		PrevHash :	"",
		Difficulty : DIFFICULTY,
	}
	genesisBlock.HashCode = calculateBlockHash(*genesisBlock)

	spew.Dump(*genesisBlock)
	Blockchain = append(Blockchain, *genesisBlock)
	server, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP Server Listening on port : 8080")
	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}
}





