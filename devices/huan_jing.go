package devices

import (
	"bytes"
	"errors"
	"log"
	"net"
	"strconv"
	"time"
)

//环境监测  噪声 扬尘 方向等
//校验crc16Modbus
//通讯参数：波特率 9600  数据位 8位  无校验位

/*
01 03 00 40 00 1C 02 BF 00 17 00 1A 01 1A 01 FB
00 B7 7F FF 7F FF 7F FF 7F FF 7F FF 7F FF 7F FF 7F FF 7F FF 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 72 B6
*/

var hjAddr uint8
var huanJingPeriod = 60 * time.Second

//var reqHuanJingTicker = time.NewTicker(huanJingPeriod) //请求环境数据周期
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

/*func huanjingInitAutoGet() {
	go func() {
			for _, id := range devTypeTable["环境"] {
				huanJingRealData(id)
			}
	}()
}*/

func checkState(stataCh <-chan bool) bool {
	select {
	case <-stataCh: //读到读写进程的状态则返回错误
		//log.Printf("状态异常")
		return false
	default:
		return true
	}
}

// 获取环境监测设备的实时数据
func readOneData(conn net.Conn, rCh chan<- []byte, datHead []byte, datLen int, stataCh chan<- bool) {
	defer func() {
		stataCh <- false //发生错误退出时，通知处理协程处理
	}()
	if conn == nil {
		return
	}
	var tempBuff []byte
	buff := make([]byte, 1024)

	for {
		n, err := conn.Read(buff)
		//log.Printf("%v读到：%x \n", conn, buff[:n])
		if err != nil {
			//panic(err.Error())
			log.Printf("读数据错误：%s\n", err.Error())
			return
		}
		tempBuff = append(tempBuff, buff[:n]...)
		//log.Printf("当前数据：%x \n", tempBuff)

		index := bytes.Index(tempBuff, datHead)
		if index == -1 {
			//log.Printf("在 %x 中位找不到数据头：%x\n", tempBuff, datHead)
			continue //继续读数据
		}
		if len(tempBuff[index:]) < datLen { //未取得完整数据
			//log.Printf("数据不完整 %x\n", tempBuff)
			continue //继续读数据
		}

		rCh <- tempBuff[index : index+datLen]
		tempBuff = tempBuff[index+datLen:]
	}
}

func huanJingStart(id uint) {
	conn := getConn(id)
	if conn == nil {
		return
	}
	defer func() {
		conn.Close()
		log.Printf("环境监测处理发生错误\n")
		unBindConn(id)
		//设置设备状态
	}()
	cmd := []byte{0x01, 0x03, 0x00, 0x00, 0xF1, 0xD8} //01 03 00 00 F1 D8  //获取实时环境命令
	rCh := make(chan []byte, 5)
	wCh := make(chan []byte)
	stataCh := make(chan bool, 1)
	timeout := time.NewTimer(huanJingPeriod * 2)
	go sendCmd(conn, wCh, stataCh)
	go readOneData(conn, rCh, []byte{0x1, 0x3, 0x0, 0x40}, 4+0x40+2, stataCh)
	for {
		var dat []byte
		var state bool
		wCh <- cmd
		timeout.Reset(huanJingPeriod * 2)
		select {
		case dat = <-rCh:
			break
		case state = <-stataCh:
			if false == state {
				return
			}
		case <-timeout.C:
			log.Printf("电表读数据超时重新发送读取数据\n")
			continue
		}
		if !checkModbusCRC16(dat) {
			log.Printf("环境数据校验失败：%s\n", dat)
			continue
		}
		var realData realtimeDataType
		//err = binary.Read(bytes.NewReader(dat[4:68]), binary.LittleEndian, realData) //0x48个字节
		err := realData.setRealData(dat[4 : 0x40+4])
		if err != nil {
			log.Printf("环境数据转换失败：%s\n", err.Error())
			continue
		}
		//fmt.Printf("环境实时数据：%v\n", realData.toServData())
		sendData("环境", id, realData.toServData())
		time.Sleep(huanJingPeriod + time.Duration(time.Now().Unix()%10))
	}
}

// 给设备发生命令
func sendCmd(conn net.Conn, ch <-chan []byte, stataCh chan<- bool) {
	defer func() {
		stataCh <- false //发生错误退出时，通知处理协程处理
	}()
	for {
		cmd := <-ch
		//log.Printf("向设备发送:%v\n", cmd)
		n, err := conn.Write(cmd)
		if err != nil {
			//panic(err.Error())
			log.Printf("写数据错误：%s\n", err.Error())
			return
		}
		if n != len(cmd) {
			//log.Printf("发送数据不完整\n")
			continue
		}
	}
}

// 将无符号数转为有符号数
func transData(dat uint16) int16 {
	if 0x8000&dat == 0x8000 { //dat 最高为1
		dat = dat & 0x7FFF
		dat = (^dat & 0x7FFF) + 1
		return int16(-dat)
	}
	return int16(dat)
}
