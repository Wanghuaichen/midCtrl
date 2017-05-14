package devices

import "testing"

func TestCrc16Modbus(t *testing.T) {
	l, h := crc16Modbus([]byte{69, 3, 0, 0, 0, 2})
	if l == 203 && h == 79 {
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
}
func TestWuShuiCRC16(t *testing.T) {
	l, h := wuShuiCRC16([]byte{0x08, 0x03, 0x00, 0x00, 0x00, 0x06})
	if l != 0x51 || h != 0xc5 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}
	l, h = wuShuiCRC16([]byte{0x08, 0x03, 0x0B, 0x30, 0x36, 0x2E, 0x38, 0x37, 0x39, 0x20, 0x32, 0x35, 0x2E, 0x30, 0x00})
	if l != 0xD6 || h != 0x04 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}
	l, h = wuShuiCRC16([]byte{0x08, 0x03, 0x06, 0x00, 0x00, 0x0A, 0x02, 0x19, 0x00})
	if l != 0xAD || h != 0xE2 {
		t.Log("CRC校验失败:", l, h)
		t.Fail()
	}

}
func TestGenerateDataJsonStr(t *testing.T) {
	str := generateDataJsonStr("51000", "总电量", "4568.26")
	if str != `{"MsgType":"Devices","ID":"51000","Action":"总电量","Data":"4568.26"}` {
		t.Fatal(str)
	}
}
