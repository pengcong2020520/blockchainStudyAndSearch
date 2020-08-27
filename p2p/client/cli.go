package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	//增加可视性
	if len(os.Args) < 2 {
		fmt.Println("Client Flag == ")
		fmt.Println("cli.exe + Flag !!")
		os.Exit(0)
	}
	tag := os.Args[1]

	//原进程地址
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 8081}
	//目标进程地址
	dstAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 8080}
	conn, _ := net.DialUDP("udp", srcAddr, dstAddr)
	conn.Write([]byte("hello !!! " + tag))
	data := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(data)
	if err != nil {
		fmt.Printf("Failede to read from udp !", err)
	}
	conn.Close()
	fmt.Println(string(data))
	peer := parseAddr(string(data[:n]))
	fmt.Printf("local: %s server: %s peer: %s\n", srcAddr, remoteAddr, peer)
	connect(srcAddr, &peer, tag)
}

func parseAddr(addr string) net.UDPAddr {
	t := strings.Split(addr, ":")
	port, _ := strconv.Atoi(t[1])
	return net.UDPAddr{
		IP:   net.ParseIP(t[0]),
		Port: port,
		Zone: "",
	}
}

func connect(src *net.UDPAddr,dst *net.UDPAddr, tag string) {
	conn, err := net.DialUDP("udp", src, dst)
	if err != nil {
		fmt.Println("Failed to dial udp! ", err)
	}
	defer conn.Close()
	if _, err := conn.Write([]byte("hand one, hi nihao ! ")); err != nil {
		fmt.Println("Failede to hand one!! ", err)
	}
	go func() {
		for {
			time.Sleep(10 * time.Second)
			if _, err = conn.Write([]byte("tag=>")); err != nil {
				fmt.Println("Failed to send msg", err)
			}
		}
	}()

	for {
		data := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Println("Failed to read data!! ", err)
		} else {
			fmt.Println("Get data %s success! ", data[:n])
		}
	}
}
