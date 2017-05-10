package devices

// DianBiaoHandleMsg 电表的消息处理
func DianBiaoHandleMsg(id string, action string) {
	switch action {
	case "每日电量":
		get每日电量(id)
	case "总电量":
	}
}

func get每日电量(id string) {
	conn := getConn(id)
	(*conn).Write([]byte("请求今天电量"))
}
