/*
电表操作说明：
he default baud rate is 9600. The data format is 8 bits, no parity, 1 stop bit.
操作：
1.读取当前总用电度数  精度0.01 kWh 范围 0~9,999,999  reg：0,1
2.写入当前总用电度数
3.读取当前功率	精度1w = 0.001Kw 范围0~16,777,215  reg：8,9
4.读取PT值		 倍率0.01 范围100~35000 reg：16
5.读取CT值		倍率1 范围1~1200        reg：17
6.读从设备地址    reg：5
7.写入从设备地址
Read Two Word Register
[Query]
PM210A_Address |Function_Code	|Register_Address	|Number of Points	|CRC
								|high   low 		|high   low 		|low    high
1~254		   |3 				|XH     XL 			|0      2 			|ZL     ZH
[notes]
XH = Register Address div 256
XL = Register Address mod 256
[Reply]
PM210A_Address	|Function_Code	|Byte_Count	|Read_Data(low word) |Read_Data(high word) |CRC
											|high low 			 |high low 			   |low high
1~254 			|3 				|4 			|Y1 Y2 				 |Y3 Y4 			   |ZL ZH
[notes]
Data = (Y3 * 16,777,216 + Y4 * 65,536 + Y1 * 256 + Y2) * Unit

电表设备使用 【51000，52000）端口
*/

package devices

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"time"
)

const (
	CRC_ERROR = "CRC_ERROR"
)

var readFormt = make([]byte, 0, 20)

//string 行为，int间隔秒数
var dianBiaoPeriod = 20 * time.Second

//var dianBiaoSync = make(chan bool, 1)

// 初始化按自动间隔获取数据
func dianBiaoInitAutoGet() {
	go diaoBiaoGetData()
	log.Printf("电表开始获取数据\n")
}

func delay() {
	rand.Seed(time.Now().UnixNano())
	delayMs := rand.Uint32()%500 + 500
	time.Sleep(time.Duration(delayMs))
}

func diaoBiaoGetData() {
	for {
		for _, id := range devTypeTable["电表"] {
			dianBiaoJSONRecod := url.Values{"kw": {""}, "pt": {""}, "ct": {""}, "record": {""}}
			conn := getConn(id)
			if conn == nil {
				log.Printf("获取连接错误：%d\n", id)
				continue
			}

			d, err := readPower(id)
			if err != nil {
				continue
			}
			dianBiaoJSONRecod["kw"] = []string{strconv.FormatInt(d, 10)}

			d, err = readTotalEnergy(id)
			if err != nil {
				continue
			}
			dianBiaoJSONRecod["record"] = []string{strconv.FormatInt(d, 10)}

			/*d, err = readtPT(id)
			if err != nil {
				continue
			}
			dianBiaoJSONRecod["pt"] = []string{strconv.FormatInt(d, 10)}

			d, err = readCT(id)
			if err != nil {
				continue
			}
			dianBiaoJSONRecod["ct"] = []string{strconv.FormatInt(d, 10)}
			*/
			dianBiaoJSONRecod["pt"] = []string{"0"}
			dianBiaoJSONRecod["ct"] = []string{"0"}
			fmt.Printf("电表发送：%v\n", dianBiaoJSONRecod)
			sendData("电表", id, dianBiaoJSONRecod)
		}
		time.Sleep(dianBiaoPeriod)
	}
}

func reqDevData(id uint, cmd []byte, addCRC func([]byte) []byte, checkCRC func([]byte) bool) (rspData []byte, err error) {
	conn := getConn(id)
	if conn == nil {
		log.Printf("获取连接错误：%d\n", id)
		err = errors.New("获取连接错误")
		return []byte{}, err
	}
	if addCRC != nil {
		cmd = addCRC(cmd)
	}

	var len int
	buff := make([]byte, 300) //接收数据的buff
	defer func() {
		conn.SetDeadline(time.Time{})
	}()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for tryTimes := 3; tryTimes > 0; tryTimes-- {
		//fmt.Printf("写入设备数据：%v\n", cmd)
		_, err = conn.Write(cmd) //发送数据到设备
		if err != nil {
			log.Printf("%d写入设备数据失败：%s\n", id, err.Error())
			return []byte{}, err
		}

		//读取设备回复的数据
		time.Sleep(time.Millisecond * 100) //等待100ms 等数据全部发完
		len, err = (conn).Read(buff)
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				log.Printf("%d读取数据超时:%s\n", id, err.Error())
				continue
			} else {
				log.Printf("%d读取数据失败:%s\n", id, err.Error()) //可能连接已经断开
				_, err := conn.Write([]byte{0})              //发送测试连接数据到设备
				if err != nil {
					log.Printf("%d写入设备数据失败：%s", id, err.Error())
					return []byte{}, err
				}
				continue
			}
		}
		if checkCRC != nil {
			if !checkCRC(buff[:len]) {
				log.Printf("%d CRC校验错误：%v\n", id, buff[:len])
				err = errors.New(CRC_ERROR)
				continue
			}
		}
		break
	}
	rspData = buff[:len]
	if err != nil {
		log.Printf("%d获取设备数据失败：%s\n", id, err.Error())
		if err.Error() == CRC_ERROR {
			log.Printf("%d获取数据时CRC校验失败\n", id)
		} else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			log.Printf("%d 不能读到设备数据，需要检查设备和转化设备连接是否正常\n", id)
			//relayError(id, "NO_Data")
		} else {
			log.Printf("%d断开连接\n", id)
			unBindConn(id) //连接已经断开
			conn.Close()
			//relayError(id, "Disconnect")
		}
		return
	}
	//fmt.Printf("收到设备数据：%v\n", buff[:len])
	return
}

func readTotalEnergy(id uint) (int64, error) {
	//构造要发送的数据，计算CRC
	//01 03 00 00 00 02 C4 0B   -->>01 03 04 09 18 09 13 3E 35
	//01 03 00 00 00 09 85 CC   -->>01 03 12 09 06 09 06 09 02 0F 9C 0F 98 0F 98 00 AC 00 7A 00 7C 48 BB
	//01 03 00 00 00 0A C5 CD   -->>01 03 14 09 1B 09 16 09 18 0F BB 0F B9 0F BD 00 CD 00 9C 00 9D 00 5A B9 22

	data := []byte{0x1, 0x3, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0}
	buff, err := reqDevData(id, data, addModebusCRC16, checkModbusCRC16)
	if err != nil {
		return 0, err
	}
	// (Y3 * 16,777,216 + Y4 * 65,536 + Y1 * 256 + Y2) * (unit=0.01)KwH
	totalEnergy := uint32(buff[5])*0x1000000 + uint32(buff[6])*0x10000 + uint32(buff[3])*0x100 + uint32(buff[4])
	return int64(totalEnergy) * 100, nil
}
func writeTotalEnergy(id string) {

}

// 获取当前功率
func readPower(id uint) (int64, error) {
	data := []byte{0x1, 0x3, 0x0, 0x8, 0x0, 0x2, 0x0, 0x0}
	buff, err := reqDevData(id, data, addModebusCRC16, checkModbusCRC16)
	if err != nil {
		return 0, err
	}
	//Data = (Y3 * 16,777,216 + Y4 * 65,536 + Y1 * 256 + Y2) * (unit=0.001) Kw
	power := uint32(buff[5])*0x1000000 + uint32(buff[6])*0x10000 + uint32(buff[3])*0x100 + uint32(buff[4])
	//powerStr := strconv.FormatFloat(float64(power)*0.001, 'f', 3, 64)
	return int64(power) * 10, nil

}

func readtPT(id uint) (int64, error) {
	data := []byte{0x1, 0x3, 0x0, 0x10, 0x0, 0x1, 0x0, 0x0}
	buff, err := reqDevData(id, data, addModebusCRC16, checkModbusCRC16)
	if err != nil {
		return 0, err
	}
	//Data = (Y1*256 + Y2) * (unit = 0.01)
	pt := uint32(buff[3])*0x100 + uint32(buff[4])
	//PtStr := strconv.FormatFloat(float64(pt)*0.01, 'f', 2, 64)

	return int64(pt) * 100, nil

}

func readCT(id uint) (int64, error) {
	data := []byte{0x1, 0x3, 0x0, 0x11, 0x0, 0x1, 0x0, 0x0}
	buff, err := reqDevData(id, data, addModebusCRC16, checkModbusCRC16)
	if err != nil {
		return 0, err
	}
	//Data = (Y1 * 256 + Y2) * (unit=1)
	ct := uint32(buff[3])*0x100 + uint32(buff[4])
	//CtStr := strconv.FormatFloat(float64(ct), 'f', 2, 64)
	//CtStr := strconv.FormatUint(uint64(ct), 10)
	return int64(ct) * 10000, nil
}

/*
   文档 ruby算法描述
    sum ^= byte
	8.times do
	    carry = (1 == sum & 1)
	    sum = 0x7FFF & (sum >> 1)
	    sum ^= 0xA001 if carry
	end
	return [sum & 0xFF, sum >> 8]

    网络C描述
	uint16_t crc16_modbus(uint8_t *data, uint_len length)
	{
	    uint8_t i;
	    uint16_t crc = 0xffff;        // Initial value
	    while(length--)
	    {
	        crc ^= *data++;            // crc ^= *data; data++;
	        for (i = 0; i < 8; ++i)
	        {
	            if (crc & 1)
	                crc = (crc >> 1) ^ 0xA001;        // 0xA001 = reverse 0x8005
	            else
	                crc = (crc >> 1);
	        }
	    }
	    return crc;
	}
*/

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
