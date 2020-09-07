package model

import "go_code/paxos/utils"

type Learner struct {
	addr string
	passMsg map[string]utils.Message //已通过的提案消息
	node utils.Node
}

func NewLearner(addr string, node utils.Node, acptorAddrs []int) *Learner {
	newLearner := &Learner{addr: addr, node: node}
	newLearner.passMsg = make(map[int]utils.Message)
	for _, aceptAddr := range acptorAddrs {
		newLearner.passMsg[aceptAddr] = utils.Message{}
	}
	return newLearner
}




