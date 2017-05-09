package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"

	"fmt"

	"./devices"
)

//type DevType string

//type DevId string

// DevInfo 设备信息 ：设备类型、设备建立的socket连接
/*type DevInfo struct {
	devType DevType
	conn    net.Conn
}*/

// serviceConn 和主服务器的连接
// var serviceConn net.Conn

// serviceAddr 主服务器地址
var serviceAddr string

// devConnTable 设备连接表,key为设备ID直接使用连接端口号做ID
var devConnTable = make(map[string](*net.Conn), 50)

// GetConn 通过ID获取当前链接
func GetConn(id string) *net.Conn {
	return devConnTable[id]
}

// servCh 客户端先将数据发送到servCh通道，然后由HandleServ处理发给服务器
var servCh = make(chan []byte, 100)

// handleDevMsg 设备消息处理函数
type handleDevMsg func(id string, action string)

// devHandle 不同类型设备消息处理函数映射表，key为设备类型
var devHandle = make(map[string]handleDevMsg, 10)

// InitDevHandle 初始化设备消息处理映射表
func InitDevHandle() {
	devHandle["塔吊"] = devices.TaDiaoHandleMsg
	devHandle["电表"] = devices.DianBiaoHandleMsg
	//devHandle["水表"] = devices.ShuiBiaoHandleMsg
	//devHandle["扬尘"] = devices.YangChenHandleMsg
}

// GetDevTypeByID 通过ID获取设备类型
func GetDevTypeByID(id string) (devType string) {
	t, _ := strconv.Atoi(id)
	switch t / 100 {
	case 50:
		devType = "塔吊"
	case 51:
		devType = "电表"
	case 52:
		devType = "水表"
	case 53:
		devType = "扬尘"

	}
	return devType
}

// InitLoger 初始化log配置
func InitLoger() error {
	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(file)
	return nil

}

func initConfig() error {
	serviceAddr = "192.168.10.112:7010"
	return nil
}

func handleBuff([]byte) []byte {
	return []byte(`{"type":{}}`)
}

// 处理每条json数据
func handleMsg(msg []byte) {
	jsonData := make(map[string]interface{})
	err := json.Unmarshal(msg, &jsonData)
	if err != nil {
		fmt.Printf("Json解析失败：%s\n", err.Error())
	}
	fmt.Println("jsonData:", jsonData)
	switch jsonData["MsgType"].(string) {
	case "Serv":
		fmt.Println(jsonData["MsgType"])
		switch jsonData["Action"].(string) {
		case "DevList":
			fmt.Println(jsonData["Action"])
			devList := make([]string, 0, 50)
			for _, id := range jsonData["Data"].([]interface{}) {
				fmt.Println(id)
				devList = append(devList, id.(string))
			}
			fmt.Println(devList)
			createDevTable(devList)
		}
	case "Devices":
		fmt.Println("case Devices")
		id := jsonData["ID"].(string)
		act := jsonData["Action"].(string)
		devHandle[GetDevTypeByID(id)](id, act)
	default:
		fmt.Println("default")
	}

}

// createDevTable 从服务器获取数据，创建设备列表
func createDevTable(devList []string) {
	for _, v := range devList {
		devConnTable[v] = nil
	}
	fmt.Printf("devConnTable:%v\n", devConnTable)
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
		//buff := make([]byte, 1024)
		//serviceConn.Read(buff)
		msg, err := bufio.NewReader(serviceConn).ReadSlice('\n')
		if err != nil {
			log.Fatal("获取服务器数据错误:", err)
		}
		fmt.Println("收到:", msg, "str:", string(msg))
		//msg = strings.TrimRight(msg, "\n")
		msg = msg[:len(msg)-1] //移除最后的'\n'
		//msg := handleBuff(buff)
		handleMsg(msg)
	}
}

// ReqDevList 向服务器请求设备列表
func ReqDevList() {
	servCh <- []byte(`{"MsgType":"Serv","Action":"DevList"}`)
}

//InitServerConn 初始化连接主服务器连接
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
func devCreateListen(port string) (net.Listener, error) {
	listen, err := net.Listen("tcp", "localhost:"+string(port))
	if err != nil {
		log.Printf("监听失败:%s,%s\n", "localhost:"+port, err.Error())
		return nil, err
	}
	return listen, nil
}

// devAcceptConn 等待设备连接，创建连接
func devAcceptConn(l net.Listener, port string) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			continue
		}
		devConnTable[port] = &conn
	}
}

// IntiDeviceConn 初始化设备链接
func IntiDeviceConn() (err error) {
	for port := range devConnTable {
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
	InitDevHandle()
	InitServerConn()
	IntiDeviceConn()
	for {

	}
}
