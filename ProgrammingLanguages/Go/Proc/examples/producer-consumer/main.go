package main

import (
	"fmt"
	"go-rem/gmp"
)

var sharedData []int

func main() {
	gmp.Init()
	fmt.Println("=== 生产者-消费者模式示例 ===\n")

	for i := 1; i <= 3; i++ {
		producerID := i
		gmp.Go(func() {
			value := producerID * 10
			sharedData = append(sharedData, value)
			fmt.Printf("生产者 %d: 生产了数据 %d\n", producerID, value)
		})
	}

	for i := 1; i <= 2; i++ {
		consumerID := i
		gmp.Go(func() {
			if len(sharedData) > 0 {
				fmt.Printf("消费者 %d: 准备消费数据\n", consumerID)
			}
		})
	}

	fmt.Println("开始调度...\n")
	gmp.Run()

	fmt.Printf("\n最终数据: %v\n", sharedData)
	fmt.Println("所有任务完成！")
}
