package main

import (
	"config"
	"devices"
	"log"
	"os"
	"serv"
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

func Init() {
	config.InitConfig()
	InitLoger()
	serv.InitServ()
	devices.IntiDevice()
}
func transServMsg() {
	for {
		devices.SendData(serv.GetData())
	}
}
func main() {
	Init()
	// 转发服务器的消息到设备侧处理
	go transServMsg()

	//转发设备侧处理后的消息到主服务器
	for {
		serv.SendData(devices.GetData())
	}
}
