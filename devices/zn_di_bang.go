package devices

import (
	"bytes"
	"log"
	"net/url"
	"strconv"
)

//yh-500
// 地磅数据读取

/*
232 通信 9600、19200bps


采用ascii码传数据
STX 数据开始 XON 0x2
CR 数据结束 换行 0x

采用方式一，9600，485 AB线反接 232 连续发生  采用ascii编码 倒叙模式
数据模式 .0600000=.0700000=.0700000=.0700000=.0700000=
*/

func znDiBangStart(id uint) {
	conn := getConn(id)
	if conn == nil {
		return
	}
	defer func() {
		conn.Close() //关闭连接
		log.Printf("移动地磅监测处理发生错误\n")
		unBindConn(id)

		//设置设备状态
	}()

	rCh := make(chan []byte, 10)
	var strPre []byte //上次数据
	strConuter := 0
	stataCh := make(chan bool, 1)
	go readOneData(conn, rCh, []byte{'='}, 10, stataCh)
	for {
		var dat []byte
		var state bool
		select {
		case dat = <-rCh:
			break
		case state = <-stataCh:
			if false == state {
				return
			}
		}
		log.Printf("移动地磅数据：%s\n", dat)
		if bytes.Equal(strPre, dat) {
			strConuter++
			if strConuter > 6 {
				//log.Printf("地磅数据：%s\n", dat) //=.0036100 -92233720
				w := znDibangDataTrans([]byte(dat)[1:8])
				if w != "0" {
					urlData := url.Values{"weight": {w}}
					//fmt.Printf("地磅发送数据:%v\n", urlData)
					sendData("智能地磅", id, urlData)
				}
				strConuter = 0 //重新计数
			}
		} else {
			strConuter = 0 //重新计数
		}
		strPre = dat //保存数据多次相同时才上报
		//time.Sleep(wuShuiPeriod)
	}
}

func znDibangDataTrans(dat []byte) string {
	r := invert(dat)
	//log.Printf("地磅数据：%s\n", dat)
	kg, err := strconv.ParseFloat(string(r), 64)
	if err != nil {
		log.Printf("d39转换数据失败:%s %s\n", string(r), err.Error())
	}
	//log.Printf("地磅转化后数据：%f\n", kg)
	return strconv.FormatInt(int64(kg*1000000000000)/100000000, 10) //乘大数，防止产生9999的小数
}
