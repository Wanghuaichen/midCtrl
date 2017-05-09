package comm

//type DevConn *net.Conn

//var devConnTable map[DevType](map[DevId]DevConn)
//var serviceAddr string
//var serviceConn net.Conn

/*
func InitComm() error {
	//建立服务器连接
	serviceConn, err := net.Dial("tcp", serviceAddr)
	if err != nil {
		log.Printf("连接服务器错:%s,%s\n", serviceAddr, err.Error())
		return err
	}
	//监听设备连接
	listen, err := net.Listen("tcp", ":7908")
	if err != nil {
		log.Printf("监听失败:%s,%s\n", ":7908", err.Error())
		return err
	}
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("监听建立连接错误:%s\n", err.Error())
			continue
		}
		go devices.HandleDeviceConn(&conn)
	}
}
*/
