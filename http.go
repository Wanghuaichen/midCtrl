package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	resp, err := http.Get("http://hao123.com")
	if err != nil {
		fmt.Println(err.Error())
	}
	buff := make([]byte, 10000)
	buff, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(buff))
	resp.Body.Close()

}
