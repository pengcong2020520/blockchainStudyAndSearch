package utils

type msgType int

const (
	MsgPrepare msgType = iota + 1
	MsgPromise
	MsgPropose
	MsgAccept
)
//定义一个消息类型
type Message struct {
	FromAddr string `json:"fromaddr"`
	ToAddr string `json:"toaddr"`
	MsgType msgType `json:"msgtype"`
	Value string `json:"value"`  //消息的内容
	curN int `json:"cur_n"`    //消息中所携带的N，即proposal给出的N值  N值用于确定通过的提案者
	preN int `json:"pre_n"`   // 消息中上一次携带的N值，保证每个提案者的提案都会被通过
}

func (msg *Message) getProposeVal() string {
	return msg.Value
}

func (msg *Message) getProposeN() string {
	return msg.curN
}