package gmp

import (
	"sync"
	"testing"
)

// 1. goid 验证
// 2. g创建初始状态验证
// 3. g task验证
func TestCreateG(t *testing.T) {
	g := newG(func() {})

	if g.goid <= 0 {
		t.Errorf("want goid > 0, got: %d", g.goid)
	}

	if g.status != _Gidle {
		t.Errorf("want new g status _Gidle, got: %v", g.status)
	}

	if g.fn == nil {
		t.Error("g.fn is nil")
	}
}

// 顺序创建 g
// 要求:
// goid顺序增长
func TestGoidUnique(t *testing.T) {
	num := 100
	goids := make([]uint64, num)

	for i := range num {
		g := newG(func() {})
		goids[i] = g.goid
	}

	// check goid是否顺序增长
	for i := range num - 1 {
		if goids[i]+1 != goids[i+1] {
			t.Error("goid add not linear")
		}
	}
}

// 并发创建 g
// 要求:
// goid 互不相同
func TestGoidConcurrent(t *testing.T) {
	num := 100
	var wg sync.WaitGroup
	goid_chan := make(chan uint64, num)

	for range num {
		wg.Add(1)
		go func() {
			defer wg.Done()
			g := newG(func() {})
			goid_chan <- g.goid
		}()
	}

	wg.Wait()
	close(goid_chan)

	seen := make(map[uint64]bool, num)

	for goid := range goid_chan {
		if _, ok := seen[goid]; ok {
			t.Error("goid not unique")
		}
		seen[goid] = true
	}
}

// 测试 ExecuteG
// 1. g.fn必须被执行
// 2. g.fn执行后, g.status必须是_Gdead
func TestExecuteG(t *testing.T) {
	is_run := false
	task := func() {
		is_run = true
	}

	g := newG(task)

	ExecuteG(g) // remtodo: 没有测试panic的情况

	if is_run != true {
		t.Error("task not run.")
	}

	if g.status != _Gdead {
		t.Error("g not dead")
	}
}

// 1. 初始化g0和m0, 检查他们的 id
// 2. 互相绑定g0和m0
// 3. 设置当前g 为g0
func TestOsinit(t *testing.T) {
	osinit()

	if g0.goid != 0 {
		t.Errorf("want g0.goid = 0, got %d", g0.goid)
	}

	if g0.status != _Gidle {
		t.Errorf("want g0.status is _Gidle, got %v", g0.status)
	}

	if m0.id != 0 {
		t.Errorf("want m0.id = 0, got %d", m0.id)
	}

	if g0.m != m0 {
		t.Error("g0.m is not m0")
	}

	if m0.curg != g0 {
		t.Error("m0.curg is not g0")
	}

	if currentG != g0 { // remtodo: will use getg() to get currentG
		t.Error("currentG is not g0")
	}
}
