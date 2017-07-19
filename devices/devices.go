// package devices 主要处理和设备相关的通信

package devices

/*
{200 [{工地大门-入口 RFID-001 1 RFID识别器1 10101}
{供电房 DIANBIAO-001 2 智能电表 10201}
{塔楼核心筒 DIANTI-001 3 电梯 10301}
{车棚旁绿化带 SHUIBIAO-001 4 智能水表 10401}
{工地南门 DIBANG-001 5 地磅 10501}
{工地污水沉降池 WUSHUI-001 6 污水监测 10601}
{工地大门东侧 ENV-001 7 环境监测1 10701}
{塔楼核心筒 TADIAO-001 8 塔吊 10801}
{工地大门－出口 RFID-002 9 RFID识别器2 10102}
{顶模电梯1 RFID-003 10 RFID识别器3 10103}
{顶模电梯2 RFID-004 11 RFID识别器4 10104}
{废料回收处 ZNDIBANG-001 12 智能地磅 11201}
{顶模 DIANTI-002 13 电梯2 10302}
{工地西区 WUSHUI-002 14 污水监测2 10602}
{工地大门入口 ENV-002 15 环境监测2 10702}
{喷淋区 PL-001 16 喷淋 11601}] }
*/
import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"midCtrl/comm"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	offline = uint(0)
	online  = uint(1)
	noData  = uint(2)
)

//Device 设备
type Device struct {
	hardwareID   uint
	port         uint
	hardwareCode string
	//devType      string
	conn  net.Conn
	state uint     //当前状态 1 正常  0 网络断开 2 设备不能返回数据
	isOk  uint     //命令执行结果
	cmd   chan int //从服务器返回命令给具体设备执行
}

// DevType 设备类型
type DevType struct {
	url     string
	devlist []uint
}

var devList = make(map[uint]*Device, 100) //设备列表

var devTypeTable = make(map[string][]uint, 10)            //设备列表的类型索引，值为该类型的所有设备
var reqDevListTicker = time.NewTicker(time.Minute * 2)    //请求列表周期
var reportStatusTicker = time.NewTicker(time.Second * 10) //状态上报周期

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
	case "DIANBIAO":
		result = "电表"
	case "SHUIBIAO":
		result = "水表"
	case "TADIAO":
		result = "塔吊"
	case "WUSHUI":
		result = "污水"
	case "ENV":
		result = "环境"
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
	case "ZNDIBANG":
		result = "智能地磅"
	case "PL":
		result = "喷淋"
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
	devTypeTable["环境"] = make([]uint, 0, 5)
	devTypeTable["RFID"] = make([]uint, 0, 5)
	devTypeTable["电梯"] = make([]uint, 0, 5)
	devTypeTable["地磅"] = make([]uint, 0, 5)
	devTypeTable["智能地磅"] = make([]uint, 0, 5)
	devTypeTable["摄像头"] = make([]uint, 0, 5)
	devTypeTable["喷淋"] = make([]uint, 0, 5)
}

func relayError(id string, errType string) {
	//json := generateDataJSONStr(id, "ERROR", errType)
	//sendData(urlTable["错误"], id, []byte(json))
}

var urlTable = map[string]string{
	"电表":   "http://39.108.5.184/smart/api/saveElectricityData",
	"水表":   "http://39.108.5.184/smart/api/saveWaterData",
	"塔吊":   "http://39.108.5.184/smart/api/saveCraneData",
	"污水":   "http://39.108.5.184/smart/api/savePhData",
	"环境":   "http://39.108.5.184/smart/api/saveEnvData",
	"RFID": "http://39.108.5.184/smart/api/checkIn",
	"电梯":   "http://39.108.5.184/smart/api/saveElevatorData",
	"地磅":   "http://39.108.5.184/smart/api/saveWeighbridgeData",
	"智能地磅": "http://39.108.5.184/smart/api/saveWeighbridgeData",
	"摄像头":  "",
	"设备列表": "http://39.108.5.184/smart/api/getHardwareList?projectId=1",
	"设备状态": "http://39.108.5.184/smart/api/reportState"}

// GetURL 获取要发消息的url
func GetURL(urlStr string) (url string) {
	return urlTable[urlStr]
}

// GetConn 通过ID获取当前链接
/*func getConn(id string) net.Conn {
	devConnTable := findDevConnTbale(id)
	return devConnTable[id]
}
*/
// BindConn 绑定连接到具体设备 并设定状态为上线
func bindConn(id uint, conn net.Conn) {
	if devList[id].conn != nil {
		log.Printf("非正常关闭连接:%d %v\n", id, conn)
		devList[id].conn.Close()
	}
	devList[id].conn = conn
	devList[id].state = online
}

// UnBindConn 解除设备的连接绑定 并设定状态为断开
func unBindConn(id uint) {
	if devList[id].conn != nil {
		log.Printf("解除绑定关闭连接:%d %v\n", id, devList[id].conn)
		devList[id].conn.Close()
		devList[id].conn = nil
	}

	devList[id].state = offline
}

func setStateNoData(id uint) {
	if devList[id].state == online {
		devList[id].state = noData
	}
}
func setStateOk(id uint) {
	if devList[id].state == noData {
		devList[id].state = online
	}
}

// reqDevList 向服务器请求设备列表
/*{ "code":200, "data":[ { "area":"生活区", "hardwareCode":"DIANBIAO-001", "hardwareId":1, "name":"智能电表", "port":10001 }, { "area":"施工区", "hardwareCode":"DIANBIAO-002", "hardwareId":2, "name":"智能电表", "port":10002 }, { "area":"大门", "hardwareCode":"RFID-001", "hardwareId":3, "name":"RFID读卡器", "port":10003 } ], "errMsg":"" } */

func reqDevList(url string) error {
	defer func(){
		if err:=recover();err!=nil{
			fmt.Printf("获取设备列表发生Panic错误：%s\n", err.Error())
		}
	}
	//sendServ([]byte(`{"MsgType":"Serv","Action":"DevList"}`))
	//fmt.Printf("reqDevList start\n")
	type jsonDev struct {
		Area         string `json:"area"`
		HardwareCode string `json:"hardwareCode"`
		HardwareID   uint   `json:"hardwareId"`
		Name         string `json:"name"`
		Port         uint   `json:"port"`
	}
	type jsonDevList struct {
		Code   int       `json:"code"`
		Data   []jsonDev `json:"data"`
		ErrMsg string    `json:"errMsg"`
	}
	var reqDevListData jsonDevList
	//http://39.108.5.184/smart/api/getHardwareList?projectId=1

	/*client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*2)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 2))
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 2,
		},
	}*/
	//fmt.Printf("reqDevList HTTP GET\n")
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("获取设备列表错误：%s\n", err.Error())
		return err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("获取设备列表内容读取失败：%s\n", err.Error())
		return err
	}
	//fmt.Printf("获取设备列表内容:n%s\n", string(content))
	err = json.Unmarshal(content, &reqDevListData)
	if err != nil {
		log.Printf("解析设备列表json数据失败：%s\n", err.Error())
		return err
	}

	log.Println(reqDevListData)
	if reqDevListData.Code != 200 {
		log.Printf("服务器错误\n")
		err := errors.New("服务器错误")
		return err
	}
	for _, v := range reqDevListData.Data {
		var dev = new(Device)
		devTypeStr, err := getDevType(v.HardwareCode)
		if err != nil {
			log.Printf("%s不存在的类型\n", v.HardwareCode)
			continue
		}
		if v.HardwareID == 3 || v.HardwareID == 13 || v.HardwareID == 8 || v.HardwareID == 15 {
			continue //8塔吊  3、13 电梯 15 环境2 是通过网页获取，不需要检测
		}
		//列表中不存在则加入列表
		if _, ok := devList[v.HardwareID]; !ok {
			dev.port = v.Port
			dev.hardwareCode = v.HardwareCode
			dev.hardwareID = v.HardwareID
			dev.conn = nil
			dev.state = offline
			dev.isOk = 1
			dev.cmd = make(chan int, 3)
			devList[dev.hardwareID] = dev
			devTypeTable[devTypeStr] = append(devTypeTable[devTypeStr], dev.hardwareID)

			//创建新的监听并等待连接
			port := strconv.FormatUint(uint64(dev.port), 10)
			listen, err := net.Listen("tcp", "0.0.0.0:"+port)
			if err != nil {
				log.Printf("监听失败:%s,%s\n", "0.0.0.0:"+port, err.Error())
				listen.Close()
				continue
			}
			if listen == nil {
				log.Println("listen == nil")
				continue
			}
			log.Printf("监听 【%s】成功\n", port)
			go devAcceptConn(listen, dev.hardwareID)
		}
	}
	//fmt.Println(devList)
	return nil
}
func getConn(id uint) net.Conn {
	if _, ok := devList[id]; ok {
		return devList[id].conn
	}
	fmt.Printf("id 不存在,%d %d\n", id, devList[id].conn)
	return nil
}

// devAcceptConn 等待设备连接，创建连接
func devAcceptConn(l net.Listener, hardwareID uint) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			l.Close()
			continue
		}
		//fmt.Printf("建立连接成功:%d %v\n", hardwareID, conn)
		bindConn(hardwareID, conn)
		devTypeStr, _ := getDevType(devList[hardwareID].hardwareCode)
		log.Printf("创建连接%d--->%s\n", hardwareID, devTypeStr)
		switch devTypeStr {
		case "塔吊":
			//taDiaoStart(hardwareID)
		case "地磅":
			go diBangD39Start(hardwareID)
		case "智能地磅":
			go znDiBangStart(hardwareID)
		case "RFID":
			go rfidStart(hardwareID)
		case "环境":
			go huanJingStart(hardwareID)
		case "污水":
			go wuShuiStart(hardwareID)
		case "水表":
			go shuiBiaoStart(hardwareID)
		case "电表":
			dianBiaoStart(hardwareID)
		case "喷淋":
			penLinStart(hardwareID)

		}
	}
}

//把获取的设备数据分装到json中
func generateDataJSONStr(id string, action string, data string) string {
	str := fmt.Sprintf(`{"MsgType":"Devices","ID":"%s","Action":"%s","Data":"%s"}`, id, action, data)
	return str
}

func reportDevStatus() {
	for _, dev := range devList {
		devState := make(url.Values)
		devState["isOk"] = []string{strconv.FormatInt(int64(dev.isOk), 10)}
		devState["state"] = []string{strconv.FormatInt(int64(dev.state), 10)}
		sendData("设备状态", dev.hardwareID, devState)
	}
}
func sendData(urlStr string, id uint, data url.Values) {
	var msg comm.MsgData
	msg.SetTime()
	msg.HdID = id
	msg.Data = data
	msg.URLStr = urlStr
	// if urlStr != "设备状态" {
	// 	log.Printf("发送%s:%v\n", urlStr, msg)
	// }
	log.Printf("发送%s:%v\n", urlStr, msg)
	comm.SendMsg(msg)
}

// handleServCmd 处理服务器返回的命令
func handleServCmd() {
	timeout := time.NewTimer(time.Second * 5)
	for {
		servCmd := comm.GetCmd()
		//log.Printf("执行命令：%v\n", servCmd)
		timeout.Reset(time.Second * 5)
		if devList[servCmd.HdID].state == 1 { //只有设备在线才发给设备
			select {
			case devList[servCmd.HdID].cmd <- servCmd.Cmd:
				break
			case <-timeout.C:
				log.Printf("发送执行命令超时，对应协程可能异常：%v\n", servCmd)
			}
		}
	}
}

// IntiDevice 初始化设备连接
func IntiDevice() error {
	initDevTypeTbale()
	//定时请求设备列表
	go func() {
		reqDevList(urlTable["设备列表"])
		for _ = range reqDevListTicker.C {
			reqDevList(urlTable["设备列表"])
		}
	}()
	//定时上报状态
	go func() {
		for _ = range reportStatusTicker.C {
			reportDevStatus()
		}
	}()
	go handleServCmd()
	return nil
}
