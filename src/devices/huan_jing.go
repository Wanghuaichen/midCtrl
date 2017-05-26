package devices

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

//环境监测  噪声 扬尘 方向等
//校验crc16Modbus
//通讯参数：波特率 9600  数据位 8位  无校验位

var hjAddr uint8
var huanJingPeriod = 20 * time.Second
var reqHuanJingTicker = time.NewTicker(huanJingPeriod) //请求环境数据周期
type realtimeDataType struct {
	ch [16]uint16
	sw [32]uint8
}

func (r realtimeDataType) toServData() map[string][]string {
	var result = make(map[string][]string, 17)
	result["pm10"] = []string{strconv.FormatInt(int64(transData(r.ch[3]))*10000, 10)}
	result["pm25"] = []string{strconv.FormatInt(int64(transData(r.ch[2]))*10000, 10)}
	result["humidity"] = []string{strconv.FormatInt(int64(transData(r.ch[5]))*1000, 10)}    //湿度
	result["temperature"] = []string{strconv.FormatInt(int64(transData(r.ch[4]))*1000, 10)} //温度
	result["windSpeed"] = []string{strconv.FormatInt(int64(transData(r.ch[0]))*1000, 10)}   //风速
	result["windFrom"] = []string{strconv.FormatInt(int64(transData(r.ch[6]))*10000, 10)}   //风向
	result["db"] = []string{strconv.FormatInt(int64(transData(r.ch[1]))*1000, 10)}          //分贝
	return result
}
func (r *realtimeDataType) setRealData(dat []byte) (err error) {
	if len(dat) != 0x40 {
		return errors.New("数据长度不正确")
	}
	for i := 0; i < 16; i++ {
		r.ch[i] = doubleByteToUint16(dat[2*i : 2*(i+1)])
	}
	//fmt.Printf("环境通道转换后数据：%v\n", r.ch)
	for i := 0; i < 32; i++ {
		r.sw[i] = dat[32+i]
	}
	return nil
}
func doubleByteToUint16(d []byte) uint16 {
	r := uint16(uint16(d[0])<<8 | uint16(d[1]))
	return r
}

//初始化环境设备，获取其地址  01
/*func initHuanJing(id uint) {
	getAddrCmd := []byte{0x00, 0x20, 0x00, 0x68}
	dat, err := reqDevData(id, getAddrCmd, nil, checkModbusCRC16)
	if err != nil {
		log.Printf("获取环境监测设备地址错误：%s\n", err.Error())
		return
	}
	hjAddr = dat[2]
}
*/

//启动定时获取实时数据

func huanjingInitAutoGet() {
	go func() {
		for _ = range reqHuanJingTicker.C {
			for _, id := range devTypeTable["环境"] {
				huanJingRealData(id)
			}
		}
	}()
}

// 获取环境监测设备的实时数据
func huanJingRealData(id uint) {
	cmd := []byte{0x01, 0x03, 0x00, 0x00, 0xF1, 0xD8} //01 03 00 00 F1 D8
	dat, err := reqDevData(id, cmd, nil, checkModbusCRC16)
	if err != nil {
		log.Printf("获取环境监测实时数据失败:%s\n", err.Error())
		return
	}
	if dat[3] != 0x40 { //固定数据长度
		log.Printf("获取环境监测实时数据格式不正确:%v\n", dat)
		return
	}
	var realData realtimeDataType
	//err = binary.Read(bytes.NewReader(dat[4:68]), binary.LittleEndian, realData) //0x48个字节
	err = realData.setRealData(dat[4:68])
	if err != nil {
		log.Printf("环境数据转换失败：%s\n", err.Error())
		return
	}
	fmt.Printf("环境实时数据：%v\n", realData.toServData())
	sendData("环境", id, realData.toServData())
}
func transData(dat uint16) int16 {
	if 0x8000&dat == 0x8000 { //dat 最高为1
		dat = dat & 0x7FFF
		dat = (^dat & 0x7FFF) + 1
		return int16(-dat)
	}
	return int16(dat)
}
