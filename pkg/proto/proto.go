package proto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

func init() {
	log.SetFlags(log.Lshortfile)

}

// 封包
func Encode(message string) ([]byte, error) {
	log.SetPrefix("[proto encode] ")
	// 读取消息的长度，转换成int32类型（占4个字节）
	length := int32(len(message))
	packet := new(bytes.Buffer)
	// 写入消息头
	err := binary.Write(packet, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	// 写入消息实体
	err = binary.Write(packet, binary.LittleEndian, []byte(message))
	if err != nil {
		return nil, err
	}
	return packet.Bytes(), nil
}

// encode a reader content
// encode a reader content
func EncodeReader(reader *bufio.Reader) ([]byte, error) {
	packet := new(bytes.Buffer)
	buf := new(bytes.Buffer)
	var length int32
	var buf2 [1024]byte
	for {
		n, err := reader.Read(buf2[:])
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		_, err = buf.Write(buf2[:n])
		if err != nil {
			return nil, err
		}
		length += int32(n)
	}
	if err := binary.Write(packet, binary.LittleEndian, length); err != nil {
		return nil, err
	}
	if err := binary.Write(packet, binary.LittleEndian, buf.Bytes()); err != nil {
		return nil, err
	}
	return packet.Bytes(), nil
}

// 解包
func Decode(reader *bufio.Reader) (string, error) {
	log.SetPrefix("[proto decode] ")

	// 读取消息的长度, 字节数组
	lengthByte, _ := reader.Peek(4)
	// 生成一个新的Buffer
	lengthBuff := bytes.NewBuffer(lengthByte)

	var length int32
	// 读取包的长度, 从 lengthBuff 中读取二进制数据到 length中
	err := binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	log.Println("binary length = ", lengthByte)
	log.Println("length0 = ", length)

	// len := binary.LittleEndian.Uint32(lengthByte)
	// log.Println("length1 = ", len)

	// Buffered returns the number of bytes that can be read from this reader.
	if int32(reader.Buffered()) < length+4 {
		return "", fmt.Errorf("message length error")
	}

	// 读取消息实体
	messageByte := make([]byte, int(4+length))
	_, err = reader.Read(messageByte)
	if err != nil {
		return "", err
	}
	return string(messageByte[4:]), nil
}

func handleTcpReader(conn net.Conn) ([]byte, error) {

	content, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	// Convert the length to an integer
	contentLength := int32(len(content))

	// Read the content
	content = make([]byte, contentLength)
	_, err = io.ReadFull(conn, content)
	if err != nil {
		return nil, err
	}

	// Return the length and content as a single byte slice
	result := make([]byte, 4+len(content))
	binary.LittleEndian.PutUint32(result[:4], uint32(contentLength))
	copy(result[4:], content)
	return result, nil
}
