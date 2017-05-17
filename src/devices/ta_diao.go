package devices

import (
	"bytes"
	"encoding/binary"
	"log"
	"time"
)

/*
	1.? 实时数据和工作数据内容一样
*/
// 包头20个字节
type packHeadType struct {
	mark    int8
	opType  int8
	len     uint16
	sn      [12]uint8
	ver     uint8
	cmd     uint8
	headCRC uint16
}

//操作记录时间 8个字节
type opTimeType struct {
	year  uint16
	month int8
	day   int8
	week  int8
	hour  uint8
	min   uint8
	sec   uint8
}

func (t opTimeType) toTimestamp() int64 {
	timestamp := time.Date(int(t.year), time.Month(t.month), int(t.day), int(t.hour), int(t.min), int(t.sec), 0, time.UTC).Unix()
	return timestamp
}
func (t *opTimeType) setTime() {
	tm := time.Now()
	t.year = uint16(tm.Year())
	t.month = int8(tm.Month())
	t.day = int8(tm.Day())
	t.week = int8(tm.Weekday())
	t.hour = uint8(tm.Hour())
	t.min = uint8(tm.Minute())
	t.sec = uint8(tm.Second())
}

// 实时数据 60个字节
type craneDataRecordType struct {
	recordTime  opTimeType //时间
	angle       float32    //转角
	radius      float32    //幅度
	height      float32    //高度
	load        float32    //吊重
	safeload    float32    //安全吊重
	percent     float32    //力矩百分比
	windspeed   float32    //风速
	obliquity   float32    //塔机倾角
	dirAngle    float32    //倾角方向，角度表示
	fall        uint16     //吊绳倍率
	outControl  uint16     //系统输出控制码状态
	earlyAlarm  uint32     //系统预警状态码
	alarm       uint32     //系统报警状态码
	peccancy    uint16     //违章造作状态码
	sensorAlarm uint16     //传感器报警状态码
}

// 将数据转为发送给服务器的形式
func (cdr craneDataRecordType) toServData() []map[string]interface{} {
	var result = make([]map[string]interface{}, 0, 16)
	result = append(result, map[string]interface{}{"记录时间": cdr.recordTime.toTimestamp()})
	result = append(result, map[string]interface{}{"转角": int64(cdr.angle * 10000)})
	result = append(result, map[string]interface{}{"幅度": int64(cdr.radius * 10000)})
	result = append(result, map[string]interface{}{"高度": int64(cdr.height * 10000)})
	result = append(result, map[string]interface{}{"吊重": int64(cdr.load * 10000)})
	result = append(result, map[string]interface{}{"安全吊重": int64(cdr.safeload * 10000)})
	result = append(result, map[string]interface{}{"力矩百分比": int64(cdr.percent * 10000)})
	result = append(result, map[string]interface{}{"风速": int64(cdr.windspeed * 10000)})
	result = append(result, map[string]interface{}{"塔机倾角": int64(cdr.obliquity * 10000)})
	result = append(result, map[string]interface{}{"倾角方向": int64(cdr.dirAngle * 10000)})
	result = append(result, map[string]interface{}{"吊绳倍率": int64(cdr.fall * 10000)})
	result = append(result, map[string]interface{}{"控制码": int64(cdr.outControl * 10000)})
	result = append(result, map[string]interface{}{"预警码": int64(cdr.earlyAlarm * 10000)})
	result = append(result, map[string]interface{}{"报警码": int64(cdr.alarm * 10000)})
	result = append(result, map[string]interface{}{"违章码": int64(cdr.peccancy * 10000)})
	result = append(result, map[string]interface{}{"违章码": int64(cdr.sensorAlarm * 10000)})
	return result
}

type realtiemRecordType struct {
	num uint32              //记录号
	cdr craneDataRecordType //具体记录
}

type runTimeDataType struct {
	recordTime opTimeType //当前记录的时间
	runSecond  uint32     //运行的秒数
}

func (rt runTimeDataType) toServData() []map[string]interface{} {
	var result = make([]map[string]interface{}, 0, 2)
	result = append(result, map[string]interface{}{"记录时间": rt.recordTime.toTimestamp()})
	result = append(result, map[string]interface{}{"转角": int64(rt.runSecond * 10000)})
	return result
}

type runTimeDataRecordType struct {
	num uint32
	rtd runTimeDataType
}

var taDiaoSendDataCh = make(chan []byte, 5)

func taDiaoSend(dat []byte) {
	taDiaoSendDataCh <- dat
}

func taDiaoWriteData(id uint) {
	conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		return
	}
	for {
		dat := <-taDiaoSendDataCh
		_, err := conn.Write(dat)
		if err != nil {
			log.Printf("写数据失败: ")
			unBindConn(id)
			return
		}
		time.Sleep(time.Millisecond * 100)
	}
}

//回复心跳包
func replyHeardBeat(dat []byte) {
	taDiaoSend(dat)
}
func taDiaoReplyMsg(pHead []byte, num uint32) {
	buff := make([]byte, 0, 32)
	err := binary.Write(bytes.NewBuffer(buff), binary.LittleEndian, num)
	if err != nil {
		log.Printf("转化数据错误：%s\n", err.Error())
	}
	low, high := tableCRC16(buff)
	buff = append(buff, low, high)
	relayDat := mergeSlice(pHead, buff)
	taDiaoSend(relayDat)
}

// 获取实时数据
func handleRealData(id uint, pHead []byte, dat []byte) {
	var realData realtiemRecordType
	if !tableCheckCRC(dat) {
		return
	}
	//去除CRC数据转换到数据接口变量中
	err := binary.Read(bytes.NewReader(dat[:len(dat)-2]), binary.LittleEndian, realData)
	if err != nil {
		log.Printf("数据转换失败：%s\n", err.Error())
	}
	sendData(urlTable["塔吊"], id, realData.cdr.toServData())
	taDiaoReplyMsg(pHead, realData.num)
}

// 处理开机时间
func handleStartTime(id uint, pHead []byte, dat []byte) {
	var st runTimeDataRecordType
	if !tableCheckCRC(dat) {
		return
	}
	//去除CRC数据转换到数据接口变量中
	err := binary.Read(bytes.NewReader(dat[:len(dat)-2]), binary.LittleEndian, st)
	if err != nil {
		log.Printf("数据转换失败：%s\n", err.Error())
	}
	sendData(urlTable["塔吊"], id, st.rtd.toServData())
	taDiaoReplyMsg(pHead, st.num)
}

// 处理时间同步
func handleTimeSyn(pHead []byte) {
	var tm opTimeType
	tm.setTime()
	buff := make([]byte, 0, 128)
	err := binary.Write(bytes.NewBuffer(buff), binary.LittleEndian, tm)
	if err != nil {
		log.Printf("转化数据错误：%s\n", err.Error())
	}
	low, high := tableCRC16(buff)
	buff = append(buff, low, high)
	dat := mergeSlice(pHead, buff)
	taDiaoSend(dat)
}

func mergeSlice(b1, b2 []byte) []byte {
	s := make([]byte, 0, len(b1)+len(b2))
	for _, v := range b1 {
		s = append(s, v)
	}
	for _, v := range b2 {
		s = append(s, v)
	}
	return s
}

// 处理塔吊上传的数据
func taDiaoDataHandle(id uint) {
	conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		return
	}
	buf := make([]byte, 0, 1024)
	len := 0  //buf中最后数据位置
	sIdx := 0 //buf中开始数据位置
	var packHead packHeadType
GO_ON_READ:
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("读数据错误：%s\n", err.Error())
		}
		len = len + n
		//buff := bytes.NewBuffer(buf)
		for {
			sIndex := bytes.IndexByte(buf[:len], byte('$')) //帧开始位置
			if sIndex == -1 {
				log.Printf("数据中未找到包起始符号:%v\n", buf[:len])
				len = 0
				sIdx = 0
				continue
			}
			packHeadData := buf[sIndex : sIndex+20]

			if tableCheckCRC(packHeadData) { //CRC校验
				err := binary.Read(bytes.NewReader(packHeadData), binary.LittleEndian, packHead)
				if err != nil {
					log.Printf("数据转换失败：%s\n", err.Error())
					unBindConn(id)
					return
				}
			} else { //包头无效，继续找下一个数据包
				sIdx++
				continue
			}
			if packHead.len >= uint16(len-sIdx-20) { //之后的数据不完整，重新再读,包头长度20
				continue GO_ON_READ
			}
			switch packHead.cmd {
			case 0x0: //心跳
				replyHeardBeat(packHeadData)
			case 0x1: //实时数据
				handleRealData(id, packHeadData, buf[int(sIndex)+20:int(sIndex)+20+int(packHead.len)])
			case 0x2: //工作工作数据
				handleRealData(id, packHeadData, buf[int(sIndex)+20:int(sIndex)+20+int(packHead.len)])
			case 0x3: //开机运行记录
				handleStartTime(id, packHeadData, buf[int(sIndex)+20:int(sIndex)+20+int(packHead.len)])
			case 0x4: //时间校对
				handleTimeSyn(packHeadData)
			case 0xA: //读取设置心跳包周期（服务器主动）
			case 0xB: //读取设置实时数据周期（服务器主动）
			case 0x14: //获取结构参数（服务器主动）
			case 0x1E: //获取限制区域数据（服务器主动）
			case 0xCD: //获取GPRS信息（服务器主动）
			default:
				log.Printf("命令参数错误\n")
			}
			sIdx = sIdx + 20 + int(packHead.len)
			//buf清除 从头开始
			if sIdx == len {
				len = 0
				sIdx = 0
			}
		}
	}

}

func taDiaoStart(id uint) {
	go taDiaoDataHandle(id)
	go taDiaoWriteData(id)
}
