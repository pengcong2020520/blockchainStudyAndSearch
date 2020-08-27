package p2p

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
)

const messageId = 0
type Message string

func main() {
	nodekey, _ := crypto.GenerateKey()
	server := p2p.Server{
		MaxPeers : 10,
		PrivateKey : nodekey,
		Name : "my nod name",
		ListenAddr : "30300",
		Protocols : []p2p.Protocol{NewProtocol()},
	}
	server.Start()
	select {}    //阻塞

}

func NewProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    "NewProtocol",
		Version: 1,
		Length:  1,
		Run:     msgHandler,
	}
}

func msgHandler(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	for {
		msg, err := ws.ReadMsg()    //从P2P网络中读取msg
		if err != nil {
			return err
		}
		var myMessage [1]Message
		err = msg.Decode(&myMessage) // 对数据进行编码并放入到myMessage中
		if err != nil {
			continue
		}

		switch myMessage[0] {
		case "foo":
			err := p2p.SendItems(ws, messageId, "bar")
			if err != nil {
				return err
			}
		default:
			fmt.Println("recvierMsg : ", myMessage)
		}
	}

	return nil
}







