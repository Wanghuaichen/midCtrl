//package serv 主要负责设备和服务的沟通

package serv

import (
	"fmt"
	"io/ioutil"
	"log"
	"midCtrl/comm"
	"midCtrl/devices"
	"net/http"

	"strconv"

	"net/url"

	sjson "github.com/bitly/go-simplejson"
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
		url := devices.GetURL(msg.URLStr)
		if url == "" {
			log.Printf("获取url失败：%d\n", msg.HdID)
			continue
		}
		dat := msg.Data
		if dat == nil {
			log.Printf("数据内容为空\n")
			continue
		}
		hdID := strconv.FormatInt(int64(msg.HdID), 10)
		dat["hdId"] = []string{hdID}
		dat["time"] = []string{msg.Time}
		reqServ(url, dat)
	}

}

func reqServ(url string, dat url.Values) {
	for {
		//log.Printf("发送:%v\n", dat)
		resp, err := http.PostForm(url, dat)
		if err != nil {
			log.Printf("发送数据失败：%s\n", err.Error())
			//resp.Body.Close()
			continue
		}

		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取返回数据失败：%s\n", err.Error())
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("serv result：%s", result)
		jsDat, err := sjson.NewJson(result)
		if err != nil {
			log.Printf("返回数据非json：%s %s\n", err.Error(), result)
			return
		}
		//sjson.NewFromReader(resp.Body)
		code, err := jsDat.Get("code").Int()
		if err != nil {
			str, _ := jsDat.String()
			log.Printf("服务器返回值错误：%s %s\n", string(result), str)
			return
		}
		if 200 != code {
			log.Printf("服务器错误：%s %d\n", string(result), code)
			return
		}
		if url == devices.GetURL("设备状态") {
			cmd, err := jsDat.Get("data").Get("cmd").Int()
			if err != nil {
				log.Printf("获取服务器返回命令错误：%s\n", err.Error())
				return
			}

			//需要执行服务返回的操作
			hdID, err := strconv.ParseUint(dat["hdId"][0], 10, 32)
			if err != nil {
				log.Printf("转化hdId错误：%s\n", err.Error())
				return
			}
			//当前只有喷淋执行命令，不能一直往chanle中发数据而没有设备去读取执行，很快就会阻塞
			if hdID != 16 { //16喷淋的ID
				return
			}
			cmdData := comm.ServCmd{HdID: uint(hdID), Cmd: cmd}
			log.Printf("要执行命令：%v\n", cmdData)
			comm.SendCmd(cmdData)
		}
		break
	}
}
