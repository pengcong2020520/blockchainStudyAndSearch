package utils

import (
	"fmt"
	"log"
	"sync"
	"time"
)

//定义一个网络连接
type network struct {
	recMsg map[string]chan Message
	trxMsg map[string]chan Message
}

//定义一个网络中的节点
type Node struct {
	Nodeid string
	Conn *network
}

var announcements = make(chan string) //广播通道

var mutex = &sync.Mutex{} //防止消息发生冲突

func (net *network) GetNode(id string) *Node{
	return &Node{
		Nodeid : id,
		Conn : net,
	}
}

// 从网络连接中发送消息msg出去
// 将msg消息写入到to的接收消息的映射中
func (net *network) sendTo(msg Message) {
	log.Printf("tx msg ----- %v -> %v : msg %v , msgType %v", msg.FromAddr, msg.ToAddr, msg.Value, msg.MsgType)
	//announcements<- fmt.Sprintf("tx msg ----- %v -> %v : msg %v , msgType %v", msg.FromAddr, msg.ToAddr, msg.Data, msg.MsgType)

	net.trxMsg[msg.FromAddr]<- msg
	mutex.Lock()
	net.recMsg[msg.ToAddr]<- msg
	mutex.Unlock()
	announcements<- fmt.Sprintf("tx msg ----- %v -> %v : msg %v , msgType %v", msg.FromAddr, msg.ToAddr, msg.Value, msg.MsgType)
}

// 接收消息msg
// 将id 在RxMsg哈希表中的消息取出，并返回
// 其中node的id即为发消息节点的地址
func (net *network) rxFrom(id string) *Message {
	select {
		case msg := <-net.recMsg[id]:
			log.Printf("rx msg ----- %v -> %v : msg %v , msgType %v", msg.FromAddr, msg.ToAddr, msg.Value, msg.MsgType)
			return &msg
		case <-time.After(time.Second):
			log.Println("can't get msg, time out!!!")
			return nil
	}
}

func (node *Node) MsgTx(msg Message) {
	node.Conn.sendTo(msg)
}

func (node *Node) MsgRx() *Message {
	return node.Conn.rxFrom(node.Nodeid)
}


