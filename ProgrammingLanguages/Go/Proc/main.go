package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, world!")

	go func() {
		fmt.Println("Hello, goroutine world!")
		time.Sleep(5 * time.Second)
	}()

	time.Sleep(6 * time.Second)
}
