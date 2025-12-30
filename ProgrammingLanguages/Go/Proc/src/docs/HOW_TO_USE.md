# 如何使用 GMP 调度器

## 🎯 快速回答

**Q: 如何在外部包使用我的 gmp？**

**A**: 非常简单！

```go
package main

import "go-rem/gmp"

func main() {
    gmp.Init()                  // 1. 初始化

    gmp.Go(func() {             // 2. 创建 Goroutine
        // 你的代码
    })

    gmp.Run()                   // 3. 运行调度器
}
```

## 📁 项目结构（已去掉 replace）

```
Proc/
├── go.mod                  # 主模块 (无需 replace)
│
├── gmp/                    # GMP 包
│   ├── types.go           # 数据结构
│   ├── proc_rem.go        # 调度器
│   ├── api.go             # 对外 API ← 使用这个！
│   └── *_test.go          # 测试文件
│
└── examples/               # 示例程序
    ├── basic/
    ├── producer-consumer/
    └── work-stealing/
```

## 🚀 使用方式

### 方式 1: 直接运行示例

```bash
# 进入项目目录
cd /path/to/Proc

# 运行示例
go run examples/basic/main.go
go run examples/work-stealing/main.go
go run examples/producer-consumer/main.go
```

### 方式 2: 创建你自己的程序

#### 步骤 1: 在项目下创建新文件

```bash
cd /path/to/Proc
mkdir my-app
cd my-app
```

#### 步骤 2: 创建 main.go

```go
package main

import (
	"fmt"
	"go-rem/gmp"  // ← 直接导入，无需 replace！
)

func main() {
	// 初始化 GMP 调度器
	gmp.Init()

	// 创建多个 Goroutine
	for i := 0; i < 5; i++ {
		i := i  // 捕获循环变量
		gmp.Go(func() {
			fmt.Printf("Task %d running\n", i)
		})
	}

	// 启动调度器
	gmp.Run()

	fmt.Println("All done!")
}
```

#### 步骤 3: 运行程序

```bash
go run main.go
```

输出：
```
Task 0 running
Task 1 running
Task 2 running
Task 3 running
Task 4 running
All done!
```

## 🎨 常见使用模式

### 模式 1: 并发计算

```go
package main

import (
	"fmt"
	"go-rem/gmp"
)

func main() {
	gmp.Init()

	// 计算平方
	numbers := []int{1, 2, 3, 4, 5}

	for _, num := range numbers {
		num := num
		gmp.Go(func() {
			result := num * num
			fmt.Printf("%d^2 = %d\n", num, result)
		})
	}

	gmp.Run()
}
```

### 模式 2: 数据处理流水线

```go
package main

import (
	"fmt"
	"go-rem/gmp"
)

func main() {
	gmp.Init()

	data := []string{"apple", "banana", "cherry"}

	// Stage 1: 处理数据
	for _, item := range data {
		item := item
		gmp.Go(func() {
			fmt.Printf("Processing: %s\n", item)

			// Stage 2: 嵌套处理
			gmp.Go(func() {
				fmt.Printf("Finished: %s\n", item)
			})
		})
	}

	gmp.Run()
}
```

### 模式 3: 工作窃取演示

```go
package main

import (
	"fmt"
	"os"
	"go-rem/gmp"
)

func main() {
	// 设置多个 P
	os.Setenv("GOMAXPROCS", "4")

	gmp.Init()

	// 创建大量任务
	for i := 0; i < 100; i++ {
		i := i
		gmp.Go(func() {
			fmt.Printf("Task %d on P\n", i)
		})
	}

	fmt.Printf("Created %d goroutines\n", gmp.GetGCount())
	gmp.Run()
}
```

## 🔧 API 参考

### gmp.Init()

**作用**: 初始化 GMP 调度器

**必须**: 是，在使用 Go() 之前必须调用

**示例**:
```go
gmp.Init()
```

### gmp.Go(fn func())

**作用**: 创建一个新的 Goroutine 来执行 fn

**类比**: 标准 Go 的 `go func() { ... }()`

**示例**:
```go
gmp.Go(func() {
    fmt.Println("Hello from goroutine!")
})
```

### gmp.Run()

**作用**: 启动调度器，运行所有 Goroutine

**特点**: 会阻塞直到所有 G 执行完毕

**示例**:
```go
gmp.Run()
```

### gmp.GetGCount() int

**作用**: 获取当前队列中 G 的数量（用于调试）

**示例**:
```go
count := gmp.GetGCount()
fmt.Printf("Queue has %d goroutines\n", count)
```

## ⚠️ 注意事项

### 1. 必须先调用 Init()

```go
// ❌ 错误
gmp.Go(func() { ... })  // panic: must call Init() first

// ✅ 正确
gmp.Init()
gmp.Go(func() { ... })
```

### 2. Run() 会阻塞

```go
gmp.Init()
gmp.Go(func() { fmt.Println("Task 1") })
gmp.Run()               // ← 阻塞直到所有 G 完成
fmt.Println("Done!")    // ← 这行会等 G 完成后才执行
```

### 3. 闭包变量捕获

```go
// ❌ 错误 - 循环变量问题
for i := 0; i < 10; i++ {
	gmp.Go(func() {
		fmt.Println(i)  // 可能全部打印 10
	})
}

// ✅ 正确 - 捕获变量
for i := 0; i < 10; i++ {
	i := i  // 创建新变量
	gmp.Go(func() {
		fmt.Println(i)  // 正确打印 0-9
	})
}
```

### 4. 不支持 channel 和 select

```go
// ❌ 不支持
ch := make(chan int)
gmp.Go(func() {
    ch <- 42  // 不会工作
})

// ✅ 使用共享变量（注意竞态）
var result int
gmp.Go(func() {
    result = 42  // 简化版可以这样做
})
```

## 🌍 环境变量

### GOMAXPROCS

设置 P（处理器）的数量

```bash
# 在命令行设置
GOMAXPROCS=4 go run main.go

# 或在代码中设置
import "os"
os.Setenv("GOMAXPROCS", "4")
gmp.Init()  // 会读取 GOMAXPROCS
```

## 📊 调试技巧

### 查看队列状态

```go
gmp.Init()

// 创建一些 G
for i := 0; i < 10; i++ {
	gmp.Go(func() { /* ... */ })
}

// 查看队列
fmt.Printf("Queue: %d goroutines\n", gmp.GetGCount())

gmp.Run()
```

### 观察执行顺序

```go
gmp.Init()

for i := 0; i < 5; i++ {
	i := i
	gmp.Go(func() {
		fmt.Printf("Start %d\n", i)
		// 模拟工作
		for j := 0; j < 1000; j++ {
			// ...
		}
		fmt.Printf("End %d\n", i)
	})
}

gmp.Run()
```

## 🎓 与标准 Go 的对比

| 标准 Go | GMP 调度器 | 说明 |
|---------|-----------|------|
| `go func() {...}` | `gmp.Go(func() {...})` | 创建协程 |
| 自动调度 | `gmp.Run()` | 需要显式调用 |
| 自动初始化 | `gmp.Init()` | 需要显式初始化 |
| `runtime.GOMAXPROCS()` | `os.Setenv("GOMAXPROCS", "4")` | 设置 P 数量 |

## 💡 完整示例

```go
package main

import (
	"fmt"
	"os"
	"go-rem/gmp"
)

func main() {
	// 1. 配置环境（可选）
	os.Setenv("GOMAXPROCS", "2")

	// 2. 初始化调度器
	gmp.Init()

	// 3. 创建任务
	tasks := []string{"Task A", "Task B", "Task C"}

	for _, task := range tasks {
		task := task  // 捕获变量
		gmp.Go(func() {
			fmt.Printf("Processing: %s\n", task)
		})
	}

	// 4. 查看状态（可选）
	fmt.Printf("Created %d goroutines\n", gmp.GetGCount())

	// 5. 运行调度器
	fmt.Println("Starting scheduler...")
	gmp.Run()

	// 6. 完成
	fmt.Println("All tasks completed!")
}
```

运行：
```bash
go run main.go
```

输出：
```
Created 3 goroutines
Starting scheduler...
Processing: Task B
Processing: Task C
Processing: Task A
All tasks completed!
```

## 📚 更多资源

- **快速开始**: [QUICKSTART.md](QUICKSTART.md)
- **示例程序**: [examples/README.md](examples/README.md)
- **实现原理**: [gmp/README.md](gmp/README.md)
- **完整文档**: [PROJECT_README.md](PROJECT_README.md)

## ❓ 常见问题

### Q: 为什么我的程序卡住了？

A: 确保调用了 `gmp.Run()`，它会阻塞直到所有 G 完成。

### Q: 可以在 Goroutine 内创建新的 Goroutine 吗？

A: 可以！

```go
gmp.Init()
gmp.Go(func() {
    fmt.Println("Parent")
    gmp.Go(func() {
        fmt.Println("Child")
    })
})
gmp.Run()
```

### Q: 如何查看有多少个 Goroutine 在运行？

A: 使用 `gmp.GetGCount()`

```go
count := gmp.GetGCount()
fmt.Printf("Active goroutines: %d\n", count)
```

---

**现在你已经知道如何使用 GMP 调度器了！** 🎉

开始你的第一个程序: `go run examples/basic/main.go`
