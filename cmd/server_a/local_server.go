package main

import (
	"flag"
	"log"
	"net"
	"sync"
	"time"

	"github.com/javyliu/proxy/internal"
	"github.com/javyliu/proxy/pkg/aescrypto"
)

var localIp *string
var serverIp *string
var key *string
var timeout *int

func main() {
	localIp = flag.String("lip", ":18305", "本地服务监听地址")
	serverIp = flag.String("rip", ":18304", "远程服务监听地址")
	key = flag.String("key", "test", "aes加密key")
	timeout = flag.Int("td", 60, "连接到远程服务器的超时时间单位 秒")

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
	var wg sync.WaitGroup

	defer log.Println("----conn is closed")

	defer conn.Close()
	bConn, err := net.DialTimeout("tcp", *serverIp, time.Second*5)

	// 发送到服务B
	if err != nil {
		log.Println(&conn, "[error_dial]", err)
		return
	}
	bConn.SetDeadline(time.Now().Add(time.Second * time.Duration(*timeout)))
	defer bConn.Close()

	aeschiper, err := aescrypto.New(*key)
	if err != nil {
		log.Println(err)
		return
	}

	clientA := internal.NewClient(conn)
	clientB := internal.NewClient(bConn)

	log.Println("AconnId:", clientA.Id, "BconnId:", clientB.Id)
	wg.Add(2)
	// stopChannel := make(chan bool, 2)

	// 加密并发送到服务B
	go func() {
		defer wg.Done()
		defer log.Println("[-------A closed]", clientA.Id)
		log.Println("start client C -> A -> B")
		// defer func() { stopChannel <- true }() // stopChannel <- true
		// defer log.Println("[---------A close]", clientA.Id)
		// aeschiper.ReadAndWrite(conn, bConn, true)
		aeschiper.ReadAndWriteStream(*clientA, *clientB, true)
	}()

	// 从服务B读取并解密然后发送到客户端
	go func() {
		// defer func() { stopChannel <- true }() // stopChannel <- true

		defer wg.Done()
		defer log.Println("[-------B  closed]", clientB.Id)
		log.Println("start client B -> A -> C")
		// aeschiper.ReadAndWrite(bConn, conn, false)
		aeschiper.ReadAndWriteStream(*clientB, *clientA, false)
	}()

	// <-stopChannel
	// time.Sleep(5 * time.Second)
	wg.Wait()

}
