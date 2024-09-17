package main

import (
	"flag"
	"log"
	"net"
	"time"

	"github.com/javyliu/proxy/internal"
	"github.com/javyliu/proxy/pkg/aescbc"
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
	defer conn.Close()
	bConn, err := net.Dial("tcp", *serverIp)
	// 发送到服务B
	if err != nil {
		log.Println(&conn, "[error_dial]", err)
		return
	}
	defer bConn.Close()

	aeschiper, err := aescbc.New(key)
	if err != nil {
		log.Println(err)
		return
	}

	clientA := internal.NewClient(conn)
	clientB := internal.NewClient(bConn)

	defer log.Println(clientA.Id, "A connection closed")
	defer log.Println(clientB.Id, "B connection closed")
	log.Println("#AconnId:", clientA.Id, "BconnId:", clientB.Id)
	// wg.Add(2)
	stopChannel := make(chan bool, 2)

	// 加密并发送到服务B
	go func() {
		// defer wg.Done()
		defer func() { stopChannel <- true }() // stopChannel <- true
		defer log.Println("[---------A close]", clientA.Id)
		// aeschiper.ReadAndWrite(conn, bConn, true)
		aeschiper.ReadAndWriteStream(*clientA, *clientB, true)
	}()

	// 从服务B读取并解密然后发送到客户端
	go func() {
		defer func() { stopChannel <- true }() // stopChannel <- true

		// defer wg.Done()
		defer log.Println("[---------B  close]", clientB.Id)
		// aeschiper.ReadAndWrite(bConn, conn, false)
		aeschiper.ReadAndWriteStream(*clientB, *clientA, false)
	}()

	<-stopChannel
	time.Sleep(5 * time.Second)
	// wg.Wait()

}
