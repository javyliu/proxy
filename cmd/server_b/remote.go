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

	localIp = flag.String("lip", ":18304", "本地服务监听地址")
	serverIp = flag.String("rip", "127.0.0.1:1080", "远程服务监听地址")
	key = flag.String("key", "test", "aes加密key")

	flag.Parse()

	log.SetPrefix("[remote] ")

	ln, err := net.Listen("tcp", *localIp)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("listen on:", *localIp)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[error_accept]", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer log.Println("conn is closed")

	defer conn.Close()

	socksConn, err := net.DialTimeout("tcp", *serverIp, 5*time.Second)
	clientA := internal.NewClient(conn)
	clientB := internal.NewClient(socksConn)
	//  to SOCKS
	if err != nil {
		log.Println(clientA.Id, "[error_dial]", err)
		return
	}
	defer socksConn.Close()

	aeschiper, err := aescrypto.New(*key)
	if err != nil {
		log.Println(err)
		return
	}

	//  to SOCKS
	go aeschiper.ReadAndWriteStream(*clientA, *clientB, false)

	//  from SOCKS
	go aeschiper.ReadAndWriteStream(*clientB, *clientA, true)
}
