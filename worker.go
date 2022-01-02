package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

//实现随机文件名
func GetRoundName(size int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ=_*&^$#%-|"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < size; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

var cnt int64 = 0

func process(conn net.Conn, fileName string) {

	atomic.AddInt64(&cnt, 1)

	defer conn.Close()

	//按照文件名创建新文件
	rand.Seed(time.Now().Unix())
	fileName2 := strings.TrimRight(fileName, ".txt")
	file, err := os.Create("/gopath/src/dis_test_file/worker/file/" + fileName2 + GetRoundName(20) + ".txt")
	if err != nil {
		fmt.Printf("os.Create() failed   %v\n", err)
		return
	}
	defer file.Close()

	//从网络中读取数据
	for {
		buf := make([]byte, 4096)
		n, err1 := conn.Read(buf)

		//写入本地文件，读多少、写多少
		file.Write(buf[:n])
		if err1 != nil {
			if err1 == io.EOF {
				fmt.Println("文件", cnt, "完成")
			} else {
				fmt.Printf("conn.Read()方法执行错误   %v\n", err)
			}
			return
		}
	}
}

func main() {

	tcpaddr, err := net.ResolveTCPAddr("tcp", ":30001")
	if err != nil {
		fmt.Println("resolve failed, err:", err)
		return
	}

	//TCP协议监听目标端口
	listen, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return
	}

	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept failed, err:", err)
			continue
		}

		buf := make([]byte, 128)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("conn.Read()   failed   %v\n", err)
		}
		fileName := string(buf[:n])
		//回写ok给发送端
		conn.Write([]byte("ok"))
		go process(conn, fileName)
		fmt.Println(cnt)
	}

}
