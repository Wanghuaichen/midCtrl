// package devices 主要处理和设备相关的通信

package devices

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

// handleDevMsg 设备消息处理函数
type handleServMsgFunc func(id string, action string)

// devHandle 不同类型设备消息处理函数映射表，key为设备类型
var servMsgHandleTable = make(map[string]handleServMsgFunc, 10)

// devConnTable 设备连接表,key为设备ID直接使用连接端口号做ID
var devConnTable = make(map[string](*net.Conn), 50)

// toServCh 所有要发给主服务器的数据先发送到toServCh通道，由中间层转发
var toServCh = make(chan []byte, 100)

// sendServ 发送数据给主服务器通道
func sendServ(data []byte) {
	toServCh <- data
}

// GetDate 获取devicse要发的数据
func GetData() []byte {
	return <-toServCh
}

// 主服务器发送来的数据先缓存到这里等待处理
var formServCh = make(chan []byte, 100)

// SendDate 将主服务器的数据放入devices通道
func SendData(data []byte) {
	formServCh <- data
}

// getServ 取出服务器发送过来的数据
func getServ() []byte {
	return <-formServCh
}

// updateDevTable 从服务器获取数据，更新设备列表
func updateDevTable(devList []string) {
	for _, id := range devList {
		_, ok := devConnTable[id]
		if !ok {
			devConnTable[id] = nil
			listen, err := net.Listen("tcp", "localhost:"+id)
			if err != nil {
				log.Printf("监听失败:%s,%s\n", "localhost:"+id, err.Error())
			}
			go devAcceptConn(listen, id)
		}
	}
	fmt.Printf("devConnTable:%v\n", devConnTable)
}

/*func getDevConnTbale() map[string](*net.Conn) {
	return devConnTable
}*/

// GetConn 通过ID获取当前链接
func getConn(id string) *net.Conn {
	return devConnTable[id]
}

// BindConn 绑定连接到具体设备
func bindConn(id string, conn *net.Conn) {
	devConnTable[id] = conn
}

// UnBindConn 解除设备的连接绑定
func unBindConn(id string) {
	devConnTable[id] = nil
}

// reqDevList 向服务器请求设备列表
func reqDevList() {
	sendServ([]byte(`{"MsgType":"Serv","Action":"DevList"}`))
}

// initDevHandle 初始化设备消息处理映射表
func initDevHandle() {
	servMsgHandleTable["塔吊"] = TaDiaoHandleMsg
	servMsgHandleTable["电表"] = DianBiaoHandleMsg
	//devHandle["水表"] = devices.ShuiBiaoHandleMsg
	//devHandle["扬尘"] = devices.YangChenHandleMsg
}

// HandleServMsg 处理服务器要操作设备的消息
/*func HandleServMsg(msg []byte) {
	jsonData := make(map[string]interface{})
	err := json.Unmarshal(msg, &jsonData)
	if err != nil {
		fmt.Printf("Json解析失败：%s\n", err.Error())
	}

}*/

// HandleDevMsg 处理设备发上来的数据
func handleDevMsg(id string, conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalln("读设备数据错误:", err.Error())
		}
		//处理后发送给主服务器
		sendServ(buff[:n])
	}
}

// getDevTypeByID 通过ID获取设备类型
func getDevTypeByID(id string) (devType string) {
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
	default:
		log.Printf("%s不能识别设备类型\n", id)
	}
	return devType
}

// devCreateListen 使用设备ID即固定端口号创建监听设备
/*func devCreateListen(port string) (net.Listener, error) {
	listen, err := net.Listen("tcp", "localhost:"+string(port))
	if err != nil {
		log.Printf("监听失败:%s,%s\n", "localhost:"+port, err.Error())
		return nil, err
	}
	return listen, nil
}
*/
// devAcceptConn 等待设备连接，创建连接
func devAcceptConn(l net.Listener, port string) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			continue
		}
		fmt.Printf("建立连接成功:%s\n", port)
		bindConn(port, &conn)
		handleDevMsg(port, conn)
	}
}

// IntiDeviceConn 初始化设备链接
/*func IntiDeviceConn() (err error) {
	for port := range devices.GetDevConnTbale() {
		l, err := devCreateListen(port)
		if err == nil {
			go devAcceptConn(l, port)
		} else {
			log.Printf("监听[%s]建立连接错误:%s\n", port, err.Error())
		}
	}
	return err

}
*/
// 处理每条json数据
func handleMsg() {
	for {
		msg := getServ()
		jsonData := make(map[string]interface{})
		err := json.Unmarshal(msg, &jsonData)
		if err != nil {
			fmt.Printf("Json解析失败：%s\n", err.Error())
		}
		//fmt.Println("jsonData:", jsonData)
		switch jsonData["MsgType"].(string) {
		case "Serv":
			fmt.Println(jsonData["MsgType"])
			switch jsonData["Action"].(string) {
			case "DevList":
				//fmt.Println(jsonData["Action"])
				devList := make([]string, 0, 50)
				for _, id := range jsonData["Data"].([]interface{}) {
					//fmt.Println(id)
					devList = append(devList, id.(string))
				}
				//fmt.Println(devList)
				updateDevTable(devList)
			}
		case "Devices":
			//fmt.Println("case Devices")
			id := jsonData["ID"].(string)
			act := jsonData["Action"].(string)
			servMsgHandleTable[getDevTypeByID(id)](id, act)
			//HandleServMsg(msg)
		default:
			fmt.Println("default")
		}
	}

}

func IntiDevice() error {
	initDevHandle()
	reqDevList()
	go handleMsg() //处理来自服务器的消息
	return nil
}

