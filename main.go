package main

import (
	"log"
	"net"
	"os"
)

type DevType string
type DevId string

// DevInfo 设备信息 ：设备类型、设备建立的socket连接
type DevInfo struct {
	devType DevType
	conn    net.Conn
}

// serviceConn 和主服务器的连接
// var serviceConn net.Conn

// serviceAddr 主服务器地址
var serviceAddr string

// devConnTable 设备连接表,DevId直接使用连接端口
var devTable = make(map[DevId](*DevInfo), 50)

func GetConn(id DevId) {
	re
}

// servCh 客户端先将数据发送到servCh通道，然后由HandleServ处理发给服务器
var servCh = make(chan []byte, 100)

type handle func(data string)

var devHandle = make(map[DevType]handle, 10)

// InitLoger 初始化log配置
func InitLoger() error {
	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.SetOutput(file)
		return nil
	}
}

func initConfig() error {
	serviceAddr = "192.168.10.112"
	return nil
}

func handleBuff([]byte) string {
	return `{"type":{}}`
}
func handleMsg(msg string) {
	var devType DevType
	var msgType string
	switch msgType {
	case "ServRsp":
	case "DevDataReq":
		devHandle[devType](msg)
	default:
	}

}

//
func createDevTable(data []) {

}

// HandleSendServ 接收设备端发来的数据集中发送到主服务器
func HandleSendServ(serviceConn net.Conn) {
	defer serviceConn.Close()
	for {
		data := <-servCh
		serviceConn.Write(data)
	}
}

// HandleRecvServ 接收到主服务器发送的数据，然后分发给各设备连接处理
func HandleRecvServ(serviceConn net.Conn) {
	defer serviceConn.Close()
	for {
		buff := make([]byte, 1024)
		serviceConn.Read(buff)
		msg := handleBuff(buff)
		handleMsg(msg)
	}
}

// InitServerConn 初始化连接主服务器连接
func ReqDevList() {
	servCh <- []byte(`{"MsgType":"ServReq","Action":"GetDevList"}`)
}
func InitServerConn() error {
	//建立服务器连接
	serviceConn, err := net.Dial("tcp", serviceAddr)
	if err != nil {
		log.Printf("连接服务器错:%s,%s\n", serviceAddr, err.Error())
		return err
	}
	go HandleSendServ(serviceConn)
	go HandleRecvServ(serviceConn)
	return nil
}

// devCreateListen 使用设备ID即固定端口号创建监听设备
func devCreateListen(port DevId) (net.Listener, error) {
	listen, err := net.Listen("tcp", "localhost:"+string(port))
	if err != nil {
		log.Printf("监听失败:%s,%s\n", "localhost:"+port, err.Error())
		return nil, err
	}
	return listen, nil
}

// devAcceptConn 等待设备连接，创建连接
func devAcceptConn(l net.Listener, port DevId) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			continue
		}
		devConnTable[port].conn = conn
	}
}

func IntiDeviceConn() (err error) {
	for port, _ := range devConnTable {
		l, err := devCreateListen(port)
		if err == nil {
			go devAcceptConn(l, port)
		} else {
			log.Printf("监听[%s]建立连接错误:%s\n", port, err.Error())
		}
	}
	return err

}

func main() {
	initConfig()
	InitLoger()
	InitServerConn()
	IntiDeviceConn()
}
