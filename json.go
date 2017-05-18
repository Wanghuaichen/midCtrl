package main

import (
	"encoding/json"
	"fmt"
)

type jsonDev struct {
	Area         string `json:"area"`
	HardwareCode string `json:"hardwareCode"`
	HardwareID   uint   `json:"hardwareId"`
	Name         string `json:"name"`
	Port         uint   `json:"port"`
}
type jsonDevList struct {
	Code   int       `json:"code"`
	Data   []jsonDev `json:"data"`
	ErrMsg string    `json:"errMsg"`
}

func main() {
	var dat jsonDevList
	//var dat interface{}
	s := `{"code":200,"data":[{"area":"工地大门","hardwareCode":"RFID-001","hardwareId":1,"name":"RFID识别器","port":10001}],"errMsg":""}`
	err := json.Unmarshal([]byte(s), &dat)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(dat)

}
