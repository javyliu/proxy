package main

import (
	"flag"
	"log"
	"net"

	"github.com/javyliu/proxy/pkg/proto"
)

func main() {
	localIp := flag.String("lip", "127.0.0.1:20000", "连接地址")
	flag.Parse()

	log.SetPrefix("[proxy client] ")
	conn, err := net.Dial("tcp", *localIp)
	if err != nil {
		log.Println("connect failed, err:", err)
		return
	}
	defer conn.Close()

	for i := 0; i < 1; i++ {
		msg := `Hello , Hello. How are you? GOPATH 是你的工作目录，所有Go项目都在这个目录下。
		
		internal 目录包含仅在当前项目内部可以被导入使用的代码。
		`
		data, err := proto.Encode(msg)
		if err != nil {
			log.Println("encode msg failed, err:", err)
			return
		}
		_, err = conn.Write(data)
		if err != nil {
			log.Println("write failed, err:", err)
			return
		}
		log.Println("send succ")
	}

}
