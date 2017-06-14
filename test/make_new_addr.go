package main

import (
	"fmt"
)

func main() {
	for i := 1; i < 3; i++ {
		m := make(map[string]string, 3)
		m["34"] = "56"
		m["78"] = "90"

		fmt.Printf("m地址:%v\n", &m)
	}
}
