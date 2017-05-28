//偶效验、停止位1、通讯ID为表计地址、波特率为2400，数据位8
//485表计对应线：绿色485B。白色485A

/*
读取当前表具ID号，主机发送：
68 10 AA AA AA AA AA AA AA 03 03 0A 81 A7 56 16
68 10 AA AA AA AA AA AA AA 03 03 0A 81 A7 56 16
当前表具ID号为78330123456789，从机应答：
FE FE 68 10 89 67 45 23 01 33 78 83 03 0A 81 A7 34 16
传输时加上 1位起始位（0）、一个偶校验位、一个停止位（1） ，共11位。
通讯波特率为2400bps
数据以BCD为BCD

请求数据
  协议头        68起始帧   10仪表类型(冷水表)    倒序为表号          厂家     控制码表示读     数据长度    数据标识     序列号     和校验   结束符
  FE FE FE FE   68        10                  44 33 22 11 00      33 78    01             03         1F 90        00         80      16
回复数据
                                           81=80+控制码      数据长度                   倒序累计用水量    单位立方米    倒序月累计      单位立方米    倒序实际时间2015-05-11 22:01:31   状态       和
FE FE FE FE 68 10 44 33 22 11 00 33 78     81               16           1F 90 00      00 77 66 55      2C           00 77 66 55    2C           31 01 22 11 05 15 20             21 84      08 16

回复数据的校验值不对
读表具
fe fe  68 10 01 00 00 05 08 00 00 01 03 90 1f 00 39 16
FE FE  68 10 44 33 22 11 00 33 78 01 03 1F 90 00 80 16

fe fe  68 10 aa aa aa aa aa aa aa 01 03 1F 90 AA 7b 16    //90和1f 可互换  AA 为序号可为00 校验为从68开始的和数据
*/

package devices

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

var shuiBiaoPeriod = 20 * time.Second

func shuiBiaoAutoGet() {
	go readSB()
	log.Println("水表开始获取数据")
}

// 读水表数据
func readSB() {
	for {
		for _, id := range devTypeTable["水表"] {
			//构造要发送的数据
			//68 10 aa aa aa aa aa aa aa 01 03 90 1F AA 7b 16
			data := []byte{0xFE, 0xFE, 0x68, 0x10, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0x01, 0x03, 0x90, 0x1F, 0xAA, 0x7B, 0x16}
			buff, err := reqDevData(id, data, nil, nil)
			if err != nil {
				continue
			}
			//FE FE 68 10 45 41 10 05 15 33 78 81 16 90 1F AA 00 59 59 00 2C FF FF FF FF 2C FF FF FF FF FF FF FF 00 00 C2 16
			//去除开头的0xFE
			//fmt.Printf("水表请求数据到：%x\n", buff)
			for {
				if buff[0] == 0xFE && len(buff) > 2 {
					buff = buff[1:]
				} else {
					break
				}
			}
			fmt.Printf("buff:%x len:%d\n", buff, len(buff))
			if len(buff) != 35 || buff[34] != 0x16 {
				fmt.Printf("水表数据长度不对：%x\n", buff)
				continue
			}
			//fmt.Printf("水量：%x\n", buff[14:18])
			zongLeiJi := getShuiLiang(buff[14:18])
			//yueLeiJi := getShuiLiang(buff[19:23])
			sData := map[string][]string{"total": {strconv.FormatInt(int64(zongLeiJi)*1000, 10)}, "record": {"0"}}
			fmt.Printf("水表发送：%v\n", sData)
			sendData("水表", id, sData)
		}

		time.Sleep(shuiBiaoPeriod)
	}
}

func getShuiLiang(dat []byte) int {
	//dat = invert(dat)
	r := 0
	for i, v := range dat {
		r = r + int(transBCD(v))*int(math.Pow10(i*2))
		//fmt.Println(r)
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
	return (dat>>4)*10 + dat&0xF
}

func sum(dat []byte) byte {
	s := byte(0)
	for _, v := range dat {
		s = s + v
	}
	return s
}
