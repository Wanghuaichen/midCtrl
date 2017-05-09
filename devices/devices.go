package devices

import (
	"encoding/json"
	"net"
)

// DianBiaoHanldeMsg 电表的消息处理
func DianBiaoHanldeMsg(msg string) {
	json.Unmarshal(msg, data)
	switch data["action"] {
	case "每日电量":
	case "总电量":
	}
}

func get每日电量() {

}

func devSend(conn net.Conn, data []byte)
{
    conn.Write(data)
}
