package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

const TIMEOUT = time.Second * 30

func main() {
	//建立一个本地的udp协议网络
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 8080})
	if err != nil  {
		fmt.Println("Failed to listen udp! ", err)
		return
	}
	fmt.Println("local addr : ", listener.LocalAddr().String())
	peers := make([]net.UDPAddr, 0, 2)
	data := make([]byte, 1024)
	for {
		n, remoteAddr, err := listener.ReadFromUDP(data)
		if err != nil || n <= 0 {
			fmt.Println("Failed to listen data from udp ! ", err)
		}
		fmt.Printf("addr: %s ==> data: %s \n", remoteAddr.String(), data[:n])
		peers = append(peers, *remoteAddr)
		if len(peers) == 2 {
			log.Printf("addr: %s <=====> addr: %s \n", peers[0].String(), peers[1].String())
			listener.WriteToUDP([]byte(peers[1].String()), &peers[0])
			listener.WriteToUDP([]byte(peers[0].String()), &peers[1])
			time.Sleep(TIMEOUT)
			return
		}
	}
}