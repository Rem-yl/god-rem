package gmp

import (
	"os"
	"sync"
	"testing"
)

func TestAPI_BasicUsage(t *testing.T) {
	// 重置状态
	initialized = false
	initOnce = sync.Once{}
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	Init()

	if !initialized {
		t.Error("Init() 应该设置 initialized = true")
	}

	// 创建一些 G
	executed := false
	Go(func() {
		executed = true
	})

	// 验证 G 被创建
	count := GetGCount()
	if count == 0 {
		t.Error("Go() 应该创建 G")
	}

	// 运行调度器
	Run()

	// 验证执行了
	if !executed {
		t.Error("G 应该被执行")
	}
}

func TestAPI_MultipleGoroutines(t *testing.T) {
	// 重置状态
	initialized = false
	initOnce = sync.Once{}
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	Init()

	counter := 0

	// 创建多个 G
	for i := 0; i < 5; i++ {
		Go(func() {
			counter++
		})
	}

	Run()

	if counter != 5 {
		t.Errorf("期望执行 5 个 G, 实际执行 %d", counter)
	}
}

func TestAPI_WithGOMAXPROCS(t *testing.T) {
	// 重置状态
	initialized = false
	initOnce = sync.Once{}
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 设置 GOMAXPROCS
	os.Setenv("GOMAXPROCS", "2")
	defer os.Unsetenv("GOMAXPROCS")

	Init()

	// 验证创建了 2 个 P
	if len(sched.allp) != 2 {
		t.Errorf("期望 2 个 P, 实际 %d", len(sched.allp))
	}

	// 创建一些 G
	sum := 0
	for i := 1; i <= 10; i++ {
		i := i
		Go(func() {
			sum += i
		})
	}

	Run()

	expected := 55 // 1+2+...+10
	if sum != expected {
		t.Errorf("期望 sum = %d, 实际 %d", expected, sum)
	}
}

func TestAPI_PanicBeforeInit(t *testing.T) {
	// 重置状态
	initialized = false
	initOnce = sync.Once{}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Go() 在 Init() 之前应该 panic")
		}
	}()

	Go(func() {})
}

func TestAPI_GetGCount(t *testing.T) {
	// 重置状态
	initialized = false
	initOnce = sync.Once{}
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 未初始化时应该返回 0
	if GetGCount() != 0 {
		t.Error("未初始化时 GetGCount() 应该返回 0")
	}

	Init()

	// 初始时应该是 0
	if GetGCount() != 0 {
		t.Error("初始时 GetGCount() 应该返回 0")
	}

	// 创建 G
	for i := 0; i < 3; i++ {
		Go(func() {})
	}

	count := GetGCount()
	if count != 3 {
		t.Errorf("创建 3 个 G 后, GetGCount() 应该返回 3, 实际 %d", count)
	}
}

func TestAPI_NestedGoroutines(t *testing.T) {
	// 重置状态
	initialized = false
	initOnce = sync.Once{}
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	Init()

	parentExecuted := false
	childExecuted := false

	Go(func() {
		parentExecuted = true

		// 在 G 内创建新的 G
		Go(func() {
			childExecuted = true
		})
	})

	Run()

	if !parentExecuted {
		t.Error("父 G 应该被执行")
	}

	if !childExecuted {
		t.Error("子 G 应该被执行")
	}
}
