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
*/

package devices

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	CRC_ERROR = "CRC_ERROR"
)

var readFormt = make([]byte, 0, 20)

// DianBiaoHandleMsg 电表的消息处理
func DianBiaoHandleMsg(id string, action string) {
	switch action {
	case "总电量":
		readTotalEnergy(id)
	case "功率":
	}
}

func reqDevData(conn net.Conn, cmd []byte, tryTimes int) (rspData []byte, err error) {
	cmd = addCRC(cmd)
	var len int
	buff := make([]byte, 20)
	fmt.Printf("请求设备数据：%v\n", tryTimes)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	for ; tryTimes > 0; tryTimes-- {
		fmt.Printf("写入设备数据：%v\n", cmd)
		_, err = conn.Write(cmd) //发送数据到设备
		if err != nil {
			log.Println("写入设备数据失败：", err.Error())
			return []byte{}, err
		}

		//读取设备回复的数据
		len, err = (conn).Read(buff)
		if err != nil {

			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				log.Printf("读取数据超时:%s\n", err.Error())
				continue
			} else {
				log.Printf("读取数据失败:%s\n", err.Error()) //可能连接已经断开
				_, err := conn.Write([]byte{0})        //发送测试连接数据到设备
				if err != nil {
					log.Println("写入设备数据失败：", err.Error())
					return []byte{}, err
				} else {
					continue
				}
			}
		}
		if !checkCRC(buff[:len]) {
			log.Printf("CRC校验错误：%v\n", buff[:len])
			err = errors.New(CRC_ERROR)
			continue
		}
		break
	}
	rspData = buff[:len]
	return
}
func readTotalEnergy(id string) {
	conn := getConn(id)
	if conn == nil {
		log.Printf("获取连接错误：%s\n", id)
		return
	}
	//构造要发送的数据，计算CRC
	data := []byte{0x1, 0x3, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0}
	buff, err := reqDevData(conn, data, 3)
	if err != nil {
		log.Printf("获取设备数据失败：%s\n", err.Error())
		if err.Error() == CRC_ERROR {
			log.Println("获取总电量时CRC 校验失败")
		} else if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			log.Println("不能读到设备数据，需要检查设备和转化设备连接是否正常")
		} else {
			log.Printf("%s断开连接\n", id)
			unBindConn(id) //连接已经断开
			conn.Close()
		}
		return
	}
	fmt.Printf("收到设备数据：%v\n", buff)
	// (Y3 * 16,777,216 + Y4 * 65,536 + Y1 * 256 + Y2) * unit
	totalEnergy := uint32(buff[5])*0x1000000 + uint32(buff[6])*0x10000 + uint32(buff[3])*0x100 + uint32(buff[4])
	totalEnergyStr := strconv.FormatFloat(float64(totalEnergy)*0.01, 'f', 2, 64)

	sendServ([]byte(generateDataJsonStr(id, "总电量", totalEnergyStr)))
}
func writeTotalEnergy(id string) {

}

// 获取当前功率
func readPower(id string) {
	data := []byte{0x1, 0x3, 0x0, 0x8, 0x0, 0x2, 0x0, 0x0}
	data = addCRC(data)
	//(*conn).Write(data) //发送数据到设备
}

func readtPT(id string) {

}

func readCT(id string) {

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

func checkCRC(data []byte) bool {
	len := len(data)
	l, h := crc16Modbus(data[:len-2])
	if l == data[len-2] && h == data[len-1] {
		return true
	} else {
		return false
	}
}

// addCRC 把数据后两位改为CRC校验码
func addCRC(data []byte) []byte {
	len := len(data)
	l, h := crc16Modbus(data[:len-2])
	data[len-2] = l
	data[len-1] = h
	return data
}
