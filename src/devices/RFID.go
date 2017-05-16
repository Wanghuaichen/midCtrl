// 协议说明
/*
串口设置：RS485/RS232 物理层   1 位起始位、 8 位数据位、 1 位停止位、 无奇偶校验， 半双工；通信波特率： 9600bps。


帧起始同步字节域   设备ID域   数据长度域   命令字       数据内容域      异或校验码域
0x7F(1Byte)       (1Byte)   N(1Byte)    1Byte        N Bytes        XOR (1 Byte)



上报卡号0x60（由控制器主动发送）
FrameHeader  ID       Length    Command Code       Data     CheckXOR
0X7F 		 0XFF     xx        0x60               xxx        xor
*/

package devices

func xorVerify(dat []byte) byte {
	xor := byte(0x0)
	for _, v := range dat {
		xor = xor ^ v
	}
	return xor
}
