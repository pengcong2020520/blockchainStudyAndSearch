
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
	POW的激励机制，使得作恶需要付出的成本远大于维护系统稳定的成本，有效解决了拜占庭将军问题，使得系统极其安全，且较为稳定。
	挖矿难度较难，但结果却可以很轻松的进行验证，使得系统的稳定性较高。
	但在挖矿的过程中，很多没挖出来的区块直接丢弃，浪费了大部分的资源。且现在矿池的出现，使得全球大部分算力集中在矿池手里，如果日后一家独大，容易出现中心化的危机。
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
	矿工不用拼算力去产生区块，且出块速度相比于POW显著提高，相比于POW，在提高tps的同时，也不会浪费太多的算力。
	经济市场长远的角度看，拥有大量代币的节点将更有效的去获取区块奖励，使得代币不易流通，对于拥有大量代币的人来说更容易吸引来自黑客的“青睐”。

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
> 一开始设定时，默认一部分具有权威说话权的节点，这些节点是维护区块链安全与性能的
核心。当然，如果超级节点中有恶意节点做了违反法则的事情，则被踢出超级节点列表。

#### 特性
可以达到百万甚至千万级别的tps，比POS高几个数量级；
超级节点之间不存在争夺出块的情况，避免了出块时浪费区块的情况，以及不会遗漏区块。

#### 代码实现
* **基础数据结构**
```
type Block struct {
	Index  int			// 高度
	Timestamp string		// 时间戳
	BPM int				// 交易信息
	PrevHash string     // 上一个哈希值
	HashCode string		// 当前的哈希值
	Delegate string    //  代理人 （超级节点）
}
```
	+ 在DPOS中有个超级节点的概念，就是这个节点的权限比其他普通节点的权限要大，超级节点主要用于维护整个区块链网络的安全性以及性能方面的需求。
	+ 超级节点加入，使得基于DPOS的区块链网络的性能提高了很多，但同时由于超级节点的加入，使得人们对基于DPOS的区块链有着争议，因为在这个网络中并不是完全的去中心化的网络。
	+ 在设计的数据结构中，`Delegate`就代表着出这个区块的超级节点。
	
* **挑选超级节点**

	DPOS的核心部分就在于如何挑选超级节点，这里用个通俗的例子说明：

	> 例如从1000个超级节点（超级节点列表）中选出需要的100个超级节点（备选超级节点），其余的900个超级节点进行投票选出10个超级节点，这10个超级节点进行顺序代理。

	基于上述思想，数据结构如下：
```
	type SelDelegates struct {  //被选中的超级节点 
		addresss string 
		votes int //其他超级节点对其的投票数
	}
	type SelDels []SelDelegates //被选中的超级节点的集合
	const AllDelList = 1000
	const SelDelList = 100
	const WinDelList = 10
	var Blockchain []Block  //定义一个区块链
	var announcements = make(chan string) // 也是一个通道 主GO TCP服务器将向所有节点广播最新的区块链
	var mutex = &sync.Mutex{}	//防止同一时间产生多个区块
	var AllDelegatesList []string   //所有的超级节点列表
	var SelDelegatesList SelDels  //被选中的超级节点
	var WinDelegatesList SelDels  //最后代理的超级节点
```
基于上述的数据结构，每当出块时，需要挑选超级节点，代码如下：
```
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
```
其中这里值得说明的一点在于，`sort.Sort(SelDelegatesList)`这个是对`SelDelegatesList`进行票数排序，最后选出前10位最终代理者。所以在前面的数据结构中，定义了一个类型`type SelDels []SelDelegates`，基于Go语言的特性，对于，只要实现了相应接口内的方法，即可使用排序这个函数。代码如下：
```
func (SelList SelDels) Len() int {
	return len(SelList)
}
func (SelList SelDels) Swap(i, j int) {
	SelList[i], SelList[j] = SelList[j], SelList[i]
}
func (SelList SelDels) Less(i, j int) bool {
	return SelList[j].votes < SelList[i].votes
}
```
函数`func (SelList SelDels) Less(i, j int) bool`中，方法的不同可以控制所设计的数据结构是大叉堆还是小叉堆。

## Merkle结构树
在上面介绍的主流共识算法中，在构建区块时，都提到了一个交易信息的元素
```
type Block struct {
	......
	BPM int				// 交易信息
	......
}
```
在区块链中，从创世区块开始，整个区块的高度就会只增不减，由于每一个区块都与之前的区块相关联，所以存储的数据将会是指数级增长。那么在区块链中，存储的数据就不宜过大，在区块链中是如何进行这种交易数据的存储？？答案是Merkle树。
那么神奇的Merkle树到底是什么东西？
简单理解，Merkle树其实就是把一大串的交易信息糅合成了一串唯一性的数字，这串唯一性的数字又极其便利去验证所有的交易。他的做法是：
* 1.先将所有交易进行hash得到对应的hash值
* 2.然后将同一时间段内的所有的交易进行排序，然后将交易两两一对
* 3.将两两结对的hash值进行组合，然后继续按顺序两两结对，直到最后只剩下一个hash值，那这个值就是唯一的值，也就记录着所有的交易信息。
基于上述思想，对Merkle树进行代码实现：

**基本数据结构**
```go
//Merkle 树
type MerkleTree struct {
	RootNode *MerkleNode
}

//Merkle 节点
type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	Data []byte
}
```
**生成Merkle树中的节点**
```go
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	merklenode := MerkleNode{}
	if left == nil && right == nil {
		merklenode.Data = data
	} else {
		//将两个数据进行合并 区块链中应该默认均为大端存储
		sum := append(left.Data, right.Data...)
		//进行 double hash
		oneHash := sha256.Sum256(sum)
		twoHash := sha256.Sum256(oneHash[:])
		//把值赋给merkle节点的data
		merklenode.Data = []byte(twoHash[:])
	}
	merklenode.Left = left
	merklenode.Right = right

	return &merklenode
}
```
生成Merkle树中的节点，如果是叶子节点 则左子树 右子树Left，Right 为nil；如果为非叶子节点  根据Left、Right生成当前节点的hash。

**构建Merkle树**
```go
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	//构建叶子节点
	for _, nodeData := range data {
		node := NewMerkleNode(nil, nil, nodeData)
		nodes = append(nodes, *node)
	}
	j := 0
	//每次循环 数据量减半
	for dataAmount := len(data); dataAmount > 1; dataAmount = (len(data)+1)/2 {
		//每两个数据拼接到一起  分为奇偶两种情况
		if dataAmount%2 == 0 {   //当data数量为偶数时
			for i := 0; i < len(data); i += 2 {
				node := NewMerkleNode(&nodes[j+i], &nodes[j+i+1], nil)
				nodes = append(nodes, *node)
			}
			j += dataAmount  //切换到下一层
		} else if dataAmount%2 == 1 {
			for i := 0; i < len(data); i += 2 {
				if i == len(data)-1 {
					node := NewMerkleNode(&nodes[j+i], &nodes[j+i], nil)
					nodes = append(nodes, *node)
				}
				node := NewMerkleNode(&nodes[j+i], &nodes[j+i+1], nil)
				nodes = append(nodes, *node)
			}
			j += dataAmount  //切换到下一层
		}
	}
	mTree := MerkleTree{&(nodes[len(nodes)-1])}
	return &mTree
}
```
其中，需要注意的是，交易的数量分为奇数和偶数两种情况。如果是偶数情况，则正常处理，如果是奇数情况就对单笔交易进行hash处理即可。


## MPT树
以太坊采用的是MPT树，是一种融合了默克尔树和前缀树两种结构树结构优点的数据结构，用来管理账户数据、生成交易集合hash的重要数据结构。
### 特性
存储任意长度的key-value键值对数据；
提供了一种快速计算所需要维护的数据以及哈希标识的一种机制；
提供了快速状态回滚机制；
提供了一种成为默克尔证明的证明方法，进行轻节点的扩展，实现简单的支付验证；

### 前缀树
用于保存关联数组，key一般为字符串，前缀树节点的位置由key的内容决定，即前缀树的key值被编码在根节点到该节点的路径中。
#### 特性
相对于hash表来说，前缀树查询拥有共同前缀的key的数据时十分高效。前缀树只需要遍历对应前缀的节点即可，而hash需要遍历整个表，对于最差的情况可能和hash一致。
相比于hash来说，前缀树不会存在hash冲突。
一次查找的效率较低，查找效率取决于key的长度，假设key长度为m，则需要进行m次IO开销，对磁盘压力较大。
容易浪费空间，如果某个key很长且没有相同的前缀时，则需要专门创建一棵树的直接用来存储该key。
### merkle树
在比特币网络中，merkle树被用来归纳一个区块中所有的交易，同时生成整个交易集合的数字指纹。此外，由于merkle树的存在，使得比特币这种公链的场景下，可以使用“轻节点”进行简单支付验证，较为高效。
#### 特性
merkle树具有树的全部特点；
mekle树叶子节点的value是数据项的内容，或是数据项的hash值；
非叶子节点的value根据其孩子节点的信息，然后按照hash算法计算得出；

#### 优势
快速重哈希：当树节点内容发生变化时，能够在前一次hash计算的基础上，仅仅需要把新的hash得出来的值进行hash重计算，便可以得到一个新的hash根用来代表整个树的状态。
轻节点扩展：可以拓展一个轻节点，轻节点的特点就是对于每个区块，仅仅需要存储约80个字节大小的区块头数据（存储父区块哈希，世界状态哈希，交易回执集合哈希），不存储交易列表以及回执列表等数据，然而通过轻节点，可以实现在非信任的公链环境中,验证某一笔交易是否正确，或者这个状态是否在当前状态树中。这一点让区块链技术应用能够运行在个人PC以及智能手机等拥有小村出容量的终端上。
#### 劣势
存储空间开销大。

### MPT树的节点
MPT树中新增了几种不同类型的树节点，用于压缩整体的树高，降低操作的复杂度。
树节点分为：空节点，分支节点，叶子节点以及扩展节点

- 空节点用来表示空串。

- 分支节点用来表示所有拥有超过一个孩子节点以上的非叶子节点，
```
type fullNode struct {
	Children [17]node
	flags nodeFlag
}
type nodeFlag struct {
	hash hashNode
	gen uint16
	dirty bool
}
```

	在进行树操作前，首先会进行一次key编码的转换，将一个字节的高低四位内容分拆成两个字节存储。key的每一位的值的范围都会在[0,15]之间，因此一个分支节点的孩子至多只有16个。减小了每个分支节点的容量，在一定程度上增加了树高。
其中分支节点中最后一个孩子是用来存储数据自身的内容的，故每个分支节点就有着至多17个孩子。

	每个分支节点都会附带一个`nodeFlag`，记录一些辅助数据：
**脏标志**：当一个节点被修改时，成为变“脏”，该标志位被置为1；
**节点哈希**：当节点变“脏”时，字段置空，否则一直存储上次的计算结果，在需要进行hash运算时，可以直接使用；
**诞生标志**：该节点第一次被载入时（或被修改时），都会赋予一个计数值作为诞生标志，这个标志代表着该节点的“新旧程度”。系统会自动清楚内容中“太旧”的节点，防止占用的内存空间过多。

- 叶子节点
定义如下：
```
type shortNode struct {
	Key   []byte
	Val   node
	flags nodeFlag
}
```
`Key`为叶子节点中剩余的Key
`Val`用来存储叶子节点的内容，存储的一个数据项的内容

- 扩展节点
定义如下：
```
type shortNode struct {
	Key   []byte
	Val   node
	flags nodeFlag
}
```
`Key`为扩展节点中剩余的Key
`Val`用来存储其孩子节点在数据库中的索引值（该索引值也是孩子节点的哈希值），存储其孩子节点的引用；
Val主要是为了建立父节点与孩子节点之间的关联，且让从数据库中读取节点时，尽量避免不必要的IO开销；



注意:

    叶子节点与扩展节点的设计时实现高压缩的关键，例如前缀树中如果有一个较长的数据前缀与其他数据并没有相同的，会造成很大的资源浪费，所以引入一个shortNode将后面较长的数据变为一个shortNode， 其中Key用来保存剩余的较长数据。
    由于叶子节点与扩展节点的定义相同，通过在Key中加入特殊的标志来区分两种类型的节点。


### key值的编码
三种编码方式：Raw编码（原生的字符） Hex编码（扩展的16进制编码） Hex-Prefix编码（16进制前缀编码）
#### Raw编码
Raw编码就是原生的Key值，不做任何改变，这种编码方式的Key是MPT对外提供接口的默认编码方式。
#### Hex编码
将原Key的高低四位分拆成两个字节进行存储，这种转换后的Key的编码方式就是Hex编码。

转换规则如下：

    将Raw编码的每个字符，根据高4位低4位拆成两个字节；
    若该Key对应的节点存储的是真实的数据项内容（该节点是叶子节点），则在末位添加一个ASCII值位16的字符作为终止标识符；
    若该Key对应的节点存储的是另外一个节点的哈希索引（即该节点是扩展节点），则不加任何字符。

Hex编码用于对内存中MPT树节点key进行编码

#### HP编码
当节点加载到内存时， 用于区分节点的类型的编码。对存储在数据库中的叶子／扩展节点的key进行编码区分。

从Hex编码转换成HP编码的规则如下：

	若原key的末尾字节的值为16（即该节点是叶子节点），去掉该字节；
	在key之前增加一个半字节，其中最低位是原本key长度的奇偶信息，key长度为奇数，则该位为1；低2位是一个特殊的终止标记符，若该节点为叶子节点，则该位为1；
	若原本key的长度为奇数，则在key之前再增加一个值为0x0的半字节；
	将原本key的内容作压缩，即将两个字符以高4位低4位进行划分，存储在一个字节中（Hex扩展的逆过程）

HP编码用于对数据库中的树节点key进行编码

![编码转换关系](http://upyun-assets.ethfans.org/uploads/photo/image/bac0731f81cb4a2d9ac71abd6f9bcdc9.jpeg "编码转换关系")

### 安全的MPT
为了解决key长度的限制问题，在传入数据时，堆数据项的key进行了一次hash计算，sha3(key)，有效避免了树中出现长度很长的路径。
但是需要在数据库中存储额外的sha3(key)与key之间的对应关系。

## 不对称加密算法（数字签名）
本部分主要是以基于RSA算法的数字签名作为实例进行不对称加密算法说明介绍。
### 非对称加密算法的特性
* 用私钥（或者公钥）对消息进行加密，需要用公钥（或者私钥）对消息进行解密才能获得消息。
* 其中，公钥可以其他人公开，私钥需要个人进行保密，且公钥无法通过推算得到私钥。
### 非对称加密技术在区块链中的应用场景
* 信息加密：发送者A使用B的公钥对消息进行加密再发给B，B可以通过自己的私钥对消息进行解密。
* 数字签名：发送者A采用自己的私钥加密信息后，发送给B，B使用A的公钥对信息进行解密，从而可以确保信息是由A发送的。
* 登录认证：由客户端使用私钥加密 登录信息 后，发送给服务器端，服务器接收后采用该客户端的公钥进行解密并认证登录信息。

### 比特币系统的非对称加密机制
1. 由操作系统底层的随机数生成器生成一个256位的私钥
2. 生成的私钥通过SHA256和Base58编码得到一个用户端使用的私钥，这个私钥长度易识别且易书写
3. 系统底层生成的私钥通过seco256k1椭圆曲线算法，可以得到一个65字节的公钥（相当于一个随机数）
4. 将这个公钥进行SHA256和RIPEMD160编码得到20个字节的公钥，这是一个公钥的摘要结果
5. 通过这个20字节的公钥，进行SHA256和Base58编码可以得到一个33字符的比特币账户地址


### 基于RSA的数字签名与验证
#### RSA加密原理RSA加密原理
* 加密：C = M^e mod n

* 解密：M = C^d mod n

> 其中(e,n)即为公钥、（d,n）即为私钥，其中的e,n,d需要符合一定条件的取值

求解过程：

	需要有两个质数p,q，满足 n = pq;
	取正整数e,d,使得ed mod (p-1)(q-1) = 1  =>  当且仅当 e 与（p-1)(q-1) 互质时，存在 d。

由于公钥公开，即e,n公开，因此破解RSA私钥，演变为对n质因数分解求p,q。
实际中，n的长度为2048位以上，研究表明n>200位时，分解n就非常困难了，故RSA算法的安全性可以得到很高的保障。

#### 实现RSA加密解密

* 首先需要导入rsa加密包：`import "crypto/rsa"`，其中实现RSA时基于PKCS#1规范。

* 生成RSA钥匙对————私钥和公钥
> 1. 生成RSA的私钥，主要借助`func GenerateKey(random io.Reader, bits int) (*PrivateKey, error)`函数；但是生成的私钥为了便于保存，则需要进行加密，根据一些权威的存储说明先进行x509序列化再进行PEM序列化，这样就可以使得生成的私钥安全保存在电脑中。
> 2. 生成RSA的公钥，主要是通过私钥生成，存储同样需要进行上述处理。

代码实现如下：
```
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
```
* 用私钥对数据进行签名
	1. 直接调用上述所生成的私钥文件，对取出的私钥文件内容进行上述序列化的反序列换，即先进行pem反序列化，再进行x509反序列化
	2. 将需要发送的消息`data`进行hash处理，这样使得传输的消息更为安全，然后用rsa包中的`func SignPKCS1v15(rand io.Reader, priv *PrivateKey, hash crypto.Hash, hashed []byte) ([]byte, error)`函数获得进行了数字签名的消息

	**代码实现**：
	```
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
	```
* 用公钥对数据进行验证
与私钥基本相同，直接上代码：
```
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
```





















