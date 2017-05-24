package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
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
	case "10701":
	case "10801":
	case "10901":
	default:
	}

}
func rfid(conn net.Conn) {
	log.Println("RFID开始发送数据")
	rfidData := []string{
		"000D60E2005142051101362850015935",
		"000D60E20051420511013624101C4935",
		"000D60E2005142051101362470196F35",
		"000D60E2005142051101362530133D35",
		"000D60E20051420511013626000FF035",
		"000D60E20051420511013624201C4A35",
		"000D60E2005142051101362480197035",
		"000D60E2005142051101362540133E35",
		"000D60E20051420511013626100D9535",
		"000D60E200514205110136217031F135",
		"000D60E20051420511013622302E8F35",
		"000D60E2005142051101362290272535",
		"000D60E2005142051101362350232735",
		"000D60E200514205110136218031F235",
		"000D60E20051420511013622402E9035",
		"000D60E2005142051101362300272635",
		"000D60E2005142051101362360232835",
		"000D60E20051420511013619304A1135",
	}
	len := len(rfidData)
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		i := r.Intn(len)
		conn.Write([]byte(rfidData[i]))
		time.Sleep(time.Millisecond * 200)
	}
}
func diBangD39(conn net.Conn) {
	log.Println("地磅开始发送数据")
	for {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		times := r.Intn(10)
		dat := []string{"23.543=", ".123=", "546.7=", "86.45=", "546.=", "23.5493=", ".1273=", "5446.7=", "86.495=", "58746.="}
		for i := times; times > 0; times-- {
			conn.Write([]byte(dat[i]))
		}
		time.Sleep(time.Millisecond * 500)
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
		case compareSlice(buff[:n], []byte{1, 3, 0, 16, 0, 1, 133, 207}):
			data = []byte{177, 3, 2, 9, 96, 255, 230}
			conn.Write(data) //24
		case compareSlice(buff[:n], []byte{1, 3, 0, 17, 0, 1, 212, 15}):
			data = []byte{83, 3, 2, 3, 32, 0, 160}
			conn.Write(data) //800
		case compareSlice(buff[:n], []byte{1, 3, 0, 0, 0, 2, 196, 11}):
			data = []byte{69, 3, 4, 187, 245, 0, 9, 10, 231}
			conn.Write(data) //6379.41
		case compareSlice(buff[:n], []byte{1, 3, 0, 8, 0, 2, 69, 201}):
			data = []byte{138, 3, 4, 187, 9, 0, 0, 53, 221}
			conn.Write(data) //47.881
		}
		log.Printf("电表发送：%v\n", data)
	}
}

func shuibiao(conn net.Conn) {
	log.Println("水表开始获取数据")
	for {
		time.Sleep(time.Second * 1)
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
			data := []byte{0x01, 0x03, 0x04, 0x03, 0x16, 0x00, 0xF5, 0xDB, 0xF4, 0x03, 0x16}
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

func main() {
	InitLoger()
	port := []string{"10101", "10102", "10103", "10104", "10201", "10301", "10401", "10501", "10601", "10701", "10801", "10901"}
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
