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
	//openCmd := []byte{0x55, 0x01, 0x12, 0x00, 0x00, 0x00, 0x01, 0x69}         //55 01 12 00 00 00 01 69
	//openReplyCmd := []byte{0x22, 0x01, 0x12, 0x00, 0x00, 0x00, 0x01, 0x36}    //22 01 12 00 00 00 01 36
	//closeCmd := []byte{0x55, 0x01, 0x11, 0x00, 0x00, 0x00, 0x01, 0x68}        //55 01 11 00 00 00 01 68
	closeReplyCmd := []byte{0x22, 0x01, 0x11, 0x00, 0x00, 0x00, 0x00, 0x34}   //22 01 11 00 00 00 00 34
	checkStatusCmd := []byte{0x55, 0x01, 0x10, 0x00, 0x00, 0x00, 0x00, 0x66}  //55 01 10 00 00 00 00 66
	open20sCmd := []byte{0x55, 0x01, 0x21, 0x00, 0x4E, 0x20, 0x01, 0xE6}      //55 01 21 00 4E 20 01 E6  //20s后断开
	open20sReplyCmd := []byte{0x22, 0x01, 0x21, 0x00, 0x00, 0x00, 0x01, 0x45} //22 01 21 00 00 00 01 45  //20s后断开
	stataCh := make(chan bool, 1)
	var cmd int
	timeout := time.NewTimer(penLinPeriod)
	go sendCmd(conn, wCh, stataCh)
	go readOneData(conn, rCh, []byte{0x22, 0x01}, 8, stataCh)
	go func() { //1分钟查询一次状态
		wCh <- checkStatusCmd
		timeout.Reset(penLinPeriod)
		log.Printf("喷淋查询状态：%v\n", checkStatusCmd)
		time.Sleep(penLinPeriod * 20)
	}()
	for {
		var dat []byte
		var state bool
		select {
		case dat = <-rCh:
			if bytes.Equal(open20sReplyCmd, dat) {
				log.Printf("电磁已经打开")
			} else if bytes.Equal(closeReplyCmd, dat) {
				log.Printf("电磁已经关闭")
			} else {
				log.Printf("状态值：%v\n", dat)
				//timeout.Stop()
				//break
			}
			timeout.Stop()
			devList[id].isOk = 1
			break
		case cmd = <-devList[id].cmd:
			log.Printf("喷淋收到服务器命令：%d\n", cmd)
			if cmd == 1 { //开电磁阀
				wCh <- open20sCmd
			}
			if cmd == 0 {
				//wCh <- closeCmd
			}
			timeout.Reset(penLinPeriod)
			break
		case state = <-stataCh: //读写错误
			if false == state {
				log.Printf("喷淋读写错误\n")
				return
			}
		case <-timeout.C:
			log.Printf("电磁阀执行命令超时\n")
			devList[id].isOk = 0
			break
		}

	}
}
