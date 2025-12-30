package main

import (
	"fmt"
	"go-rem/gmp"
)

func main() {
	gmp.Init()
	fmt.Println("=== GMP 调度器示例 ===\n")

	gmp.Go(func() {
		fmt.Println("Goroutine 1: Hello from G1!")
	})

	gmp.Go(func() {
		fmt.Println("Goroutine 2: Hello from G2!")
	})

	gmp.Go(func() {
		fmt.Println("Goroutine 3: Computing...")
		sum := 0
		for i := 1; i <= 10; i++ {
			sum += i
		}
		fmt.Printf("Goroutine 3: Sum of 1-10 = %d\n", sum)
	})

	fmt.Println("开始调度...\n")
	gmp.Run()
	fmt.Println("\n所有 Goroutine 执行完毕！")
}
