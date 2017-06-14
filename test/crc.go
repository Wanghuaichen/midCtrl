package main

import (
	"devices"
	"fmt"
)

func main() {
	l, h := devices.Crc16Modbus([]byte{0, 3, 0, 0, 0, 2})
	fmt.Printf()
}
