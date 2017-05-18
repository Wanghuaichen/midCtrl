package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println(time.Now().Unix())
goon:
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			if j > 5 {
				break goon
			}
			fmt.Printf("i=%d,j=%d\n", i, j)
		}
	}

}
