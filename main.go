package main

import (
	"log"
	"midCtrl/devices"
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
	if(logPath!=""){
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
	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()
	Init()
	// f, err := os.OpenFile("./tarce.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// defer f.Close()
	// err = trace.Start(f)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }
	// defer trace.Stop()
	go serv.StartMsgToServer()
	// 转发服务器的消息到设备侧处理
	//go transServMsg()

	//转发设备侧处理后的消息到主服务器
	for {
		//data := devices.GetData()
		//fmt.Printf("转发到服务器的数据：%v %s \n", data, string(data))
		//serv.SendData(data)
		time.Sleep(time.Minute * 1)
		log.Println("run main")
	}
}
