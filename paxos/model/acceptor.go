package model

import "go_code/paxos/utils"

func NewAcceptor(id string, node *utils.Node, learners []int) *Acceptor {
	return &Acceptor{
		AcptId: id,
		Learners: learners,
		Node : node,
	}
}

type Acceptor struct {
	AcptId string
	Learners []int
	AcptMsg *utils.Message
	PromiseMsg *utils.Message
	Node *utils.Node
}

func (acptor *Acceptor) run() {
	for {
		msg := acptor.Node.MsgRx()
		if msg == nil {
			continue
		}
		switch msg.MsgType {
		case utils.MsgPrepare:
			promiseMsg :=
		case utils.MsgPromise:

		case utils.MsgPropose:

		}
	}
}

func (acptor *Acceptor) rxPrepare(prepareMsg utils.Message) *utils.Message {

}

