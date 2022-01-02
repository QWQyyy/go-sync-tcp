package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

//最大goroutine数量
var maxRoutineNum = 300

//锁打开文件
var lock = make(chan int, maxRoutineNum)

//锁连接数
var lock_conn = make(chan int, maxRoutineNum)

//最大连接数
var maxConn int

//循环次数
var maxLoop = 10

//测试响应时间
var tcp_times []float64

//测试响应百分位数
var tcp_times_bfw []float64

var wg sync.WaitGroup

func cnn_tcp(filePath string, i int) {
	defer wg.Done()

	//获取文件属性
	fileInfo, err := os.Stat(filePath)
	fileName := fileInfo.Name()
	if err != nil {
		fmt.Printf("os.Stat()   failed   %v\n", err)
		return
	}

	var conn net.Conn

	//计时开始
	start := time.Now().UnixNano()

	//建立TCP连接
	lock_conn <- 1
	if i%3 == 0 {
		tcpaddr1, err := net.ResolveTCPAddr("tcp", "vm1:30001")
		if err != nil {
			fmt.Println("resolve failed, err:", err)
			return
		}
		conn, err := net.DialTCP("tcp", nil, tcpaddr1)
		if err != nil {
			fmt.Println("dial failed, err:", err)
			return
		}
		trans_file(conn, fileName, filePath)
	} else if i%3 == 1 {
		tcpaddr2, err := net.ResolveTCPAddr("tcp", "vm2:30001")
		if err != nil {
			fmt.Println("resolve failed, err:", err)
			return
		}
		conn, err = net.DialTCP("tcp", nil, tcpaddr2)
		if err != nil {
			fmt.Println("dial failed, err:", err)
			return
		}
		trans_file(conn, fileName, filePath)
	} else {
		tcpaddr3, err := net.ResolveTCPAddr("tcp", "vm3:30001")
		if err != nil {
			fmt.Println("resolve failed, err:", err)
			return
		}
		conn, err = net.DialTCP("tcp", nil, tcpaddr3)
		if err != nil {
			fmt.Println("dial failed, err:", err)
			return
		}
		trans_file(conn, fileName, filePath)
	}
	<- lock_conn

	//计算时间
	end := time.Now().UnixNano()
	tcp_times[i] = float64(end-start) / 1000000.0
	tcp_times_bfw[i] += float64(end-start) / 1000000.0
}

func trans_file(conn net.Conn, fileName string, filePath string) {

	//发送文件名给接收端
	_, err := conn.Write([]byte(fileName))
	if err != nil {
		fmt.Printf("conn.Write(fileName)   failed   %v\n", err)
		return
	}
	//读取服务器回发数据go run
	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("conn.Read(buf)   failed   %v\n", err)
		return
	}

	if string(buf[:n]) == "ok" {
		lock <- 1
		//只读打开文件
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("os.Open()   failed   %v\n", err)
			return
		}
		defer file.Close()

		buf2 := make([]byte, 4096)

		for {
			//从本地文件中读数据，写给网络接收端，读多少，写多少
			n, err := file.Read(buf2) //file读到buf2
			if err != nil {
				if err != io.EOF {
					fmt.Printf("file.Read() failed   %v\n", err)
				}
				break
			}
			//写到网络socket中
			_, err = conn.Write(buf2[:n])
			if err != nil {
				fmt.Printf("Write conn failed! err:%v\n", err)
			}
		}
		<-lock
	}

}

var res = make([]float64, 10, 20)

func test(filePath string, i int) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	wg.Add(maxConn)
	//filePath := "/gopath/src/dis_test_file/master/emm.txt"
	for i := 0; i < maxConn; i++ {
		go cnn_tcp(filePath, i)
	}
	var t float64 = 0
	for _, v := range tcp_times {
		t += v
	}
	wg.Wait()

	//记录每次循环的平均响应时间
	res[i] = t / float64(maxConn)
}

func main() {

	flag.Parse()
	s := flag.Args()
	filePath := s[0]
	maxConn0, err := strconv.Atoi(s[1])
	if err != nil {
		fmt.Println("please reviewer you parameter!")
		return
	}

	maxConn = maxConn0

	//测试响应时间
	tcp_times = make([]float64, maxConn, maxConn)

	//测试响应百分位数
	tcp_times_bfw = make([]float64, maxConn, maxConn)

	var t float64

	for i := 0; i < maxLoop; i++ {
		test(filePath, i)
		t += res[i]
	}

	//所有循环次数的平均响应时间的平均值
	fmt.Println(t/float64(maxLoop), "ms")

	for _, v := range tcp_times_bfw {
		v /= 10
		fmt.Println(v)
	}

}
