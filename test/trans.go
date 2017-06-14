package main

import (
	"fmt"
)

func main() {
	v := transData(0xFF3D)
	fmt.Println(v)
}

func transData(dat uint16) int16 {
	if 0x8000&dat == 0x8000 { //dat 最高为1
		dat = dat & 0x7FFF
		dat = (^dat & 0x7FFF) + 1
		return int16(-dat)
	}
	return int16(dat)
}
