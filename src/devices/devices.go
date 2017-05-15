// package devices 主要处理和设备相关的通信

package devices

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const (
	offline = uint(0)
	online  = uint(1)
)

//Device 设备
type Device struct {
	hardwareID uint
	port       uint
	//hardwareCode string
	//devType      string
	conn  net.Conn
	state uint
}

// DevType 设备类型
type DevType struct {
	url     string
	devlist []uint
}

var devList = make(map[uint]*Device, 100) //设备列表

var devTypeTable = make(map[string][]uint, 10) //设备列表的类型索引，值为该类型的所有设备

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

func getDevType(dev string) (result string, err error) {
	devType := strings.Split(dev, "-")[0]
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
		err = errors.New("设备类型不存在")
	}
	return result, err
}
func initDevTypeTbale() {
	devTypeTable["电表"] = make([]uint, 0, 5)
	devTypeTable["水表"] = make([]uint, 0, 5)
	devTypeTable["塔吊"] = make([]uint, 0, 5)
	devTypeTable["污水"] = make([]uint, 0, 5)
	devTypeTable["扬尘"] = make([]uint, 0, 5)
	devTypeTable["噪音"] = make([]uint, 0, 5)
	devTypeTable["RFID"] = make([]uint, 0, 5)
	devTypeTable["电梯"] = make([]uint, 0, 5)
	devTypeTable["地磅"] = make([]uint, 0, 5)
	devTypeTable["摄像头"] = make([]uint, 0, 5)
}

func relayError(id string, errType string) {
	json := generateDataJSONStr(id, "ERROR", errType)
}

var urlTable = map[string]string{
	"设备列表": "",
	"电表":   ""}

// GetConn 通过ID获取当前链接
/*func getConn(id string) net.Conn {
	devConnTable := findDevConnTbale(id)
	return devConnTable[id]
}
*/
// BindConn 绑定连接到具体设备
func bindConn(id uint, conn net.Conn) {
	devList[id].conn = conn
	devList[id].state = online
}

// UnBindConn 解除设备的连接绑定
func unBindConn(id uint) {
	devList[id].conn = nil
	devList[id].state = offline
}

// reqDevList 向服务器请求设备列表
/*{ "code":200, "data":[ { "area":"生活区", "hardwareCode":"DIANBIAO-001", "hardwareId":1, "name":"智能电表", "port":10001 }, { "area":"施工区", "hardwareCode":"DIANBIAO-002", "hardwareId":2, "name":"智能电表", "port":10002 }, { "area":"大门", "hardwareCode":"RFID-001", "hardwareId":3, "name":"RFID读卡器", "port":10003 } ], "errMsg":"" } */

func reqDevList(url string) error {
	//sendServ([]byte(`{"MsgType":"Serv","Action":"DevList"}`))
	type jsonDev struct {
		area         string
		hardwareCode string
		hardwareID   uint
		name         string
		port         uint
	}
	type jsonDevList struct {
		code   int
		data   []jsonDev
		errMsg string
	}
	var reqDevListData jsonDevList
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("获取设备列表错误：%s\n", err.Error())
		return err
	}
	var content []byte
	defer resp.Body.Close()
	content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("获取设备列表内容读取失败：%s\n", err.Error())
		return err
	}
	err = json.Unmarshal(content, &reqDevListData)
	if err != nil {
		log.Printf("解析设备列表json数据失败：%s\n", err.Error())
		return err
	}
	if reqDevListData.code != 200 {
		log.Printf("服务器错误\n")
		err := errors.New("服务器错误")
		return err
	}
	for _, v := range reqDevListData.data {
		var dev = new(Device)
		devTypeStr, err := getDevType(v.hardwareCode)
		if err != nil {
			log.Printf("%s不存在的类型\n", v.hardwareCode)
			continue
		}
		//列表中不存在则加入列表
		if _, ok := devList[dev.hardwareID]; !ok {
			dev.port = v.port
			//dev.hardwareCode = v.hardwareCode
			dev.hardwareID = v.hardwareID
			dev.conn = nil
			dev.state = offline
			devList[dev.hardwareID] = dev
			devTypeTable[devTypeStr] = append(devTypeTable[devTypeStr], dev.hardwareID)

			//创建新的监听并等待连接
			port := strconv.FormatUint(uint64(dev.port), 10)
			listen, err := net.Listen("tcp", "localhost:"+port)
			if err != nil {
				log.Printf("监听失败:%s,%s\n", "localhost:"+port, err.Error())
				continue
			}
			if listen == nil {
				log.Println("listen == nil")
				continue
			}
			go devAcceptConn(listen, dev.hardwareID)
		}
	}
	fmt.Println(devList)
	return nil
}
func getConn(id uint) net.Conn {
	if _, ok := devList[id]; ok {
		return devList[id].conn
	}
	return nil
}

// devAcceptConn 等待设备连接，创建连接
func devAcceptConn(l net.Listener, hardwareID uint) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			continue
		}
		fmt.Printf("建立连接成功:%d\n", hardwareID)
		bindConn(hardwareID, conn)
	}
}

//把获取的设备数据分装到json中
func generateDataJSONStr(id string, action string, data string) string {
	str := fmt.Sprintf(`{"MsgType":"Devices","ID":"%s","Action":"%s","Data":"%s"}`, id, action, data)
	return str
}

type msg struct {
}
type msgData struct {
	id   uint
	data int64
}

func sendData(url string, jsonData []byte) error {
	dat := bytes.NewBuffer([]byte(jsonData))
	resp, err := http.Post(url, "application/json;charset=utf-8", dat)
	if err != nil {
		log.Printf("发送数据失败：%s\n", err.Error())
		return err
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取返回数据失败：%s\n", err.Error())
		return err
	}
	fmt.Printf("%s", result)
	return nil
}

// IntiDevice 初始化设备连接
func IntiDevice() error {
	initDevTypeTbale()
	reqDevList(urlTable["设备列表"])
	dianBiaoIntAutoGet()
	return nil
}
