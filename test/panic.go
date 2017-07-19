package main

import (
	"fmt"
	"time"
)

func p() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("err:%s\n", err)
		}
	}()
	panic("挂了")
	//go k()
}
func k() {
	panic("死了")
}
func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("err:%s\n", err)
		}
	}()

	go p()
	for {
		fmt.Println("kkkkkkk")
		time.Sleep(time.Second * 5)
	}
}
