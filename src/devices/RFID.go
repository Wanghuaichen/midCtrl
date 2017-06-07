package devices

/*
协议说明
串口设置：RS485/RS232 物理层   1 位起始位、 8 位数据位、 1 位停止位、 无奇偶校验， 半双工；通信波特率： 9600bps。


帧起始同步字节域   设备ID域   数据长度域   命令字       数据内容域      异或校验码域
0x7F(1Byte)       (1Byte)   N(1Byte)    1Byte        N Bytes        XOR (1 Byte)

10401

上报卡号0x60（由控制器主动发送）
FrameHeader  ID       Length    Command Code       Data     CheckXOR
0X7F 		 0XFF     xx        0x60               xxx        xor

7F 00 0D 60 E2 00 00 15 67 16 02 17 03 30 EF 52 F8
7F 00 0D 60 E2 00 00 15 67 16 02 17 03 30 EF 52 F8
E2 00 00 15 67 16 02 17 03 30 EF 52
E2 00 00 15 24 03 02 18 04 10 EA 6C


7F 00 0D 60 E2 00 00 15 24 03 02 18 04 10 EA 6C FD
E2 00 00 15 24 03 02 18 04 10 EA 6C
A105031007195


7F 00 0D 60 E2 00 00 15 24 19 01 12 14 80 82 C3 F9
E2 00 00 15 24 19 01 12 14 80 82 C3
A105031007191

7F 00 0D 60 E2 00 00 15 24 03 01 53 06 30 D7 30 FE
E2 00 00 15 24 03 01 53 06 30 D7 30
A105031007192

E2 00 51 42 05 11 01 35 20 30 41 CF
*/

import (
	"log"
	"net/url"
	"strconv"
	"strings"
)

/*func rfidStart(id uint) {
	go rfidDataHandle(id)
	log.Printf("RFID:%d  开始接受数据\n", id)
}*/

func rfidStart(id uint) {
	conn := getConn(id)
	if conn == nil {
		log.Printf("%d RFID获取连接失败\n", id)
		return
	}
	defer func() {
		conn.Close() //关闭连接
		log.Printf("RFID监测处理发生错误\n")
		unBindConn(id)
		//设置设备状态
	}()

	rCh := make(chan []byte, 10)
	stataCh := make(chan bool, 1)
	//fmt.Printf("RFID接受%d数据%v\n", id, conn)
	go readOneData(conn, rCh, []byte{0x7F, 0x00, 0x0D, 0x60}, 17, stataCh)
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
		//fmt.Printf("RFID:%v\n", dat)
		userID := bytesToString(dat[4:16])
		rfid := url.Values{"rfid": {userID}}
		//fmt.Printf("RFID发送%d：%v\n", id, rfid)
		sendData("RFID", id, rfid) //发送数据给服务器
	}
}

/*
//等待rfid上报数据
func rfidDataHandle(id uint) {
	conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		log.Printf("RFID conn不存在")
		return
	}
	buf := make([]byte, 1024)
	len := 0  //buf中最后数据位置
	sIdx := 0 //buf中开始数据位置
	tryTimes := 0
RFID_GO_ON_READ:
	for {
		//buff := bufio.NewReader(conn)
		//dat, err := buff.ReadBytes(byte(0x7F))
		n, err := conn.Read(buf[len:])
		if err != nil {
			log.Printf("读数据错误：%s\n", err.Error())
			if tryTimes > 3 {
				unBindConn(id)
				return
			}
			tryTimes++
			continue
		}
		tryTimes = 0
		len = len + n
		for {
			sIndex := bytes.IndexByte(buf[sIdx:len], byte(0x75)) //帧开始位置
			if sIndex == -1 {
				log.Printf("数据中未找到包起始符号:%v  sIdex:%d  len:%d\n", buf[sIdx:len], sIdx, len)
				len = 0
				sIdx = 0
				continue RFID_GO_ON_READ
			}
			//下一组
			if sIndex+17 > len {
				//len = 0
				//sIdx = 0
				continue RFID_GO_ON_READ
			}
			dat := buf[sIndex+1 : sIndex+17]
			//log.Printf("RFID读到数据:%x\n", dat)

			if dat[0] != 0x00 && dat[1] != 0x0D && dat[2] != 0x60 { //id未设置为0，数据长度为13=0xD,0x60上报命令
				log.Printf("RFID 数据格式错误")
				continue
			}
			userID := bytesToString(dat[3:15])
			rfid := url.Values{"rfid": {userID}}
			fmt.Printf("RFID发送：%v\n", rfid)
			sendData("RFID", id, rfid) //发送数据给服务器
			sIdx += 17
			if sIdx == len {
				len = 0
				sIdx = 0
				continue RFID_GO_ON_READ
			}
		}
	}
}
*/
func xorVerify(dat []byte) byte {
	xor := byte(0x0)
	for _, v := range dat {
		xor = xor ^ v
	}
	return xor
}

func bytesToString(dat []byte) string {
	str := ""
	for _, b := range dat {
		ch := strconv.FormatInt(int64(b), 16)
		if len(ch) == 1 {
			str += "0"
		}
		str += ch
	}
	str = strings.ToUpper(str)
	return str
}
