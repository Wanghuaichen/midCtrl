package main

import (
	"log"
	"midCtrl/devices"
	"midCtrl/httpServ"
	"midCtrl/serv"
	"os"
	"time"
)

// serviceConn 和主服务器的连接
// var serviceConn net.Conn

// serviceAddr 主服务器地址
// var serviceAddr string

// InitLoger 初始化log配置
func InitLoger(logPath string) error {
	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
		if err != nil {
			return err
		}
		log.SetOutput(file)
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil

}

// Init 初始化程序
func Init() {
	//config.InitConfig()
	InitLoger("")
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
	//启动http server
	go httpServ.ServStart()
	// 转发服务器的消息到设备侧处理
	go serv.StartMsgToServer()
	for {
		//data := devices.GetData()
		//fmt.Printf("转发到服务器的数据：%v %s \n", data, string(data))
		//serv.SendData(data)
		time.Sleep(time.Minute * 1)
		log.Println("run main")
	}
}
