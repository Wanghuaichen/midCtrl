// package devices 主要处理和设备相关的通信

package devices

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"errors"
	"net/http"
	"io/ioutil"
)


const (
	offline = uint(0)
)
// handleDevMsg 设备消息处理函数
type handleServMsgFunc func(id string, action string)

// devHandle 不同类型设备消息处理函数映射表，string 是key为设备类型
var servMsgHandleTable = make(map[string]handleServMsgFunc, 10)

// devConnTable 设备连接表,key为设备ID直接使用连接端口号做ID
//var devConnTable = make(map[string](net.Conn), 50)
type devConnTable map[string](net.Conn)



//设备类型
type Device struct{
	hardwareId uint
	port	   uint
	hardwareCode string
	conn	net.Conn
	handle  handleServMsgFunc
	state	uint
}
var devList = make([]Device,100,100)  //设备列表

//以port为索引，保存设备列表
var portTable = make(map[uint]uint)

// 第一个string为设备类型，值为在devlist中对应的索引
var devTypeTable = make(map[string][]uint, 10)
/*
TADIAO-001= 塔吊
YANGCHEN-001= 扬尘监测
DIANTI-001  =  电梯
RFID-001 = RFID识别器
SHUIBIA-O001 = 智能水表
DIANBIAO-001= 智能电表
WUSHUI-001= 污水监测
DIBANG-001 = 地磅
SHEXIANGTOU-001 = 摄像头
*/

func getDevType(dev string)(result string,err error){
	devType := strings.Split(dev,"-")[0]
	switch devType {
		case "TADIAO":
		result = "电表"
		case "SHUIBIA":
		result = "水表"
		case "TADIAO":
		result = "塔吊"
		case "WUSHUI":
		result = "污水"
		case "YANGCHEN":
		result = "扬尘"
		case "ZAOYIN":
		result = "噪音"
		case "RFID":
		result = "RFID"
		case "DIANTI":
		result = "电梯"
		case "DIBANG":
		result = "地磅"
		case "SHEXIANGTOU":
		result = "摄像头"
		default:
			err=errors.New("设备类型不存在")
	}
	return result, err
}
func initDevTypeTbale() {
	devTypeTable["电表"] = make([]uint, 0, 10)
	devTypeTable["水表"] = make([]uint, 0, 10)
	devTypeTable["塔吊"] = make([]uint, 0, 10)
	devTypeTable["污水"] = make([]uint, 0, 10)
	devTypeTable["扬尘"] = make([]uint, 0, 10)
	devTypeTable["噪音"] = make([]uint, 0, 10)
	devTypeTable["RFID"] = make([]uint, 0, 10)
	devTypeTable["电梯"] = make([]uint, 0, 10)
	devTypeTable["地磅"] = make([]uint, 0, 10)
	devTypeTable["摄像头"] = make([]uint, 0, 10)
}

// 返回id对应设备类型的所有连接表
/*func findDevConnTbale(id string) devConnTable {
	ID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("转化%s错误：%s\n", id, err)
		return nil
	}
	if ID >= 51000 && ID < 52000 {
		return devTypeTable["电表"]
	}
	if ID >= 52000 && ID < 53000 {
		return devTypeTable["水表"]
	}
	if ID >= 53000 && ID < 54000 {
		return devTypeTable["塔吊"]
	}
	if ID >= 54000 && ID < 55000 {
		return devTypeTable["污水"]
	}
	if ID >= 55000 && ID < 56000 {
		return devTypeTable["扬尘"]
	}
	if ID >= 56000 && ID < 57000 {
		return devTypeTable["噪音"]
	}
	if ID >= 57000 && ID < 58000 {
		return devTypeTable["RFID"]
	}
	return nil
}*/

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
func relayError(id string, errType string) {
	json := generateDataJsonStr(id, "ERROR", errType)
	sendServ([]byte(json))
}

// updateDevTable 从服务器获取数据，更新设备列表
/*
func updateDevTable(devList []string) {
	for _, id := range devList {
		devConnTable := findDevConnTbale(id)
		if devConnTable == nil {
			log.Printf("%s是一个无效的ID，它不存在于各设备\n", id)
			continue
		}
		_, ok := devConnTable[id]
		if !ok {
			devConnTable[id] = nil
			listen, err := net.Listen("tcp", "localhost:"+id)
			if err != nil {
				log.Printf("监听失败:%s,%s\n", "localhost:"+id, err.Error())
				continue
			}
			if listen == nil {
				log.Println("listen == nil")
				continue
			}
			go devAcceptConn(listen, id)
		}
	}
	fmt.Printf("devConnTable:%v\n", devTypeTable)
}
*/
// GetConn 通过ID获取当前链接
/*func getConn(id string) net.Conn {
	devConnTable := findDevConnTbale(id)
	return devConnTable[id]
}
*/
// BindConn 绑定连接到具体设备
func bindConn(id string, conn net.Conn) {
	devConnTable := findDevConnTbale(id)
	devConnTable[id] = conn
}

// UnBindConn 解除设备的连接绑定
func unBindConn(id string) {
	devConnTable := findDevConnTbale(id)
	devConnTable[id] = nil
}

// reqDevList 向服务器请求设备列表
/*{ "code":200, "data":[ { "area":"生活区", "hardwareCode":"DIANBIAO-001", "hardwareId":1, "name":"智能电表", "port":10001 }, { "area":"施工区", "hardwareCode":"DIANBIAO-002", "hardwareId":2, "name":"智能电表", "port":10002 }, { "area":"大门", "hardwareCode":"RFID-001", "hardwareId":3, "name":"RFID读卡器", "port":10003 } ], "errMsg":"" } */

func reqDevList(url string)error {
	//sendServ([]byte(`{"MsgType":"Serv","Action":"DevList"}`))
	type jsonDev struct{
		area string
		hardwareCode string
		hardwareId uint
		name string
		port uint
	}
	type jsonDevList struct{
		code int
		data []jsonDev
		errMsg string
	}
	var reqDevListData jsonDevList
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("获取设备列表错误：%s\n",err.Error())
		return err
	}
	var content []byte
	defer resp.Body.Close()
	content, err =  ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("获取设备列表内容读取失败：%s\n",err.Error())
		return err
	}
	err = json.Unmarshal(content,&reqDevListData)
	if err!=nil {
		log.Printf("解析设备列表json数据失败：%s\n",err.Error())
		return err
	}
	if reqDevListData.code != 200 {
		log.Printf("服务器错误\n")
		err := errors.New("服务器错误")
		return err
	}
	for _,v:= range reqDevListData.data {
		var dev Device
		dev.port = v.port
		dev.hardwareCode = v.hardwareCode
		dev.hardwareId = v.hardwareId
		dev.conn = nil
		dev.handle =getDevHandle(v.hardwareCode)
		dev.state = offline
		devList = append(devList,dev)
	}
	fmt.Println(devList)
	//fmt.Println(string(content))

}

// initDevHandle 初始化设备消息处理映射表
func initDevHandle() {
	servMsgHandleTable["塔吊"] = TaDiaoHandleMsg
	servMsgHandleTable["电表"] = DianBiaoHandleMsg
	//devHandle["水表"] = devices.ShuiBiaoHandleMsg
	//devHandle["扬尘"] = devices.YangChenHandleMsg
}
func getDevHandle(hardwareCode string)handleServMsgFunc{
	typeStr, err:= getDevType(hardwareCode)
	if err!=nil{
		log.Println(err.Error())
	}
	return servMsgHandleTable[typeStr]
}
// HandleDevMsg 处理设备发上来的数据
func handleDevMsg(id string, conn net.Conn) {
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Println("读设备数据错误:", err.Error())
		}
		//处理后发送给主服务器
		fmt.Printf("收到 %s：%v %s\n", id, buff[:n], string(buff[:n]))
		sendServ(buff[:n])
	}
}

// getDevTypeByID 通过ID获取设备类型
func getDevTypeByID(id string) (devType string) {
	t, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("数字转化失败：%s\n", err.Error())
		return ""
	}
	//fmt.Printf("%s-->%d t/100=%v\n", id, t, t/100)
	switch t / 1000 {
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

// devAcceptConn 等待设备连接，创建连接
func devAcceptConn(l net.Listener, port string) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			continue
		}
		fmt.Printf("建立连接成功:%s\n", port)
		bindConn(port, conn)
		//handleDevMsg(port, conn)
	}
}

func checkID(id string) bool {
	if id == "" {
		return false
	}
	return true
}

// 处理每条json数据
func handleMsg() {
	for {
		msg := getServ()
		jsonData := make(map[string]interface{})
		err := json.Unmarshal(msg, &jsonData)
		if err != nil {
			fmt.Printf("Json解析失败：%s\n", err.Error())
			continue
		}
		//fmt.Println("jsonData:", jsonData)
		if jsonData == nil {
			log.Println("json数据为空")
			continue
		}
		switch jsonData["MsgType"].(string) {
		case "Serv":
			fmt.Println(jsonData["MsgType"])
			switch jsonData["Action"].(string) {
			case "DevList":
				//fmt.Println(jsonData["Action"])
				devList := make([]string, 0, 50)
				for _, id := range jsonData["Data"].([]interface{}) {
					//fmt.Println(id)  //后期需要判断ID是否符合要求
					if !checkID(id.(string)) {
						continue
					}
					devList = append(devList, id.(string))
				}
				//fmt.Println(devList)
				updateDevTable(devList)
			}
		case "Devices":
			//fmt.Println("case Devices")
			id := jsonData["ID"].(string)
			act := jsonData["Action"].(string)
			//servMsgHandleTable[getDevTypeByID(id)](id, act)
			handle := servMsgHandleTable[getDevTypeByID(id)]
			if handle == nil {
				log.Printf("获取处理函数错误：%s\n", getDevTypeByID(id))
			} else {
				handle(id, act)
			}
			//HandleServMsg(msg)
		default:
			fmt.Println("default")
		}
	}

}

//把获取的设备数据分装到json中
func generateDataJsonStr(id string, action string, data string) string {
	str := fmt.Sprintf(`{"MsgType":"Devices","ID":"%s","Action":"%s","Data":"%s"}`, id, action, data)
	return str
}

// IntiDevice 初始化设备连接
func IntiDevice() error {
	initDevHandle()
	initDevTypeTbale()
	reqDevList()
	go handleMsg() //处理来自服务器的消息
	dianBiaoIntAutoGet()
	return nil
}
