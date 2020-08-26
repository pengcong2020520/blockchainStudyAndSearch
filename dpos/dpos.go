package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"log"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Block struct {
	Index  int			// 高度
	Timestamp string		// 时间戳
	BPM int				// 交易信息
	PrevHash string     // 上一个哈希值
	HashCode string		// 当前的哈希值
	Delegate string    //  代理人 （超级节点）
}

type SelDelegates struct {
	addresss string
	votes int
}

type SelDels []SelDelegates

const AllDelList = 1000
const SelDelList = 100
const WinDelList = 10

var Blockchain []Block  //定义一个区块链
var announcements = make(chan string) // 也是一个通道 主GO TCP服务器将向所有节点广播最新的区块链
var mutex = &sync.Mutex{}	//防止同一时间产生多个区块


var AllDelegatesList []string   //所有的超级节点

var SelDelegatesList SelDels  //被选中的超级节点
var WinDelegatesList SelDels  //最后代理的超级节点

//生成区块函数
//由旧区块 、 新的BPM结构 、 验证者  产生新的区块
func generateBlock(oldBlock Block, BPM int, address string) (Block) {
	newBlock := &Block {
		Index : oldBlock.Index + 1,
		Timestamp : time.Now().Format("20060102150405"),
		BPM	: BPM,
		PrevHash : oldBlock.HashCode,
		Delegate : address,
	}
	newBlockHash := newBlock.calculateBlockHash()
	newBlock.HashCode = newBlockHash
	return *newBlock
}

//对一个Block进行hash  将一个block的所有字段连接到一起后 再转换为一个hash值
func (block Block) calculateBlockHash() string {
	return calculateHash(fmt.Sprintf("%v%v%v%v",
		block.Index, block.Timestamp, block.BPM, block.PrevHash, block.Delegate))
}

func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// 验证区块内容
func isBlockValid(newBlock, oldBlock Block) bool {
	if newBlock.Index == oldBlock.Index + 1 && newBlock.PrevHash == oldBlock.HashCode && newBlock.calculateBlockHash() == newBlock.HashCode {
		return true
	}
	return false
}

//挑选超级节点（例如从1000个超级节点中选出需要的100个超级节点  其余的900个超级节点进行投票选出10个超级节点 这10个超级节点进行顺序代理）
func pickNode(conn net.Conn) SelDels {
	/// 1.挑选SelDelegatesList
	indexs := make(map[int]bool, SelDelList)
	for i := 0; i < SelDelList; i++ {
		output:
		index := rand.Intn(AllDelList) //随机出一个节点的索引
		if _, ok := indexs[index]; !ok {
			goto output
		}
		indexs[index] = true
		mutex.Lock()
		SelDelegatesList[i].addresss = AllDelegatesList[index]
		mutex.Unlock()
	}

	// 广播SelDelegatesList
	go func(){
		for {
			msg := <- announcements
			io.WriteString(conn, msg)
		}
	}()

	/// 2. 其他超级节点进行投票
	for {
		var address string
		DelsVoted := make(map[string]bool, AllDelList)
		go func() {
			for {
				count := 0
				for _, v := range DelsVoted {
					if v == true {
						count++
					}
				}
				if count == AllDelList {
					return
				}
			}
		}()
		io.WriteString(conn, "Enter Vote index :")
		scanBalance := bufio.NewScanner(conn)
		index, _ := strconv.Atoi(scanBalance.Text())
		mutex.Lock()
		SelDelegatesList[index].votes = SelDelegatesList[index].votes + 1
		DelsVoted[address] = true
		mutex.Unlock()
	}

	/// 3.选出票数最高的前10位代理人
	sort.Sort(SelDelegatesList)
	WinDelegatesList = SelDelegatesList[:WinDelList]
	log.Println("WinDelegatesList")
	return WinDelegatesList
}

func main() {
	//err := godotenv.Load()
	//if err != nil {
	//	log.Fatal(err)
	//}
	// 创建初始区块
	genesisBlock := &Block{
		Index:     0,
		Timestamp: time.Now().Format("20060102150405"),
		BPM:       0,
		PrevHash:  "",
		Delegate:  "",
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

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		pickNode(conn)
	}
}


func (SelList SelDels) Len() int {
	return len(SelList)
}

func (SelList SelDels) Swap(i, j int) {
	SelList[i], SelList[j] = SelList[j], SelList[i]
}

func (SelList SelDels) Less(i, j int) bool {
	return SelList[j].votes < SelList[i].votes
}