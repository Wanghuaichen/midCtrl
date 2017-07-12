package comm

import (
	"encoding/binary"
	"log"
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

// ServCmd 服务器返回的命令
type ServCmd struct {
	HdID uint
	Cmd  int
}

const (
	//MB 兆字节
	MB = 1024 * 1024
	//KB 千字节
	KB = 1024
)
const cmdQMaxNum = 10
const msgQMaxNun = 1000
// SetTime 给消息添加上当前的时间
func (msg *MsgData) SetTime() {
	msg.Time = strconv.FormatInt(time.Now().Unix(), 10)
}

var msgQ = make(chan MsgData, msgQMaxNun)
var cmdQ = make(chan ServCmd, cmdQMaxNum)
var byteConter = int(0)
var msgNum = int(0)
var cmdQNum = int(0)

//SendCmd 将命令消息发送到命令消息队列
func SendCmd(servCmd ServCmd) {
	if(cmdQNum>=cmdQMaxNum){
		log.Printf("cmd消息队列满\n")
		return
	}
	cmdQ <- servCmd
	cmdQNum++
}

//GetCmd 从命令消息队列中取出消息
func GetCmd() (servCmd ServCmd) {
	servCmd <-cmdQ
	cmdQNum--
	return servCmd
}

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
	log.Printf("%v消息已经放入发送队列:%v\n", msg, msgNum)
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
