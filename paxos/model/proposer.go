package model

import (
	"go_code/paxos/utils"
	"log"
)

//定义一个提案者结构体
type Proposer struct {
	addr string
	value string
	acptors map[string]utils.Message
	node utils.Node
	curN int
	random int
}

func (pro *Proposer) prepare() []utils.Message {
	pro.curN++

	txMsgNum := 0
	var msgList []utils.Message
	log.Printf("proposer have %v major msg", len(pro.acptor))
	for acptorAddr, _ := range pro.acptors {
		msg := utils.Message{
			FromAddr : pro.addr,
			ToAddr : acptorAddr,
			MsgType : utils.MsgPropose,
			curN : pro.getCurN(),
			Value : pro.value,
		}
		msgList = append(msgList,msg)
		txMsgNum++
		if txMsgNum > pro.majority() {
			break
		}
	}
	return msgList
}

func (pro *Proposer) getCurN() {
	pro.curN = pro.random<<len(pro.addr) | pro.addr
}

func (pro *Proposer) majority() {

}