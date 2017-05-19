//package serv 主要负责设备和服务的沟通

package serv

import (
	"comm"
	"devices"
	"fmt"
	"io/ioutil"
	"log"
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
		hdID := strconv.FormatInt(int64(msg.HdID), 10)
		dat["hdId"] = []string{hdID}
		dat["time"] = []string{msg.Time}
		reqServ(url, dat)
	}

}

func reqServ(url string, dat url.Values) {
	for {
		fmt.Printf("发送:%v\n", dat)
		resp, err := http.PostForm(url, dat)
		if err != nil {
			log.Printf("发送数据失败：%s\n", err.Error())
			resp.Body.Close()
			continue
		}

		result, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("读取返回数据失败：%s\n", err.Error())
		}
		defer resp.Body.Close()
		//fmt.Printf("%s", result)
		jsDat, err := sjson.NewJson(result)
		if err != nil {
			log.Printf("返回数据非json：%s\n", err.Error())
		}
		//sjson.NewFromReader(resp.Body)
		code, err := jsDat.Get("code").Int()
		if err != nil {
			str, _ := jsDat.String()
			log.Printf("服务器返回值错误：%s\n", str)
		} else if 200 != code {
			log.Printf("服务器错误：%d\n", code)
			continue
		}
		break
	}
}
