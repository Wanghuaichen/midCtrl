package config

var serviceAddr string

// InitConfig 初始化配置
func InitConfig() error {
	serviceAddr = "localhost:7010"
	return nil
}

//GetServiceAddr 取得连接服务器的地址
func GetServiceAddr() string {
	return serviceAddr
}
