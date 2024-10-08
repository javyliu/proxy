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
	var wg sync.WaitGroup

	defer log.Println("----conn is closed")

	defer conn.Close()

	socksConn, err := net.DialTimeout("tcp", *serverIp, time.Second*5)

	//  to SOCKS
	if err != nil {
		log.Println(&conn, "[error_dial]", err)
		return
	}
	socksConn.SetDeadline(time.Now().Add(30 * time.Second))

	defer socksConn.Close()

	aeschiper, err := aescrypto.New(*key)
	if err != nil {
		log.Println(err)
		return
	}

	clientA := internal.NewClient(conn)
	clientB := internal.NewClient(socksConn)

	log.Println("AconnId:", clientA.Id, "BconnId:", clientB.Id)

	// stopChannel := make(chan bool, 2)
	// defer close(stopChannel)

	wg.Add(2)
	//  to SOCKS
	go func() {
		// defer func() { stopChannel <- true }() // stopChannel <- true

		defer wg.Done()
		defer log.Println("[-------B closed]", clientA.Id)
		log.Println("start client B -> S")
		// defer log.Println("[---------A close]", clientA.Id)

		// aeschiper.ReadAndWrite(conn, socksConn, false)
		aeschiper.ReadAndWriteStream(*clientA, *clientB, false)
	}()

	//  from SOCKS
	go func() {
		// defer func() { stopChannel <- true }() // stopChannel <- true

		defer wg.Done()
		defer log.Println("[-------S closed]", clientB.Id)
		log.Println("start client B -> A")
		// defer log.Println("[---------B  close]", clientB.Id)

		// aeschiper.ReadAndWrite(socksConn, conn, true)
		aeschiper.ReadAndWriteStream(*clientB, *clientA, true)

	}()
	// <-stopChannel
	// time.Sleep(5 * time.Second)
	wg.Wait()
}
