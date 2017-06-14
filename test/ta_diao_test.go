package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strconv"
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
func (cdr craneDataRecordType) toServData() map[string][]string {
	var result = make(map[string][]string, 17)
	result["angle"] = []string{strconv.FormatInt(int64(cdr.angle*10000), 10)}
	result["radius"] = []string{strconv.FormatInt(int64(cdr.radius*10000), 10)}
	result["height"] = []string{strconv.FormatInt(int64(cdr.height*10000), 10)}
	result["load"] = []string{strconv.FormatInt(int64(cdr.load*10000), 10)}
	result["safeload"] = []string{strconv.FormatInt(int64(cdr.safeload*10000), 10)}
	result["percent"] = []string{strconv.FormatInt(int64(cdr.percent*10000), 10)}
	result["windspeed"] = []string{strconv.FormatInt(int64(cdr.windspeed*10000), 10)}
	result["obliquity"] = []string{strconv.FormatInt(int64(cdr.obliquity*10000), 10)}
	result["dirAngle"] = []string{strconv.FormatInt(int64(cdr.dirAngle*10000), 10)}
	result["fall"] = []string{strconv.FormatInt(int64(cdr.fall*10000), 10)}
	result["outControl"] = []string{strconv.FormatInt(int64(cdr.outControl*10000), 10)}
	result["earlyAlarm"] = []string{strconv.FormatInt(int64(cdr.earlyAlarm*10000), 10)}
	result["alarm"] = []string{strconv.FormatInt(int64(cdr.alarm*10000), 10)}
	result["peccancy"] = []string{strconv.FormatInt(int64(cdr.peccancy*10000), 10)}
	result["sensorAlarm"] = []string{strconv.FormatInt(int64(cdr.sensorAlarm*10000), 10)}
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

func (rt runTimeDataType) toServData() map[string][]string {
	var result = make(map[string][]string, 3)
	result["runSecond"] = []string{strconv.FormatInt(int64(rt.runSecond*10000), 10)}
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
	//conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		return
	}
	for {
		dat := <-taDiaoSendDataCh
		_, err := conn.Write(dat)
		if err != nil {
			log.Printf("写数据失败: ")
			//unBindConn(id)
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
	//sendData("塔吊", id, realData.cdr.toServData())
	log.Println(realData.cdr.toServData())
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
	//sendData("塔吊", id, st.rtd.toServData())
	log.Println(st.rtd.toServData())
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
	//conn := getConn(id)
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
				continue GO_ON_READ
			}
			packHeadData := buf[sIndex : sIndex+20]

			if tableCheckCRC(packHeadData) { //CRC校验
				err := binary.Read(bytes.NewReader(packHeadData), binary.LittleEndian, packHead)
				if err != nil {
					log.Printf("数据转换失败：%s\n", err.Error())
					//unBindConn(id)
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

var conn net.Conn

func initConn() {
	l, err := net.Listen("tcp", "localhost:10101")
	if err != nil {
		log.Println(err.Error())
		return
	}
	conn, err = l.Accept()
	if err != nil {
		log.Println(err.Error())
		return
	}
}

func taDiaoStart(id uint) {
	go taDiaoDataHandle(id)
	go taDiaoWriteData(id)
	log.Println("塔吊数据获取开始")
}
func crc16Modbus(data []byte) (low byte, high byte) {
	sum := uint16(0xFFFF)
	for _, v := range data {
		sum ^= uint16(v)
		for i := 0; i < 8; i++ {
			if 0 != (sum & 0x1) {
				sum = (sum >> 1) ^ 0xA001
			} else {
				sum = 0x7FFF & (sum >> 1)
			}
		}

	}
	low = byte(sum & 0xFF)
	high = byte((sum >> 8) & 0xFF)
	return low, high
}

func checkModbusCRC16(data []byte) bool {
	len := len(data)
	l, h := crc16Modbus(data[:len-2])
	if l == data[len-2] && h == data[len-1] {
		return true
	}
	return false
}

// addModebusCRC16 把数据后两位改为CRC校验码
func addModebusCRC16(data []byte) []byte {
	len := len(data)
	l, h := crc16Modbus(data[:len-2])
	data[len-2] = l
	data[len-1] = h
	return data
}

/* CRC 高位字节值表 */
var auchCRCHi = [...]byte{
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
	0x00, 0xC1, 0x81, 0x40, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0,
	0x80, 0x41, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1,
	0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0, 0x80, 0x41,
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1,
	0x81, 0x40, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x00, 0xC1, 0x81, 0x40,
	0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1,
	0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x00, 0xC1, 0x81, 0x40,
	0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0,
	0x80, 0x41, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
	0x00, 0xC1, 0x81, 0x40, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
	0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40, 0x00, 0xC1, 0x81, 0x40,
	0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1,
	0x81, 0x40, 0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
	0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41, 0x01, 0xC0,
	0x80, 0x41, 0x00, 0xC1, 0x81, 0x40}

/* CRC低位字节值表*/
var auchCRCLo = [...]byte{
	0x00, 0xC0, 0xC1, 0x01, 0xC3, 0x03, 0x02, 0xC2, 0xC6, 0x06,
	0x07, 0xC7, 0x05, 0xC5, 0xC4, 0x04, 0xCC, 0x0C, 0x0D, 0xCD,
	0x0F, 0xCF, 0xCE, 0x0E, 0x0A, 0xCA, 0xCB, 0x0B, 0xC9, 0x09,
	0x08, 0xC8, 0xD8, 0x18, 0x19, 0xD9, 0x1B, 0xDB, 0xDA, 0x1A,
	0x1E, 0xDE, 0xDF, 0x1F, 0xDD, 0x1D, 0x1C, 0xDC, 0x14, 0xD4,
	0xD5, 0x15, 0xD7, 0x17, 0x16, 0xD6, 0xD2, 0x12, 0x13, 0xD3,
	0x11, 0xD1, 0xD0, 0x10, 0xF0, 0x30, 0x31, 0xF1, 0x33, 0xF3,
	0xF2, 0x32, 0x36, 0xF6, 0xF7, 0x37, 0xF5, 0x35, 0x34, 0xF4,
	0x3C, 0xFC, 0xFD, 0x3D, 0xFF, 0x3F, 0x3E, 0xFE, 0xFA, 0x3A,
	0x3B, 0xFB, 0x39, 0xF9, 0xF8, 0x38, 0x28, 0xE8, 0xE9, 0x29,
	0xEB, 0x2B, 0x2A, 0xEA, 0xEE, 0x2E, 0x2F, 0xEF, 0x2D, 0xED,
	0xEC, 0x2C, 0xE4, 0x24, 0x25, 0xE5, 0x27, 0xE7, 0xE6, 0x26,
	0x22, 0xE2, 0xE3, 0x23, 0xE1, 0x21, 0x20, 0xE0, 0xA0, 0x60,
	0x61, 0xA1, 0x63, 0xA3, 0xA2, 0x62, 0x66, 0xA6, 0xA7, 0x67,
	0xA5, 0x65, 0x64, 0xA4, 0x6C, 0xAC, 0xAD, 0x6D, 0xAF, 0x6F,
	0x6E, 0xAE, 0xAA, 0x6A, 0x6B, 0xAB, 0x69, 0xA9, 0xA8, 0x68,
	0x78, 0xB8, 0xB9, 0x79, 0xBB, 0x7B, 0x7A, 0xBA, 0xBE, 0x7E,
	0x7F, 0xBF, 0x7D, 0xBD, 0xBC, 0x7C, 0xB4, 0x74, 0x75, 0xB5,
	0x77, 0xB7, 0xB6, 0x76, 0x72, 0xB2, 0xB3, 0x73, 0xB1, 0x71,
	0x70, 0xB0, 0x50, 0x90, 0x91, 0x51, 0x93, 0x53, 0x52, 0x92,
	0x96, 0x56, 0x57, 0x97, 0x55, 0x95, 0x94, 0x54, 0x9C, 0x5C,
	0x5D, 0x9D, 0x5F, 0x9F, 0x9E, 0x5E, 0x5A, 0x9A, 0x9B, 0x5B,
	0x99, 0x59, 0x58, 0x98, 0x88, 0x48, 0x49, 0x89, 0x4B, 0x8B,
	0x8A, 0x4A, 0x4E, 0x8E, 0x8F, 0x4F, 0x8D, 0x4D, 0x4C, 0x8C,
	0x44, 0x84, 0x85, 0x45, 0x87, 0x47, 0x46, 0x86, 0x82, 0x42,
	0x43, 0x83, 0x41, 0x81, 0x80, 0x40}

func tableCRC16(data []byte) (low byte, high byte) {
	uchCRCHi := uint8(0xFF) /* 高CRC字节初始化 */
	uchCRCLo := uint8(0xFF) /* 低CRC 字节初始化 */
	crcIndex := uint8(0)    /* CRC循环中的索引 */

	for _, v := range data {
		crcIndex = uchCRCHi ^ uint8(v) /* 计算CRC */
		uchCRCHi = uchCRCLo ^ auchCRCHi[crcIndex]
		uchCRCLo = auchCRCLo[crcIndex]
	}
	low = uchCRCLo
	high = uchCRCHi
	return
}

func tableCheckCRC(data []byte) bool {
	len := len(data)
	if len < 2 {
		return false
	}
	l, h := tableCRC16(data[:len-2])
	if l == data[len-2] && h == data[len-1] {
		return true
	}
	return false
}

// dianBiaoAddCRC 把数据后两位改为CRC校验码
func wuShuiAddCRC(data []byte) []byte {
	len := len(data)
	l, h := tableCRC16(data[:len-2])
	data[len-2] = l
	data[len-1] = h
	return data
}

// LRC 校验
/*
Uint8 dsp_lrc_check(Uint8 buf[], Uint16 len)
{
 Uint16 iCount  = 0;
 Uint8 lrcValue = 0x00;

 for(iCount = 0; iCount < len ; iCount ++) {

   lrcValue = lrcValue + buf[iCount];

  }

// return ((unsigned char)((~lrcValue) + 1));      //两种操作都能实现
 return ((unsigned char)(-((char)lrcValue)));

}
*/
func modbusLRC(data []byte) (low, high byte) {
	return 0, 0
}
