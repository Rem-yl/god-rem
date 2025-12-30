# GMP 调度器使用示例

本目录包含了如何在外部程序中使用 GMP 调度器的示例。

## 快速开始

### 1. 基础示例（basic）

最简单的使用方式，展示如何创建和运行 Goroutine。

```bash
cd examples/basic
go run main.go
```

**输出示例：**
```
=== GMP 调度器示例 ===

开始调度...

Goroutine 2: Hello from G2!
Goroutine 3: Computing...
Goroutine 3: Sum of 1-10 = 55
Goroutine 1: Hello from G1!

所有 Goroutine 执行完毕！
```

### 2. 生产者-消费者示例（producer-consumer）

展示多个 Goroutine 协作的场景。

```bash
cd examples/producer-consumer
go run main.go
```

### 3. 工作窃取示例（work-stealing）

展示 GMP 的负载均衡和工作窃取机制。

```bash
cd examples/work-stealing
go run main.go
```

## API 使用说明

### 核心 API

```go
package main

import "go-rem/gmp"

func main() {
    // 1. 初始化调度器（必须首先调用）
    gmp.Init()

    // 2. 创建 Goroutine（类似 go func()）
    gmp.Go(func() {
        // 你的代码
    })

    // 3. 启动调度器
    gmp.Run()
}
```

### 完整 API

| 函数 | 说明 |
|------|------|
| `gmp.Init()` | 初始化 GMP 调度器，必须首先调用 |
| `gmp.Go(fn func())` | 创建新的 Goroutine 执行 fn |
| `gmp.Run()` | 启动调度器，运行所有 Goroutine |
| `gmp.GetGCount()` | 获取当前队列中 G 的数量（调试用）|

### 环境变量

- **GOMAXPROCS**: 设置 P 的数量（默认为 1）
  ```bash
  GOMAXPROCS=4 go run main.go
  ```

## 使用模式

### 模式 1: 简单并发任务

```go
gmp.Init()

// 创建多个独立任务
for i := 0; i < 10; i++ {
    taskID := i
    gmp.Go(func() {
        fmt.Printf("Task %d running\n", taskID)
    })
}

gmp.Run()
```

### 模式 2: 带闭包的任务

```go
gmp.Init()

data := []int{1, 2, 3, 4, 5}

for _, value := range data {
    v := value // 注意：捕获循环变量
    gmp.Go(func() {
        result := v * v
        fmt.Printf("%d^2 = %d\n", v, result)
    })
}

gmp.Run()
```

### 模式 3: 嵌套创建 Goroutine

```go
gmp.Init()

gmp.Go(func() {
    fmt.Println("Parent goroutine")

    // 在 Goroutine 内创建新的 Goroutine
    gmp.Go(func() {
        fmt.Println("Child goroutine")
    })
})

gmp.Run()
```

## 与标准 Go 的对比

| 标准 Go | GMP 调度器 |
|---------|-----------|
| `go func() { ... }` | `gmp.Go(func() { ... })` |
| 自动启动 | 需要调用 `gmp.Init()` 和 `gmp.Run()` |
| 运行时调度 | 显式调度 |
| 真正的 OS 线程 | 简化的单线程调度 |

## 注意事项

### ⚠️ 重要限制

1. **必须调用 Init()**
   ```go
   gmp.Init() // 必须首先调用
   gmp.Go(func() { ... })
   ```

2. **Run() 会阻塞**
   ```go
   gmp.Run() // 阻塞直到所有 G 执行完
   fmt.Println("这行会在所有 G 完成后执行")
   ```

3. **简化实现的限制**
   - 不支持 channel
   - 不支持 select
   - 不支持真正的并发（单线程调度）
   - 没有抢占机制
   - 需要手动调用 Run()

4. **闭包变量捕获**
   ```go
   // ❌ 错误：循环变量问题
   for i := 0; i < 10; i++ {
       gmp.Go(func() {
           fmt.Println(i) // 可能全部打印 10
       })
   }

   // ✅ 正确：捕获变量
   for i := 0; i < 10; i++ {
       i := i // 创建新变量
       gmp.Go(func() {
           fmt.Println(i) // 正确打印 0-9
       })
   }
   ```

## 调试技巧

### 查看队列状态

```go
gmp.Init()

// 创建一些 G
for i := 0; i < 5; i++ {
    gmp.Go(func() { /* ... */ })
}

// 查看队列中的 G 数量
count := gmp.GetGCount()
fmt.Printf("当前队列中有 %d 个 G\n", count)

gmp.Run()
```

### 设置多个 P 观察工作窃取

```bash
# 在运行前设置环境变量
export GOMAXPROCS=4
go run main.go
```

或在代码中设置：
```go
import "os"

os.Setenv("GOMAXPROCS", "4")
gmp.Init()
```

## 学习建议

1. **从简单示例开始**
   - 先运行 `basic` 示例理解基本用法
   - 观察输出理解执行顺序

2. **实验不同的 GOMAXPROCS**
   - 尝试 1, 2, 4 个 P
   - 观察任务执行顺序的变化

3. **阅读源码**
   - 查看 `src/gmp/api.go` 理解 API 实现
   - 对比 `src/gmp/proc_rem.go` 理解调度逻辑

4. **创建自己的示例**
   - 修改现有示例
   - 创建新的使用场景

## 常见问题

### Q: 为什么我的 Goroutine 没有执行？

A: 确保调用了 `gmp.Run()`：
```go
gmp.Init()
gmp.Go(func() { fmt.Println("Hello") })
gmp.Run() // 必须调用！
```

### Q: 如何控制 Goroutine 的执行顺序？

A: GMP 调度器不保证执行顺序。如果需要顺序执行，请按顺序调用或使用同步机制。

### Q: 可以在 Goroutine 内创建新的 Goroutine 吗？

A: 可以！但要确保在 `gmp.Run()` 之前创建的 G 都被调度执行。

### Q: 为什么设置 GOMAXPROCS 没有效果？

A: 确保在 `gmp.Init()` **之前**设置环境变量：
```go
os.Setenv("GOMAXPROCS", "4")
gmp.Init() // 会读取 GOMAXPROCS
```

## 下一步

- 阅读 `src/gmp/README.md` 了解实现原理
- 查看测试用例了解更多用法
- 尝试扩展 API 添加新功能

---

**项目地址**: `go-rem/gmp`
**文档**: `src/gmp/README.md`
**源码**: `src/gmp/proc_rem.go`
