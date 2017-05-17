//偶效验、停止位1、通讯ID为表计地址、波特率为2400，数据位8
//485表计对应线：绿色485B。白色485A

/*
读取当前表具ID号，主机发送：
68 10 AA AA AA AA AA AA AA 03 03 0A 81 A7 56 16
当前表具ID号为78330123456789，从机应答：
FE FE 68 10 89 67 45 23 01 33 78 83 03 0A 81 A7 34 16

数据以BCD为BCD

请求数据
  协议头        68起始帧   10仪表类型(冷水表)    倒序为表号          厂家     控制码表示读     数据长度    数据标识     序列号     和校验   结束符
//FE FE FE FE   68        10                  44 33 22 11 00      33 78    01             03         1F 90        00         80      16
回复数据
                                           81=80+控制码      数据长度                   倒序累计用水量    单位立方米    倒序月累计      单位立方米    倒序实际时间2015-05-11 22:01:31   状态       和
FE FE FE FE 68 10 44 33 22 11 00 33 78     81               16           1F 90 00      00 77 66 55      2C           00 77 66 55    2C           31 01 22 11 05 15 20             21 84      08 16

回复数据的校验值不对

*/
package devices

import (
	"math"
)

// 读水表数据
func readSB(id uint) {
	//构造要发送的数据，计算CRC
	data := []byte{0xFE, 0xFE, 0xFE, 0xFE, 0x68, 0x10, 0x44, 0x33, 0x22, 0x11, 0x00, 0x33, 0x78, 0x01, 0x03, 0x1F, 0x90, 0x00, 0x80, 0x16}
	buff, err := reqDevData(id, data, nil, nil)
	if err != nil {
		return
	}
	//Data = (Y1*256 + Y2) * (unit = 0.01)
	//wuShui := buff[3:14]
	zongLeiJi := getShuiLiang(buff[18:22])
	//sendServ([]byte(generateDataJsonStr(id, "污水", string(wuShui))))
	sData := []map[string]interface{}{{"ph": int64(zongLeiJi) * 1000}}

	sendData(urlTable["水表"], id, sData)
}

func getShuiLiang(dat []byte) int {
	dat = invert(dat)
	r := 0
	for i, v := range dat {
		r = r + int(transBCD(v))*int(math.Pow10(i))
	}
	return r
}
func invert(dat []byte) []byte {
	len := len(dat)
	r := make([]byte, 0, len)
	for i := len - 1; i >= 0; i-- {
		r = append(r, dat[i])
	}
	return r
}

//将BCD码转为实际值
func transBCD(dat byte) uint8 {
	return dat&0xF0*10 + dat&0xF
}

func sum(dat []byte) byte {
	s := byte(0)
	for _, v := range dat {
		s = s + v
	}
	return s
}
