package devices

import (
	"bytes"
	"encoding/binary"
	"log"
	"strconv"
)

//环境监测  噪声 扬尘 方向等
//校验crc16Modbus
//通讯参数：波特率 9600  数据位 8位  无校验位

var hjAddr uint8

type realtimeDataType struct {
	ch1  uint16
	ch2  uint16
	ch3  uint16
	ch4  uint16
	ch5  uint16
	ch6  uint16
	ch7  uint16
	ch8  uint16
	ch9  uint16
	ch10 uint16
	ch11 uint16
	ch12 uint16
	ch13 uint16
	ch14 uint16
	ch15 uint16
	ch16 uint16
	sw1  uint16
	sw2  uint16
	sw3  uint16
	sw4  uint16
	sw5  uint16
	sw6  uint16
	sw7  uint16
	sw8  uint16
	sw9  uint16
	sw10 uint16
	sw11 uint16
	sw12 uint16
	sw13 uint16
	sw14 uint16
	sw15 uint16
	sw16 uint16
	sw17 uint16
	sw18 uint16
	sw19 uint16
	sw20 uint16
	sw21 uint16
	sw22 uint16
	sw23 uint16
	sw24 uint16
	sw25 uint16
	sw26 uint16
	sw27 uint16
	sw28 uint16
	sw29 uint16
	sw30 uint16
	sw31 uint16
	sw32 uint16
}

func (r realtimeDataType) toServData() map[string][]string {
	var result = make(map[string][]string, 17)
	result["pm10"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)}
	result["pm25"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)}
	result["humidity"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)}    //湿度
	result["temperature"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)} //温度
	result["windSpeed"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)}   //风速
	result["windFrom"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)}    //风向
	result["db"] = []string{strconv.FormatInt(int64(r.ch1*10000), 10)}          //分贝
	return result
}

//初始化环境设备，获取其地址
func initHuanJing(id uint) {
	getAddrCmd := []byte{0x00, 0x20, 0x00, 0x68}
	dat, err := reqDevData(id, getAddrCmd, nil, checkModbusCRC16)
	if err != nil {
		log.Printf("获取环境监测设备地址错误：%s\n", err.Error())
		return
	}
	hjAddr = dat[2]
}

// 获取环境监测设备的实时数据
func huanJingRealData(id uint) {
	cmd := []byte{hjAddr, 0x03, 0x00, 0x00, 0x00, 0x00}
	dat, err := reqDevData(id, cmd, addModebusCRC16, checkModbusCRC16)
	if err != nil {
		log.Printf("获取环境监测实时数据失败:%s\n", err.Error())
		return
	}
	if dat[3] != 0x40 { //固定数据长度
		log.Printf("获取环境监测实时数据格式不正确:%v\n", dat)
		return
	}
	var realData realtimeDataType
	binary.Read(bytes.NewReader(dat[4:68]), binary.LittleEndian, realData)

}
func transData(dat uint16) int16 {
	if 0x8000&dat == 0x8000 { //dat 最高为1
		dat = dat & 0x7FFF
		dat = (^dat & 0x7FFF) + 1
		return int16(-dat)
	}
	return int16(dat)
}
