package devices

import "testing"

func TestCrc16Modbus(t *testing.T) {
	t.Fail()
	l, h := crc16Modbus([]byte{69, 3, 0, 0, 0, 2})
	if l == 203 && h == 79 {
		t.Log("CRC校验成功\n")
	}

	l, h = crc16Modbus([]byte{255, 3, 0, 0, 0, 2}) //209 213   FF 03 00 00 00 02 D1 D5
	if l != 203 && h != 79 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
		t.Log("CRC校验成功\n")
	}

	l, h = crc16Modbus([]byte{138, 3, 4, 187, 9, 0, 0})
	if l != 53 || h != 221 {
		t.Log("CRC校验失败\n")
		t.Fail()
	}

	l, h = crc16Modbus([]byte{1, 3, 0, 141, 0, 5})
	if l == 21 && h == 226 {
		t.Log("CRC校验成功\n")
	}

	l, h = crc16Modbus([]byte{1, 3, 0, 141, 0, 5})
	if l != 21 || h != 226 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}
	//环境的CRC校验数据
	l, h = crc16Modbus([]byte{0x01, 0x03, 0x00, 0x02, 0x00, 0x01})
	if l != 0x25 || h != 0xCA {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}
}
func TestWuShuiCRC16(t *testing.T) {
	l, h := tableCRC16([]byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x06})
	if l != 0xc5 || h != 0xc8 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}
	//字符串形式是错误
	l, h = tableCRC16([]byte{0x01, 0x03, 0x04, 0x03, 0x16, 0x00, 0xF5, 0xDB, 0xF4})
	if l != 0x03 || h != 0x16 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}
	//数字模式
	l, h = tableCRC16([]byte{0x08, 0x03, 0x06, 0x00, 0x00, 0x0A, 0x02, 0x19, 0x00})
	if l != 0xAD || h != 0xE2 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}

}

/*func TestGenerateDataJsonStr(t *testing.T) {
	str := generateDataJsonStr("51000", "总电量", "4568.26")
	if str != `{"MsgType":"Devices","ID":"51000","Action":"总电量","Data":"4568.26"}` {
		t.Fatal(str)
	}
}
*/

func TestSum(t *testing.T) {
	b := []byte{0x68, 0x10, 0x44, 0x33, 0x22, 0x11, 0x00, 0x33, 0x78, 0x81, 0x16, 0x1F, 0x90, 0x00, 0x00, 0x77, 0x66, 0x55, 0x2C, 0x00, 0x77, 0x66, 0x55, 0x2C, 0x31, 0x01, 0x22, 0x11, 0x05, 0x15, 0x20, 0x21, 0x84}
	s := sum(b)
	if s != 0x08 {
		t.Fatal(s)
	}

	s = sum([]byte{0x68, 0x10, 0x69, 0x05, 0x90, 0x05, 0x15, 0x33, 0x78, 0x04, 0x04, 0xA0, 0x17, 0x01, 0x99})
	if s != 0x94 {
		t.Fatal(s)
	}
	s = sum([]byte{0x68, 0x10, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0x03, 0x03, 0x0A, 0x81, 0xA7})
	if s != 0x56 {
		t.Fatal(s)
	}
}

func TestXOR(t *testing.T) {
	//7F 00 0D 60 E2 00 51 20 37 0D 02 65 20 00 48 B2 F9
	s := []byte{0x00, 0x0D, 0x60, 0xE2, 0x00, 0x00, 0x15, 0x24, 0x19, 0x01, 0x12, 0x14, 0x80, 0x82, 0xC3}
	cx := xorVerify(s)
	if cx != 0xF9 {
		t.Log("cx=", cx)
		t.Fail()
	}
	cs := sum(s)
	if cs != 0xF9 {
		t.Log("cs=", cs)
		t.Fail()
	}
}

func TestDibangDataTrans(t *testing.T) {
	dat := "=.0036100"
	w := dibangDataTrans([]byte(dat)[1:])
	if w != "163000000" {
		t.Log("w=", w)
		t.Fail()
	}
}
