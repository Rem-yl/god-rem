package gmp

import (
	"sync"
)

// 导出的 API，供外部使用

var (
	initialized bool
	initOnce    sync.Once
)

// Init 初始化 GMP 调度器
// 必须在使用 Go() 之前调用一次
func Init() {
	initOnce.Do(func() {
		schedinit()
		initialized = true
	})
}

// Go 创建一个新的 Goroutine 来执行 fn
// 类似于 go func() { ... }
func Go(fn func()) {
	if !initialized {
		panic("gmp.Init() must be called before gmp.Go()")
	}
	newproc(fn)
}

// Run 启动调度器并运行所有 Goroutine
// 这个函数会阻塞直到所有 G 执行完毕
func Run() {
	if !initialized {
		panic("gmp.Init() must be called before gmp.Run()")
	}
	schedule()
}

// GetGCount 获取当前队列中 G 的数量（用于调试）
func GetGCount() int {
	if !initialized {
		return 0
	}

	total := len(sched.runq) // 全局队列

	// 加上所有 P 的本地队列
	for _, pp := range sched.allp {
		if pp != nil {
			total += int(pp.runqtail - pp.runqhead)
			if pp.runnext != nil {
				total++
			}
		}
	}

	return total
}
