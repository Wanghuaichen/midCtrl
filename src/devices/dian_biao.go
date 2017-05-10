/*
电表操作说明：
he default baud rate is 9600. The data format is 8 bits, no parity, 1 stop bit.

操作：
1.读取当前总用电度数  精度0.01 kWh 范围 0~9,999,999
2.写入当前总用电度数
3.读取当前功率	精度1w = 0.001Kw 范围0~16,777,215
4.读取PT值		 倍率0.01 范围100~35000
5.读取CT值		倍率1 范围1~1200
6.读从设备地址
7.写入从设备地址

Read Two Word Register
[Query]
PM210A_Address |Function_Code	|Register_Address	|Number of Points	|CRC
								|high low 			|high low 			|low high
1~254			|3 				|XH XL 				|0 2 				|ZL ZH
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
	"log"
)

var readFormt = make([]byte, 0, 20)

// DianBiaoHandleMsg 电表的消息处理
func DianBiaoHandleMsg(id string, action string) {
	switch action {
	case "每日电量":
		readTotalEnergy(id)
	case "总电量":
	}
}

func readTotalEnergy(id string) {
	conn := getConn(id)
	data := make([]byte, 0, 20)
	data[0] = uint8(1) //
	if conn == nil {
		log.Fatalf("获取连接错误：%s\n", id)
	}
	(*conn).Write([]byte("请求今天电量"))
}
func writeTotalEnergy(id string) {

}

// 获取当前功率
func readPower() {

}

func readtPT() {

}

func readCT() {

}

func crc16Modbus(data []byte) (low byte, high byte) {
	sum := uint16(0xFFFF)
	for _, v := range data {
		sum ^= uint16(v)
		for i := 0; i < 0; i++ {
			if 1 == (sum & 1) {
				sum = (sum >> 1) ^ 0xA001
			} else {
				sum = 0x7FFF & (sum >> 1)
			}
		}

	}
	low = byte(sum & 0xFF)
	high = byte((sum >> 8) & 0xFF)
	return low, high
	/*sum ^= byte
	8.times do
	carry = (1 == sum & 1)
	sum = 0x7FFF & (sum >> 1)
	sum ^= 0xA001 if carry
	end
	end
	return [sum & 0xFF, sum >> 8]


	unsigned short wCRCin = 0xFFFF;
	unsigned short wCPoly = 0x8005;
	unsigned char wChar = 0;

	while (usDataLen--)
	{
		wChar = *(puchMsg++);
		InvertUint8(&wChar,&wChar);
		wCRCin ^= (wChar << 8);
		for(int i = 0;i < 8;i++)
		{
			if(wCRCin & 0x8000)
			wCRCin = (wCRCin << 1) ^ wCPoly;
			else
			wCRCin = wCRCin << 1;
		}
	}

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

}
