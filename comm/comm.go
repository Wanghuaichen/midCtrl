//package serv 主要负责设备和服务的沟通

package serv

import (
	"bufio"
	"config"
	"fmt"
	"io"
	"log"
	"net"
)

// serviceConn 和主服务器的连接
// var serviceConn net.Conn

// serviceAddr 主服务器地址
//var serviceAddr string

// recvCh 客户端先将数据发送到servCh通道，然后由HandleServ处理发给服务器
var recvCh = make(chan []byte, 100)

// sendCh 从服务器收到数据，发送到sendCh通道供，其他模块读取
var sendCh = make(chan []byte, 100)

// SendServ 将数据发送到主服务器
func SendData(data []byte) {
	recvCh <- data
}

func GetData() []byte {
	return <-sendCh
}

// HandleSendServ 数据集中发送到主服务器
func handleSendServ(serviceConn net.Conn) {
	defer serviceConn.Close()
	for {
		data := <-recvCh
		serviceConn.Write(data)
	}
}

// HandleRecvServ 接收到主服务器发送的数据，然后分发给各设备连接处理
func handleRecvServ(serviceConn net.Conn) {
	defer serviceConn.Close()
	for {
		//buff := make([]byte, 1024)
		//serviceConn.Read(buff)
		msg, err := bufio.NewReader(serviceConn).ReadSlice('\n')
		if err == io.EOF {
			continue // 继续等待读
		} else if err != nil {
			log.Fatal("获取服务器数据错误:", err)
		}
		fmt.Println("收到:", msg, "str:", string(msg))
		//msg = strings.TrimRight(msg, "\n")
		msg = msg[:len(msg)-1] //移除最后的'\n'
		//msg := handleBuff(buff)
		sendCh <- msg
		//handleMsg(msg)
	}
}

//InitServerConn 初始化连接主服务器连接
func InitServ() error {
	//建立服务器连接
	serviceAddr := config.GetServiceAddr()
	serviceConn, err := net.Dial("tcp", serviceAddr)
	if err != nil {
		log.Printf("连接服务器错:%s,%s\n", serviceAddr, err.Error())
		return err
	}
	go handleSendServ(serviceConn) //处理发送给主服务器的数据
	go handleRecvServ(serviceConn) //接收处理来着主服务器的数据
	return nil
}
