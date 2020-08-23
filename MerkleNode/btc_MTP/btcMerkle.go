package btc_MTP

import "crypto/sha256"

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

//生成Merkle树中的节点  如果是叶子节点 则左子树 右子树Left，Right 为nil，
// 					   如果为非叶子节点  根据Left right生成当前节点的hash
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

//构建merkle tree
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

func big2little(data []byte) {
	for i, j := 0, len(data) - 1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

