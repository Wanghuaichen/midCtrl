package main

import (
	"fmt"
	"net"
	"time"
)

func compareSlice(s1 []byte, s2 []byte) bool {
	if len(s1) != len(s1) {
		return false
	}
	for i, v := range s1 {
		if v != s2[i] {
			return false
		}
	}
	return true
}
func connServ(addr string) {
	conn, err := net.Dial("tcp", "localhost:"+addr)
	if err != nil {
		fmt.Printf("连接：%s失败  %s\n", addr, err.Error())
		return
	}
	fmt.Printf("连接：%s成功\n", addr)
	switch addr {
	case "10101":
	case "10201": //电表
		go dianbiao(conn)
	case "10301":
	case "10401": //水表
		go shuibiao(conn)
	case "10501":
	case "10601": //污水
		go wushui(conn)
	case "10701":
	case "10801":
	case "10901":
	default:
	}

}

func dianbiao(conn net.Conn) {
	for {
		time.Sleep(time.Second * 1)
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取电表失败  %s\n", ok.Error())
			continue
		}
		fmt.Printf("收到：%v \n", buff[:n])
		switch {
		case compareSlice(buff[:n], []byte{1, 3, 0, 16, 0, 1, 133, 207}):
			conn.Write([]byte{177, 3, 2, 9, 96, 255, 230}) //24
		case compareSlice(buff[:n], []byte{1, 3, 0, 17, 0, 1, 212, 15}):
			conn.Write([]byte{83, 3, 2, 3, 32, 0, 160}) //800
		case compareSlice(buff[:n], []byte{1, 3, 0, 0, 0, 2, 196, 11}):
			conn.Write([]byte{69, 3, 4, 187, 245, 0, 9, 10, 231}) //6379.41
		case compareSlice(buff[:n], []byte{1, 3, 0, 8, 0, 2, 69, 201}):
			conn.Write([]byte{138, 3, 4, 187, 9, 0, 0, 53, 221}) //47.881
		}
	}
}

func shuibiao(conn net.Conn) {
	for {
		time.Sleep(time.Second * 1)
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取水表失败  %s\n", ok.Error())
			continue
		}
		fmt.Printf("收到：%v \n", buff[:n])
		switch {
		case compareSlice(buff[:n], []byte{0xFE, 0xFE, 0xFE, 0xFE, 0x68, 0x10, 0x44, 0x33, 0x22, 0x11, 0x00, 0x33, 0x78, 0x01, 0x03, 0x1F, 0x90, 0x00, 0x80, 0x16}):
			conn.Write([]byte{0xFE, 0xFE, 0xFE, 0xFE, 0x68, 0x10, 0x44, 0x33, 0x22, 0x11, 0x00, 0x33, 0x78, 0x81, 0x16, 0x1F, 0x90, 0x00, 0x00, 0x77, 0x66, 0x55, 0x2C, 0x00, 0x77, 0x66, 0x55, 0x2C, 0x31, 0x01, 0x22, 0x11, 0x05, 0x15, 0x20, 0x21, 0x84, 0x08, 0x16, 0x25, 0x26, 0x27, 0x28, 0x29, 0x30, 0x31, 0x32, 0x33, 0x34}) //24
		}
	}
}

func wushui(conn net.Conn) {
	for {
		time.Sleep(time.Second * 1)
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取水表失败  %s\n", ok.Error())
			continue
		}
		fmt.Printf("收到：%v \n", buff[:n])
		switch {
		case compareSlice(buff[:n], []byte{0x08, 0x03, 0x00, 0x00, 0x00, 0x06, 0x51, 0xc5}):
			conn.Write([]byte{0x08, 0x03, 0x06, 0x00, 0x00, 0x0A, 0x02, 0x19, 0x00, 0xAD, 0xE2})
		}
	}
}

func main() {
	port := []string{"10101", "10201", "10301", "10401", "10501", "10601", "10701", "10801", "10901"}
	for _, v := range port {
		go connServ(v)
	}
	for {

	}
}
