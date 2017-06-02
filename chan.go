package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan bool, 1)
	go send(ch)
	<-ch //空就一直等待
	fmt.Println("读到数据")
}

func send(ch chan<- bool) {
	time.Sleep(time.Second * 20)
	fmt.Println("发送数据")
	ch <- true
}
