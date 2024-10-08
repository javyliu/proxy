package aescrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
	"net"

	"github.com/javyliu/proxy/internal"
)

// func init() {
// 	log.SetPrefix("[aescbc] ")
// }

type AesChiper struct {
	Block   *cipher.Block
	Iv      *[]byte
	AconnId string
	BconnId string
}

func New(key string) (*AesChiper, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	_, err = rand.Read(iv)
	if err != nil {
		return nil, err
	}

	return &AesChiper{
		Block: &block,
		Iv:    &iv,
	}, nil
}

func (c *AesChiper) Encrypt(data []byte) []byte {
	padData := pad(data, aes.BlockSize)

	cbc := cipher.NewCBCEncrypter(*c.Block, *c.Iv)

	ciphertext := make([]byte, len(padData)+aes.BlockSize)
	copy(ciphertext[:aes.BlockSize], *c.Iv)
	cbc.CryptBlocks(ciphertext[aes.BlockSize:], padData)
	return ciphertext
}

func (c *AesChiper) Decrypt(data []byte) []byte {

	iv := data[:aes.BlockSize]
	cbc := cipher.NewCBCDecrypter(*c.Block, iv)

	size := len(data) - aes.BlockSize
	plaintext := make([]byte, size)
	cbc.CryptBlocks(plaintext, data[aes.BlockSize:])
	return unpad(plaintext)
}

func pad(data []byte, blockSize int) []byte {

	// paddingSize := blockSize - len(data)%blockSize

	// return append(data, bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)...)

	padding := blockSize - (len(data) % blockSize)
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := 0; i < padding; i++ {
		padded[len(data)+i] = byte(padding)
	}
	return padded
}
func unpad(data []byte) []byte {

	length := len(data)
	paddingSize := int(data[length-1])

	plaintext := data[:(length - paddingSize)]

	return plaintext
}

// 读取数据并写入
// true 代表先加密后写入bConn false 代表先解密后写入bConn
func (c *AesChiper) ReadAndWrite(conn net.Conn, bConn net.Conn, encrypt bool) error {

	bufSize := 102400
	// 先做加密时需要做padd操作，此时会加1-16（aes.BlockSize）个字节，另外再加上16个字节的iv
	var readSize int
	if encrypt {
		readSize = bufSize - 32
	} else {
		readSize = bufSize
	}

	for {
		buf := make([]byte, bufSize)
		n, err := conn.Read(buf[:readSize])
		if err != nil {
			if err == io.EOF {
				log.Println(c.AconnId, "[read_eof]", err)
				// 这里有问题，如果在遇到io.EOF之时，buf中还有数据，应该再次处理后传到bConn中, 所以加上sendToBconn方法

				if n > 0 {
					if err := sendToBconn(bConn, buf[:n], c, encrypt); err != nil {
						return err
					}

				}
				break
			} else {
				log.Println(c.AconnId, "[error_read]", err)
				return err
			}
		}

		//log.Println(c.AconnId, "[read_length]:", n, encrypt)

		if err := sendToBconn(bConn, buf[:n], c, encrypt); err != nil {
			return err
		}
	}

	return nil
}

func sendToBconn(bConn net.Conn, buf []byte, c *AesChiper, encrypt bool) error {
	var outdata []byte
	if encrypt {
		outdata = c.Encrypt(buf)
		log.Println(c.AconnId, "[加密后]:", len(outdata))
	} else {
		outdata = c.Decrypt(buf)
		log.Println(c.BconnId, "[解密后]:", len(outdata))
	}

	_, err := bConn.Write(outdata)

	if err != nil {
		if err == io.ErrClosedPipe {
			log.Println(c.BconnId, "[remote_closed]", err)

		} else {
			log.Println(c.BconnId, "[error_write]", err)
		}
		return err
	}
	return nil
}

// encrypt: true 代表 Client -> A -> B （从Client读取数据并加密发送到B）
// encrypt: false 代表 B -> Socket （从B解密数据并发送到Socket）

func (c *AesChiper) ReadAndWriteStream(src internal.Client, dst internal.Client, encrypt bool) error {
	var iv []byte
	var stream cipher.Stream
	if encrypt {
		iv = *c.Iv
		// 将 IV 写入目标连接，以便解密时使用
		if _, err := dst.Conn.Write(iv); err != nil {
			return err
		}
		stream = cipher.NewCFBEncrypter(*c.Block, iv)
	} else {
		iv = make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(src.Conn, iv); err != nil {

			return err
		}
		stream = cipher.NewCFBDecrypter(*c.Block, iv)

	}

	writer := cipher.StreamWriter{S: stream, W: dst.Conn}
	_, err := io.Copy(writer, src.Conn)
	log.Println("[io_copy]", src.Id, dst.Id)

	if err != nil {
		log.Println("[error_copy]", src.Id, dst.Id, err)
		return err
	}
	return nil
}
