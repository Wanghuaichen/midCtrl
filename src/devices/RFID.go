// 协议说明
/*
串口设置：RS485/RS232 物理层   1 位起始位、 8 位数据位、 1 位停止位、 无奇偶校验， 半双工；通信波特率： 9600bps。


帧起始同步字节域   设备ID域   数据长度域   命令字       数据内容域      异或校验码域
0x7F(1Byte)       (1Byte)   N(1Byte)    1Byte        N Bytes        XOR (1 Byte)

10401

上报卡号0x60（由控制器主动发送）
FrameHeader  ID       Length    Command Code       Data     CheckXOR
0X7F 		 0XFF     xx        0x60               xxx        xor
*/
/*
7F 00 0D 60 E2 00 00 15 67 16 02 17 03 30 EF 52 F8 7F 00 0D 60 E2 00 00 15 67 16 02 17 03 30 EF 52 F8
E2 00 00 15 67 16 02 17 03 30 EF
E2 00 00 15 24 03 02 18 04 10 EA


7F 00 0D 60 E2 00 00 15 24 03 02 18 04 10 EA 6C FD
E2 00 00 15 24 03 02 18 04 10 EA
A105031007195


7F 00 0D 60 E2 00 00 15 24 19 01 12 14 80 82 C3 F9
E2 00 00 15 24 19 01 12 14 80 82
A105031007191

7F 00 0D 60 E2 00 00 15 24 03 01 53 06 30 D7 30 FE
E2 00 00 15 24 03 01 53 06 30 D7
A105031007192
*/
package devices

import "log"

//等待rfid上报数据
func rfidDataHandle(id uint) {
	conn := getConn(id)
	defer conn.Close()
	if conn == nil {
		return
	}
	buf := make([]byte, 0, 1024)
	len := 0 //buf中最后数据位置
	//sIdx := 0 //buf中开始数据位置
GO_ON_READ:
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("读数据错误：%s\n", err.Error())
			continue GO_ON_READ
		}
		len = len + n

	}
}

func xorVerify(dat []byte) byte {
	xor := byte(0x0)
	for _, v := range dat {
		xor = xor ^ v
	}
	return xor
}
