package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"

	"github.com/javyliu/proxy/pkg/proto"
)

var localIp *string
var serverIp *string

func main() {
	localIp = flag.String("lip", "127.0.0.1:20000", "本地服务监听地址")
	serverIp = flag.String("rip", "127.0.0.1:20001", "远程服务监听地址")

	log.SetPrefix("[proxy server] ")

	flag.Parse()

	listener, err := net.Listen("tcp", *localIp)
	if err != nil {
		log.Printf("create listener failed, err: %v", err)
		return
	}
	defer listener.Close()
	log.Println("listened on", *localIp)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener accept failed, err:", err)
			return
		}
		go process(conn)
	}
}

func process(conn net.Conn) {
	// process handles a single connection, reading messages from it and printing them.
	defer conn.Close()
	reader := bufio.NewReader(conn)
	conn, err := net.Dial("tcp", *serverIp)
	if err != nil {
		log.Printf("connect to %v failed, err: %v \n", *serverIp, err)
		return
	}
	defer conn.Close()

	// 把本地数据加密后发送到远端
	for {
		msg, err := proto.Decode(reader)

		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("read from conn failed, err:", err)
			return
		}
		log.Println("read from conn succ, msg:", msg)
	}

}
