package devices

import "comm"
import "utilies"

type DevType string
type DevId string

type DeviceInfo struct {
    id DevId
    type DevType
    mac string
}

func getDeviceInfo(mac string) (DeviceInfo,error){

}

func getDeviceMac() (macAddr string,err error){

}


func HandleDeviceConn(conn DevConn){
    defer conn.Close()
    macAddr,err:=getDeviceMac()
    if err !=nil {
        log.Printf("获取连接设备Mac地址失败:%s\n",err.Error())
    }
    devInfo, err := getDeviceInfo(macAddr)

    if err !=nil {
        log.Printf("获取连接设备信息失败:%s\n",err.Error())
    }


}
