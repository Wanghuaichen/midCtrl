package comm

import (
	"encoding/binary"
	"net/url"
	"strconv"
	"time"
)

// MsgData 发个服务器的数据消息类型
type MsgData struct {
	HdID   uint       `json:"hdId"`
	Data   url.Values `json:"data"`
	Time   string     `json:"time"`
	URLStr string     //在url表中查找具体的URL
}

const (
	//MB 兆字节
	MB = 1024 * 1024
	//KB 千字节
	KB = 1024
)

// SetTime 给消息添加上当前的时间
func (msg *MsgData) SetTime() {
	msg.Time = strconv.FormatInt(time.Now().Unix(), 10)
}

var msgQ = make(chan MsgData, 10000)
var byteConter = int(0)
var msgNum = int(0)

//GetMsgNum 返回消息队列中的消息数
func GetMsgNum() int {
	return msgNum
}

// GetMsgMemory 返回消息累计占用字节
func GetMsgMemory() int {
	return msgNum
}

// SendMsg 发送消息到消息队列
func SendMsg(msg MsgData) {
	msgQ <- msg
	msgNum++
	byteConter = byteConter + binary.Size(msg)
}

// GetMsg 从消息队列获取消息
func GetMsg() (msg MsgData) {
	msg = <-msgQ
	msgNum--
	byteConter = byteConter - binary.Size(msg)
	return
}
