package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/javyliu/proxy/internal"
	"github.com/javyliu/proxy/pkg/aescrypto"
)

// var wg sync.WaitGroup

var localIp *string
var serverIp *string
var key *string

func main() {
	localIp = flag.String("lip", ":18305", "本地服务监听地址")
	serverIp = flag.String("rip", ":18304", "远程服务监听地址")
	key = flag.String("key", "test", "aes加密key")

	flag.Parse()
	log.SetPrefix("[local] ")
	ln, err := net.Listen("tcp", *localIp)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("listen on :", *localIp)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("[accept]", conn.RemoteAddr())
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer log.Println("conn is closed")
	defer conn.Close()
	// bConn, err := net.Dial("tcp", *serverIp)
	bConn, err := net.DialTimeout("tcp", *serverIp, 5*time.Second)
	//  to SOCKS
	clientA := internal.NewClient(conn)
	clientB := internal.NewClient(bConn)
	if err != nil {
		log.Println(clientA.Id, "[error_dial]", err)
		return
	}
	defer bConn.Close()

	aeschiper, err := aescrypto.New(*key)
	if err != nil {
		log.Println(err)
		return
	}

	//  to SOCKS
	go aeschiper.ReadAndWriteStream(*clientA, *clientB, true)

	//  from SOCKS
	go aeschiper.ReadAndWriteStream(*clientB, *clientA, false)
}
