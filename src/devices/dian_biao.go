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
