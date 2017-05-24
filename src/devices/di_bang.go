package devices

import (
	"bufio"
	"log"
	"net/url"
	"strconv"
	"strings"
)

//d39
// 地磅数据读取

/*
485 通信 9600、19200bps
采用ascii码传数据
STX 数据开始 XON 0x2
CR 数据结束 换行 0x

采用方式一，9600，485 AB线反接 232 连续发生  采用ascii编码 倒叙模式
数据模式 .0600000=.0700000=.0700000=.0700000=.0700000=
*/

func diBangD39Start(id uint) {
	go hanldeD39(id)
	log.Println("地磅开始处理数据")
}

func hanldeD39(id uint) {
	conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		return
	}
	strPre := "" //上次数据
	strConuter := 0
	tryTimes := 0
	for {
		buff := bufio.NewReader(conn)
		dat, err := buff.ReadString('=')
		if err != nil {
			log.Printf("地磅读数据错误：%s\n", err.Error())
			tryTimes++
			if tryTimes > 5 {
				unBindConn(id)
				//上报链接断开
				return
			}
			continue
		}
		tryTimes = 0
		if len(dat) != 9 { //一帧数据为9个字节包括等号
			continue
		}
		if strings.Compare(strPre, dat) == 0 {
			strConuter++
			if strConuter > 5 {
				w := dibangDataTrans([]byte(dat)[:len(dat)-1])
				urlData := url.Values{"weight": {w}}
				sendData("地磅", id, urlData)
			}
		} else {
			strConuter = 0 //重新计数
		}
		strPre = dat //保存数据多次相同时才上报
	}

}

func dibangDataTrans(dat []byte) string {
	r := invert(dat)
	kg, err := strconv.ParseFloat(string(r), 32)
	if err != nil {
		log.Printf("d39转换数据失败:%s %s\n", string(r), err.Error())
	}
	return strconv.FormatInt(int64(kg*10000), 10)
}
