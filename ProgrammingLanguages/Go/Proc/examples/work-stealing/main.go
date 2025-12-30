package main

import (
	"fmt"
	"go-rem/gmp"
	"os"
)

func main() {
	os.Setenv("GOMAXPROCS", "2")
	gmp.Init()

	fmt.Println("=== 工作窃取示例（GOMAXPROCS=2）===\n")
	fmt.Println("创建 10 个 Goroutine，观察它们如何在 2 个 P 之间调度...\n")

	for i := 1; i <= 10; i++ {
		taskID := i
		gmp.Go(func() {
			result := taskID * taskID
			fmt.Printf("Task %d: 计算完成, %d * %d = %d\n", taskID, taskID, taskID, result)
		})
	}

	fmt.Printf("创建后队列中的 G 数量: %d\n\n", gmp.GetGCount())
	fmt.Println("开始调度...\n")
	gmp.Run()

	fmt.Println("\n所有任务完成！")
	fmt.Println("注意：由于工作窃取机制，任务执行顺序可能不同于创建顺序")
}
