package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

func compareSlice(s1 []byte, s2 []byte) bool {
	if len(s1) != len(s2) {
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
	//conn, err := net.Dial("tcp", "39.108.5.184:"+addr)
	conn, err := net.Dial("tcp", "localhost:"+addr)
	if err != nil {
		fmt.Printf("连接：%s失败  %s\n", addr, err.Error())
		return
	}
	fmt.Printf("连接：%s成功\n", addr)
	switch addr {
	case "10101": //RFID
		go rfid(conn)
	case "10102":
		go rfid(conn)
	case "10103":
		go rfid(conn)
	case "10104":
		go rfid(conn)
	case "10201": //电表
		go dianbiao(conn)
	case "10301":
	case "10401": //水表
		go shuibiao(conn)
	case "10501": //地磅
		go diBangD39(conn)
	case "10601": //污水
		go wushui(conn)
	case "10602": //污水
		go wushui(conn)
	case "10701": //环境
		go huanjing(conn)
	case "10801": //塔吊
		//tadiao(conn)
	case "10901":
	case "11601":
	case "11201":
	default:
	}

}
func rfid(conn net.Conn) {
	log.Println("RFID开始发送数据")
	rfidData := []string{
		"7F000D60E2005120370D011520004783FD",
		"7F000D60E2005120370D01132000478235",
		"7F000D60E2005120370D01122000477A35",
		"7F000D60E2005120370D01112000478135",
		"7F000D60E2005120370D01102000477935",
	}
	len := len(rfidData)
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		i := r.Intn(len)
		rd := stringToByte(rfidData[i])
		fmt.Printf("发送RFID:%x\n", rd)
		conn.Write(rd)
		time.Sleep(time.Millisecond * 100000)
	}
}
func stringToByte(dat string) []byte {
	r := make([]byte, 0, 17)
	for i := 0; i < len(dat); i += 2 {
		hex, _ := strconv.ParseUint(dat[i:i+2], 16, 8)
		r = append(r, uint8(hex))
	}
	return r
}
func diBangD39(conn net.Conn) {
	log.Println("地磅开始发送数据")
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		times := r.Intn(6)
		index := r.Intn(10)
		dat := []string{"23.54300=", ".1230000=", "546.7000=", "86.45000=", "546.0000=", "23.54930=", ".1273000=", "5446.700=", "86.49500=", "58746.00="}
		for i := index; index > 0; index-- {
			for k := times; k > 0; k-- {
				conn.Write([]byte(dat[i]))
				time.Sleep(time.Millisecond * 300)
			}
		}
	}
}
func znDiBang(conn net.Conn) {
	log.Println("移动地磅开始发送数据")
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		times := r.Intn(6)
		index := r.Intn(10)
		dat := []string{"=23.54300", ".1230000=", "546.7000=", "86.45000=", "546.0000=", "23.54930=", ".1273000=", "5446.700=", "86.49500=", "58746.00="}
		for i := index; index > 0; index-- {
			for k := times; k > 0; k-- {
				conn.Write([]byte(dat[i]))
				time.Sleep(time.Millisecond * 300)
			}
		}
	}
}
func dianbiao(conn net.Conn) {
	log.Println("电表开始获取数据")
	for {
		//time.Sleep(time.Second * 1)
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取电表失败  %s\n", ok.Error())
			continue
		}
		log.Printf("电表收到：%v \n", buff[:n])
		var data []byte
		switch {
		case compareSlice(buff[:n], []byte{0x01, 0x03, 0x00, 0x09, 0x00, 0x01, 0x14, 0x09}):
			data = []byte{177, 3, 2, 9, 96, 255, 230}
			conn.Write(data) //24
		case compareSlice(buff[:n], []byte{1, 3, 0, 17, 0, 1, 212, 15}):
			data = []byte{83, 3, 2, 3, 32, 0, 160}
			conn.Write(data) //800
		case compareSlice(buff[:n], []byte{0x01, 0x03, 0x00, 0x09, 0x00, 0x02, 0x14, 0x09}):
			data = []byte{0x01, 0x03, 0x04, 0x0A, 0x00, 0x02, 0x96, 0x78, 0xE5}
			conn.Write(data) //6379.41
		case compareSlice(buff[:n], []byte{0x01, 0x03, 0x00, 0x00, 0x0A, 0x02, 0x14, 0x09}):
			data = []byte{0x01, 0x03, 0x04, 0x04, 0x03, 0x02, 0x16, 0x8B, 0xAD}
			conn.Write(data) //47.881
		}
		log.Printf("电表发送：%v\n", data)
	}
}

func shuibiao(conn net.Conn) {
	log.Println("水表开始获取数据")
	for {
		//time.Sleep(time.Second * 1)
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取水表失败  %s\n", ok.Error())
			continue
		}
		fmt.Printf("水表收到：%v \n", buff[:n])
		switch {
		case compareSlice(buff[:n], []byte{0xFE, 0xFE, 0x68, 0x10, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0xAA, 0x01, 0x03, 0x90, 0x1F, 0xAA, 0x7B, 0x16}):
			conn.Write([]byte{0xFE, 0xFE, 0x68, 0x10, 0x45, 0x41, 0x10, 0x05, 0x15, 0x33, 0x78, 0x81, 0x16, 0x90, 0x1F, 0xAA, 0x00, 0x59, 0x59, 0x00, 0x2C, 0xFF, 0xFF, 0xFF, 0xFF, 0x2C, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0xC2, 0x16}) //00595900
			conn.Write([]byte{0xFE, 0xFE, 0x68, 0x10, 0x45, 0x41, 0x10, 0x05, 0x15, 0x33, 0x78, 0x81, 0x16, 0x90, 0x1F, 0xAA, 0x00, 0x46, 0x60, 0x00, 0x2C, 0xFF, 0xFF, 0xFF, 0xFF, 0x2C, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0xB6, 0x16})
		}
	}
}

func wushui(conn net.Conn) {
	log.Println("污水开始获取数据")
	for {
		time.Sleep(time.Second * 1)
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取水表失败  %s\n", ok.Error())
			continue
		}
		fmt.Printf("污水收到：%v \n", buff[:n])
		switch {
		case compareSlice(buff[:n], []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x06, 0xc5, 0xc8}): //01 03 00 00 00 06 c5 c8
			data := []byte{0x01, 0x03, 0x04, 0x03, 0x16, 0x00, 0xF5, 0xDB, 0xF4}
			n, err := conn.Write(data)
			if err != nil {
				log.Printf("发送污水失败：%d,%s\n", n, err.Error())
			}
			log.Printf("发送污水数据:%v\n", data)
		}
	}
}

func tadiao(conn net.Conn) {
	for {
		data := []byte{'$', 'R', 0x00, 0x00, '0', '1', '2', '3', '4', '5', '6', '7', '2', '3', '4', '5', '6', '7', '2', 0x4, 0x43, 0x54}
		n, err := conn.Write(data)
		if err != nil {
			log.Printf("发送塔吊失败：%d,%s\n", n, err.Error())
		}
		log.Printf("发送塔吊数据:%v\n", data)
		time.Sleep(1 * time.Second)
	}
}

func huanjing(conn net.Conn) {
	log.Println("环境开始获取数据")
	for {
		buff := make([]byte, 1024)
		n, ok := conn.Read(buff)
		if ok != nil {
			fmt.Printf("读取环境失败  %s\n", ok.Error())
			continue
		}
		fmt.Printf("环境收到：%v \n", buff[:n])
		switch {
		case compareSlice(buff[:n], []byte{0x01, 0x03, 0x00, 0x00, 0xF1, 0xD8}):
			conn.Write([]byte{0x01, 0x03, 0x00, 0x40, 0x00, 0x12, 0x02, 0xD0, 0x00, 0x2B, 0x00, 0x36, 0x01, 0x22, 0x02, 0x86, 0x00, 0x84, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x7F, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xEF, 0x1C})
		}
	}
}
func main() {
	InitLoger()
	port := []string{"10101", "10102", "10103", "10104", "10201", "10301", "10401", "10501", "10601", "10701", "10801", "10901", "11601", "11201"}
	for _, v := range port {
		go connServ(v)
	}
	for {
		time.Sleep(time.Second * 100)
		fmt.Println("main run")
	}
}

// InitLoger 初始化log配置
func InitLoger() error {
	file, err := os.OpenFile("./logClent.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(file)
	return nil

}
