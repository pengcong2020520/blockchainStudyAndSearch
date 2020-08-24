
**本部分主要结合比特币、以太坊对区块链的共识算法，数据结构进行探究**


## 共识算法
区块链，个人理解为是一种分布式有顺序的账本，而共识算法就是维护这个账本安全与性能的核心。
所以对于共识算法来说，主要期望的目的为：抗ASIC性，易验证性。

主流的共识算法有POW、POS以及DPOS等，下面就对主流的共识算法进行一定的分析。
### POW

#### 概念
+ 工作量证明，通过对某项计算工作的工作量最终得到相应的奖励（挖矿奖励）。
> 类似于解一道比较难的数学题，而计算该数学题的过程相当于只能通过枚举法进行，最终解出来的值达到一定的难度系数即满足需求的工作量，可以获得区块奖励。

#### 特性
	由于工作量证明的复杂性，使得极其安全，且较为稳定。
	但在挖矿的过程中，很多没挖出来的区块直接丢弃，浪费了大部分得资源。
#### 代码实现
* **基础数据结构**
	```
	type Block struct {
		Index  int			// 高度
		Timestamp string		// 时间戳
		BPM int				// 交易信息
		PrevHash string     // 上一个哈希值
		HashCode string		// 当前的哈希值
		Difficulty int    // 区块难度
		Nonce string    //防止双花
	}
	```
相对于其他共识算法来说，POW的区块数据结构独有的部分为 `Difficulty` 与 `Nonce`.

		Difficulty的设计控制了出块的时间，代表的是难度系数，也就是上文提到的一道很难的数学题的难度。
		Nonce的设计十分巧妙，因为相比于线下交易以及目前的集中式模式来说，区块链作为一种分布式账本的技术，“重放攻击”是一个比较关键的问题。Nonce的设计使得每笔交易或者信息独一无二，有效避免了“重放攻击”。
其中值得说明的地方在于BPM，代表的是交易信息，但是在区块链中使用了一种数据结构对交易信息进行存储，在下面的Merkle树部分会进行详细说明。

* **区块链出块相关数据结构**
	```go
	var Blockchain []Block  //定义一个区块链

	var mutex = &sync.Mutex{} //防止同一时间产生多个区块

	var announcements = make(chan string) //出块结束后向所有节点进行广播
	```
* **生成区块函数**
```go
//生成区块函数
//由旧区块 、 新的BPM结构 、 验证者  产生新的区块
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
```
其中产生区块的函数中，POW有个独有的产生Nonce的部分，这个部分主要是让矿工进行挖矿，直到有人将正确的Hash找到，即产生一个合理的Nonce随机数。其中验证hash值的代码如下：
```go
// 验证hash是否有效
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
```
	其中 新区块中的preHash存储上一个区块的Hash；
	for 循环通过循环改变Nonce,然后选出符合相应难度系数的Nonce；
	isHashValid 判断hash是否满足当前的难度系数，如果难度系数为2，则当前hash的前缀有2个0。

* **基于POW的区块链网络代码main函数**
```
func main() {
	// 创建初始区块
	genesisBlock := &Block{
		Index :	0,
		Timestamp :	time.Now().Format("20060102150405"),
		BPM :	0,
		PrevHash :	"",
		Difficulty : DIFFICULTY,
	}
	genesisBlock.HashCode = calculateBlockHash(*genesisBlock)

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
```

### POS

#### 概念
+ 权益证明，通过质押的方式参与区块的验证，最终通过随机选取，选出一个验证者获得奖励。
> 类似于一大部分人，每个人拿出一部分token来参与竞选验证者，最终选出一名获胜者获得验证权，对区块进行验证，从而获得奖励。

#### 特性
	出块速度相比于POW显著提高，对算力的浪费也相对减少。

#### 代码实现
* **基础数据结构**
	```
type Block struct {
	   BlockNumber  int			// 高度
	   Timestamp string		// 时间戳
	   Info int				// 交易信息
	   prevHash string     // 上一个哈希值
	   HashCode string		// 当前的哈希值
	   Validator string    // 区块验证者  其中POW为difficulty
}
	```
相对于其他共识算法来说，POS的区块数据结构独有的部分为 `Validator`。
		Validator 指代的是区块的验证者，每次出块从参与质押的验证者中选出一个拥有记账权节点来进行进行出块。
		
* **关于选择拥有记账权节点的数据结构**
	```go
	var Blockchain []Block  
	//定义一个区块链

	var tempBlocks []Block 
	// 临时存储单元 在区块被选出来之前 ，临时存储在这里 该单元最多包含多少个区块

	var candidateBlocks = make(chan Block) 
// 临时候选通道， 任何一个节点在提出一个新块时 都会把他先发到这个通道里
//最终会从该候选通道中选出一个区块

	var announcements = make(chan string) 
	// 也是一个通道 主GO TCP服务器将向所有节点广播最新的区块链

	var mutex = &sync.Mutex{}	
	//防止同一时间产生多个区块

	var validators = make(map[string]int)  
	//节点的临时存储map 同时也会保存每个节点质押的token数
	```
* **生成区块函数**
```go
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
```
本部分的主要逻辑：
	通过上一个区块`oldBlock`（主要用于验证），这个区块的相关交易信息`BPM`、区块最终选出的验证者`Validator`来产生一个新的区块。
而在POS中进行出块有个比较独有的地方在于选出验证者，代码如下：
```go
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
```
对以上代码的说明：
其中借助go语言协程并发的特性，通过`net.Conn`的链接从客户端读取需要的信息，例如验证者的`token`数以及输入的交易信息；对读取的信息进行序列化后，选取记账权的节点。

* **选取记账权节点**
```go
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
```
选择获取记账权节点中，通过conn读取到的验证者的token数，分配验证者在获取记账权中的比例。一般来说，验证者质押的token数越大，获取记账权的概率也就越高。

### DPOS
#### 概念
+ 委托代理权益证明，初始含有一些超级节点，每次会选出一部分超级节点负责代理出块。
> 一开始设定时，默认一部分具有权威说话权的节点，这些节点是维护区块链安全与性能的核心。

#### 特性
可以达到百万甚至千万级别的tps，比POS高几个数量级；
超级节点之间不存在争夺出块的情况，避免了出块时浪费区块的情况，以及不会遗漏区块。





