package main

import (
	"fmt"
	"go-rem/gmp"
)

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════╗")
	fmt.Println("║     欢迎使用 GMP 调度器 - Go Runtime 学习项目      ║")
	fmt.Println("╚═══════════════════════════════════════════════════╝")
	fmt.Println()

	gmp.Init()
	fmt.Println("✓ 调度器已初始化")
	fmt.Println("✓ 创建 Goroutine...")
	fmt.Println()

	gmp.Go(func() {
		fmt.Println("  → Goroutine 1: 开始执行")
		count := 0
		for i := 1; i <= 5; i++ {
			count += i
		}
		fmt.Printf("  → Goroutine 1: 1+2+3+4+5 = %d\n", count)
	})

	gmp.Go(func() {
		fmt.Println("  → Goroutine 2: Hello from GMP!")
	})

	gmp.Go(func() {
		fmt.Println("  → Goroutine 3: 开始计算")
		numbers := []int{1, 2, 3, 4, 5}
		for _, n := range numbers {
			result := n * n
			fmt.Printf("  → Goroutine 3: %d² = %d\n", n, result)
		}
	})

	count := gmp.GetGCount()
	fmt.Printf("\n✓ 当前队列中有 %d 个 Goroutine 等待执行\n\n", count)
	fmt.Println("━━━━━━━━━━━━━━ 开始调度 ━━━━━━━━━━━━━━")

	gmp.Run()

	fmt.Println("\n━━━━━━━━━━━━━━ 调度完成 ━━━━━━━━━━━━━━")
	fmt.Println("✓ 所有 Goroutine 已执行完毕！")
	fmt.Println("📚 想了解更多？")
	fmt.Println("  - 快速开始: cat QUICKSTART.md")
	fmt.Println("  - 使用指南: cat HOW_TO_USE.md")
	fmt.Println("  - 运行示例: go run examples/basic/main.go")
}
