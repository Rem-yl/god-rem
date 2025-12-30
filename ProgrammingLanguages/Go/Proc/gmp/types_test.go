package gmp

import (
	"fmt"
	"testing"
)

func TestCreateG(t *testing.T) {
	task := func() {
		fmt.Println("test create g")
	}

	g := newG(task)

	if g.goid <= 0 {
		t.Errorf("want g.goid > 0, got: %d", g.goid)
	}

	if g.status != _Gidle {
		t.Errorf("want init g.status is _Gidle, got: %v", g.status)
	}

	if g.fn == nil {
		t.Error("g.fn is nil")
	}
}

func TestGoidUnique(t *testing.T) {
	num := 100
	var goids []uint64

	task := func() {
		fmt.Println("TestGoidUnique")
	}

	for range num {
		g := newG(task)
		goids = append(goids, g.goid)
	}

	for i := range num - 1 {
		if goids[i]+1 != goids[i+1] {
			t.Error("goid add not linear")
		}
	}
}

func TestGoidConcurrent(t *testing.T) {
	num := 100
	task := func() {
		fmt.Println("TestGoidConcurrent")
	}
	goids := make(chan uint64, num)
	seen := make(map[uint64]bool, num)

	for range num {
		go func() {
			g := newG(task)
			goids <- g.goid
		}()
	}

	var collected []uint64
	for i := 0; i < num; i++ {
		collected = append(collected, <-goids)
	}

	for _, goid := range collected {
		if _, ok := seen[goid]; ok {
			t.Error("goid not unique")
		}
		seen[goid] = true
	}
}

func TestExecuteG(t *testing.T) {
	is_run := false
	task := func() {
		is_run = true
	}

	g := newG(task)
	g.status = _Grunnable // 设置为可运行状态

	ExecuteG(g)

	if is_run != true {
		t.Error("task not run.")
	}

	if g.status != _Gdead {
		t.Error("g not dead")
	}
}

// 测试 ExecuteG 传入 nil
func TestExecuteG_NilG(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ExecuteG(nil) should panic")
		}
	}()

	ExecuteG(nil) // 应该 panic
}

// 测试 ExecuteG 传入 nil 函数
func TestExecuteG_NilFn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ExecuteG with nil fn should panic")
		}
	}()

	g := &g{
		goid:   1,
		status: _Grunnable,
		fn:     nil, // fn 为 nil
	}

	ExecuteG(g) // 应该 panic
}

// 测试 ExecuteG 传入错误状态的 G
func TestExecuteG_BadStatus(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ExecuteG with wrong status should panic")
		}
	}()

	g := newG(func() {})
	g.status = _Grunning // 设置为已运行状态

	ExecuteG(g) // 应该 panic
}

// 测试 ExecuteG 传入已死亡的 G
func TestExecuteG_DeadG(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("ExecuteG with dead G should panic")
		}
	}()

	g := newG(func() {})
	g.status = _Gdead // 设置为已死亡状态

	ExecuteG(g) // 应该 panic
}

// Phase 1: 测试 getg() 和 setg()
func TestGetgSetg(t *testing.T) {
	// 先初始化 g0 和 m0
	initG0M0()

	// 初始应该返回 g0
	gp := getg()
	if gp != g0 {
		t.Errorf("getg() should return g0, got goid=%d", gp.goid)
	}

	// 创建新的 g 并切换
	task := func() {}
	newg := newG(task)
	setg(newg)

	// 验证 getg() 返回新的 g
	if getg() != newg {
		t.Error("getg() should return the new g")
	}

	// 恢复 g0
	setg(g0)
	if getg() != g0 {
		t.Error("getg() should return g0 after restore")
	}
}

// Phase 1: 测试 g0 和 m0 的初始化
func TestInitG0M0(t *testing.T) {
	initG0M0()

	// 验证 g0
	if g0 == nil {
		t.Fatal("g0 should not be nil")
	}
	if g0.goid != 0 {
		t.Errorf("g0.goid should be 0, got %d", g0.goid)
	}
	if g0.status != _Gidle {
		t.Errorf("g0.status should be _Gidle")
	}

	// 验证 m0
	if m0 == nil {
		t.Fatal("m0 should not be nil")
	}
	if m0.id != 0 {
		t.Errorf("m0.id should be 0, got %d", m0.id)
	}

	// 验证 g0 <-> m0 的双向关联
	if g0.m != m0 {
		t.Error("g0.m should point to m0")
	}
	if m0.curg != g0 {
		t.Error("m0.curg should point to g0")
	}

	// 验证 currentG 被设置为 g0
	if getg() != g0 {
		t.Error("currentG should be g0 after init")
	}
}

// Phase 1: 测试 schedinit
func TestSchedinit(t *testing.T) {
	// 重置全局变量
	g0 = nil
	m0 = nil
	currentG = nil

	// 调用 schedinit
	schedinit()

	// 验证初始化完成
	if g0 == nil {
		t.Fatal("g0 should be initialized")
	}
	if m0 == nil {
		t.Fatal("m0 should be initialized")
	}
	if getg() != g0 {
		t.Error("should be running on g0")
	}
	if sched.maxmcount != 10000 {
		t.Errorf("sched.maxmcount should be 10000, got %d", sched.maxmcount)
	}
}
