// 协议说明
/*
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
package devices

import "bufio"
import "strconv"
import "net/url"

//等待rfid上报数据
func rfidDataHandle(id uint) {
	conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		return
	}
	tryTimes := 0
	for {
		buff := bufio.NewReader(conn)
		dat, err := buff.ReadBytes(0x7F)
		if err != nil {
			tryTimes++
			if tryTimes > 5 {
				unBindConn(id)
				return
			}
			continue
		}
		tryTimes = 0
		if dat[0] != 0x00 && dat[1] != 0x0D && dat[2] != 0x60 { //id未设置为0，数据长度为13=0xD,0x60上报命令
			continue
		}
		userID := bytesToString(dat[3:15])
		rfid := url.Values{"rfid": {userID}}
		sendData("RFID", id, rfid) //发送数据给服务器
	}
}

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
		str += strconv.FormatInt(int64(b), 16)
	}
	return str
}
