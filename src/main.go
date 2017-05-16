package main

import (
	"config"
	"devices"
	"log"
	"os"
)

// serviceConn 和主服务器的连接
// var serviceConn net.Conn

// serviceAddr 主服务器地址
var serviceAddr string

// InitLoger 初始化log配置
func InitLoger() error {
	file, err := os.OpenFile("./log.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		return err
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(file)
	return nil

}

// Init 初始化程序
func Init() {
	config.InitConfig()
	InitLoger()
	//serv.InitServ()
	devices.IntiDevice()
}

/*func transServMsg() {
	for {
		data := serv.GetData()
		fmt.Printf("转发到设备侧的数据：%v %s \n", data, string(data))
		devices.SendData(data)
	}
}*/
func main() {
	Init()
	// 转发服务器的消息到设备侧处理
	//go transServMsg()

	//转发设备侧处理后的消息到主服务器
	for {
		//data := devices.GetData()
		//fmt.Printf("转发到服务器的数据：%v %s \n", data, string(data))
		//serv.SendData(data)
	}
}
