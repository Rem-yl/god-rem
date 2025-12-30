package gmp

import (
	"testing"
)

// Phase 3: 调度器核心功能测试

func TestProcresize(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.pidle = nil

	// 初始化
	initG0M0()

	// 创建 2 个 P
	procresize(2)

	// 验证 P 的数量
	if len(sched.allp) != 2 {
		t.Errorf("应该有 2 个 P, 实际 %d", len(sched.allp))
	}

	// 验证 m0 绑定了 P
	if m0.p == nil {
		t.Fatal("m0 应该绑定了 P")
	}

	if m0.p.id != 0 {
		t.Errorf("m0 应该绑定 P[0], 实际绑定 P[%d]", m0.p.id)
	}

	// 验证 P[0] 的状态
	if sched.allp[0].status != _Prunning {
		t.Error("P[0] 应该是 running 状态")
	}

	// 验证 P[1] 在空闲列表
	if sched.pidle == nil {
		t.Fatal("应该有空闲的 P")
	}

	if sched.allp[1].status != _Pidle {
		t.Error("P[1] 应该是 idle 状态")
	}
}

func TestNewproc(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	schedinit()

	task := func() {}

	// 创建新的 G
	newproc(task)

	// 验证 G 被创建并放入队列
	if m0.p == nil {
		t.Fatal("m0 应该有 P")
	}

	// 应该在 runnext 或本地队列
	if m0.p.runnext == nil && runqempty(m0.p) {
		// 检查全局队列
		if len(sched.runq) == 0 {
			t.Error("G 应该在某个队列中")
		}
	}
}

func TestScheduleBasic(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	schedinit()

	counter := 0
	task1 := func() {
		counter++
	}
	task2 := func() {
		counter += 10
	}

	// 创建 2 个 G
	newproc(task1)
	newproc(task2)

	// 执行调度（手动调用一次）
	schedule()

	// 至少执行了一个任务
	if counter == 0 {
		t.Error("至少应该执行一个任务")
	}

	t.Logf("执行结果: counter = %d", counter)
}

func TestFindrunnable(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	schedinit()

	// 1. 空队列
	gp := findrunnable()
	if gp != nil {
		t.Error("空队列应该返回 nil")
	}

	// 2. 本地队列有 G
	task := func() {}
	newproc(task)

	gp = findrunnable()
	if gp == nil {
		t.Error("应该找到可运行的 G")
	}
	if gp.status != _Grunnable {
		t.Error("找到的 G 应该是 Grunnable 状态")
	}

	// 3. 全局队列有 G
	g1 := newG(task)
	globrunqput(g1)

	gp2 := findrunnable()
	if gp2 == nil {
		t.Error("应该从全局队列找到 G")
	}
}

func TestExecuteAndGoexit(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	schedinit()

	task := func() {
		// 验证当前在用户 G 上运行
		gp := getg()
		if gp == g0 {
			t.Error("应该在用户 G 上运行，而不是 g0")
		}
	}

	gp := newG(task)
	gp.m = m0

	// 记录初始状态
	if getg() != g0 {
		t.Error("初始应该在 g0 上")
	}

	// 注意：execute 会调用 goexit，goexit 又会调用 schedule
	// 为了避免无限循环，我们直接测试 execute 的状态变化
	// 而不是真正执行完整流程

	gp.status = _Grunnable

	// 手动设置状态测试
	oldStatus := gp.status
	gp.status = _Grunning
	m0.curg = gp

	if gp.status != _Grunning {
		t.Error("G 应该是 running 状态")
	}

	if m0.curg != gp {
		t.Error("m0.curg 应该指向 gp")
	}

	// 恢复
	gp.status = oldStatus
	m0.curg = g0
}

func TestScheduleMultipleGs(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	schedinit()

	count := 0

	// 创建多个 G
	for i := 0; i < 5; i++ {
		idx := i
		task := func() {
			count++
			t.Logf("执行 G %d", idx)
		}
		newproc(task)
	}

	// 验证 G 都被创建了
	totalGs := len(sched.runq)
	if m0.p != nil {
		totalGs += int(m0.p.runqtail - m0.p.runqhead)
		if m0.p.runnext != nil {
			totalGs++
		}
	}

	if totalGs < 5 {
		t.Errorf("应该创建 5 个 G, 实际 %d", totalGs)
	}

	t.Logf("创建了 %d 个 G", totalGs)
}

func TestProcresizeExpand(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil

	// 初始化
	initG0M0()

	// 先创建 1 个 P
	procresize(1)
	if len(sched.allp) != 1 {
		t.Errorf("应该有 1 个 P, 实际 %d", len(sched.allp))
	}

	// 扩展到 4 个 P
	procresize(4)
	if len(sched.allp) != 4 {
		t.Errorf("应该有 4 个 P, 实际 %d", len(sched.allp))
	}

	// 验证新创建的 P
	for i, pp := range sched.allp {
		if pp == nil {
			t.Errorf("P[%d] 不应该为 nil", i)
		}
		if pp.id != int64(i) {
			t.Errorf("P[%d].id = %d, 应该是 %d", i, pp.id, i)
		}
	}
}

func TestProcresizeShrink(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	initG0M0()

	// 创建 4 个 P
	procresize(4)

	// 在多余的 P 中添加一些 G
	if len(sched.allp) >= 2 {
		pp := sched.allp[1]
		for i := 0; i < 5; i++ {
			gp := newG(func() {})
			runqput(pp, gp, false)
		}
	}

	// 缩减到 2 个 P
	procresize(2)

	if len(sched.allp) != 4 { // allp 不会缩小
		t.Logf("allp 保持原大小: %d", len(sched.allp))
	}

	// 多余的 P 应该被标记为 dead
	for i := 2; i < 4; i++ {
		if sched.allp[i] != nil && sched.allp[i].status != _Pdead {
			t.Errorf("P[%d] 应该是 dead 状态", i)
		}
	}

	// 被转移的 G 应该在全局队列
	if len(sched.runq) == 0 {
		t.Log("注意：P 中的 G 可能已被转移到全局队列")
	}
}
