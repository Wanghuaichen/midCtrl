//package serv 主要负责设备和服务的沟通

package serv

import (
	"bytes"
	"comm"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// serviceConn 和主服务器的连接
// var serviceConn net.Conn

// serviceAddr 主服务器地址
//var serviceAddr string

// recvCh 客户端先将数据发送到servCh通道，然后由HandleServ处理发给服务器

//StartMsgToServer 消息发送到服务器
func StartMsgToServer() {
	for {
		msg := comm.GetMsg()
		jsonData, err := json.Marshal(msg)
		if err != nil {
			log.Printf("转化发送数据%v错误：%s\n", msg, err.Error())
		}
		dat := bytes.NewBuffer(jsonData)
		for {
			fmt.Printf("发送:%s\n", dat.String())
			resp, err := http.Post("http://39.108.5.184/smart/api/saveElectricityData", "application/json;charset=utf-8", dat)
			if err != nil {
				log.Printf("发送数据失败：%s\n", err.Error())
				resp.Body.Close()
				continue
			}

			result, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("读取返回数据失败：%s\n", err.Error())
			}
			fmt.Printf("%s", result)
			resp.Body.Close()
			break
		}
	}
}
