package httpServ

import (
	"fmt"
	"html/template"
	"log"
	"midCtrl/devices"
	"net/http"
)

const logTemp = `
`

func showDevicesStatus(w http.ResponseWriter, r *http.Request) {
	//log.Println("show devices status")
	HTMLTemp, err := template.New("show devices status").Parse(statusTemp)
	if err != nil {
		log.Printf("解析设备状态显示模板错误：%s\n", err.Error())
	}
	list := devices.GetDevStatusInfo()

	if list == nil {
		fmt.Fprintf(w, "当前没有设备")
		return
	}
	//fmt.Printf("%v\n", list)
	HTMLTemp.Execute(w, list)
}

func showLog() {

}

// ServStart 启动http服务器
func ServStart() {
	log.Println("http server statr")
	http.HandleFunc("/", showDevicesStatus)  //设置访问的路由
	err := http.ListenAndServe(":9090", nil) //设置监听的端口
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
