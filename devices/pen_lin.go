package devices

import (
	"bytes"
	"log"
	"time"
)

var penLinPeriod = 5 * time.Second

func penLinStart(id uint) {
	conn := getConn(id)
	if conn == nil {
		return
	}
	defer func() {
		conn.Close() //关闭连接
		log.Printf("喷淋监测处理发生错误\n")
		unBindConn(id)

		//设置设备状态
	}()

	rCh := make(chan []byte, 10)
	wCh := make(chan []byte)
	openCmd := []byte{0x55, 0x01, 0x12, 0x00, 0x00, 0x00, 0x01, 0x69}       //55 01 12 00 00 00 01 69
	openReplyCmd := []byte{0x22, 0x01, 0x12, 0x00, 0x00, 0x00, 0x01, 0x36}  //22 01 12 00 00 00 01 36
	closeCmd := []byte{0x55, 0x01, 0x11, 0x00, 0x00, 0x00, 0x01, 0x68}      //55 01 11 00 00 00 01 68
	closeReplyCmd := []byte{0x22, 0x01, 0x11, 0x00, 0x00, 0x00, 0x00, 0x34} //22 01 11 00 00 00 00 34
	stataCh := make(chan bool, 1)
	var cmd int
	timeout := time.NewTimer(penLinPeriod)
	go sendCmd(conn, wCh, stataCh)
	go readOneData(conn, rCh, []byte{0x22, 0x01}, 8, stataCh)
	for {
		var dat []byte
		var state bool
		select {
		case dat = <-rCh:
			if bytes.Equal(openReplyCmd, dat) {
				log.Printf("电磁已经打开")
			}
			if bytes.Equal(closeReplyCmd, dat) {
				log.Printf("电磁已经关闭")
			}
			timeout.Stop()
			devList[id].isOk = 1
			break
		case cmd = <-devList[id].cmd:
			if cmd == 1 { //开电磁阀
				wCh <- openCmd
			}
			if cmd == 2 {
				wCh <- closeCmd
			}
			timeout.Reset(penLinPeriod)
			break
		case state = <-stataCh:
			if false == state {
				return
			}
		case <-timeout.C:
			log.Printf("电磁阀执行命令超时\n")
			devList[id].isOk = 0
			break
		}
		//time.Sleep(wuShuiPeriod)
	}
}
