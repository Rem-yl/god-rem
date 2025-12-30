package gmp

import (
	"testing"
)

// Phase 2: 测试 P 的本地队列操作

func TestRunqPutGet(t *testing.T) {
	// 创建一个 P
	pp := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	// 测试空队列
	if !runqempty(pp) {
		t.Error("新创建的 P 队列应该为空")
	}

	// 创建一些 G
	g1 := newG(func() {})
	g2 := newG(func() {})
	g3 := newG(func() {})

	// 放入队列
	runqput(pp, g1, false)
	runqput(pp, g2, false)
	runqput(pp, g3, false)

	// 检查队列不为空
	if runqempty(pp) {
		t.Error("队列不应该为空")
	}

	// 检查状态
	if g1.status != _Grunnable {
		t.Error("g1 应该是 Grunnable 状态")
	}

	// 取出 G（应该按 FIFO 顺序）
	got1 := runqget(pp)
	if got1 != g1 {
		t.Errorf("期望获取 g1, 实际获取 goid=%d", got1.goid)
	}

	got2 := runqget(pp)
	if got2 != g2 {
		t.Errorf("期望获取 g2, 实际获取 goid=%d", got2.goid)
	}

	got3 := runqget(pp)
	if got3 != g3 {
		t.Errorf("期望获取 g3, 实际获取 goid=%d", got3.goid)
	}

	// 队列应该为空
	if !runqempty(pp) {
		t.Error("队列应该为空")
	}

	// 再次获取应该返回 nil
	got4 := runqget(pp)
	if got4 != nil {
		t.Error("空队列应该返回 nil")
	}
}

func TestRunqRunnext(t *testing.T) {
	// 创建一个 P
	pp := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	g1 := newG(func() {})
	g2 := newG(func() {})
	g3 := newG(func() {})

	// 使用 next=true 放入 runnext
	runqput(pp, g1, true)

	// 检查 runnext
	if pp.runnext != g1 {
		t.Error("g1 应该在 runnext")
	}

	// 再放入一个到 runnext（应该把 g1 挤到队列）
	runqput(pp, g2, true)

	if pp.runnext != g2 {
		t.Error("g2 应该在 runnext")
	}

	// 普通放入
	runqput(pp, g3, false)

	// 获取顺序：runnext(g2), queue(g1), queue(g3)
	got1 := runqget(pp)
	if got1 != g2 {
		t.Errorf("应该先获取 runnext 的 g2, 实际获取 goid=%d", got1.goid)
	}

	got2 := runqget(pp)
	if got2 != g1 {
		t.Errorf("应该获取 g1, 实际获取 goid=%d", got2.goid)
	}

	got3 := runqget(pp)
	if got3 != g3 {
		t.Errorf("应该获取 g3, 实际获取 goid=%d", got3.goid)
	}
}

func TestRunqFull(t *testing.T) {
	// 创建一个 P
	pp := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	// 填满队列 (256 个)
	var gs [256]*g
	for i := 0; i < 256; i++ {
		gs[i] = newG(func() {})
		runqput(pp, gs[i], false)
	}

	// 检查队列满了
	if pp.runqtail-pp.runqhead != 256 {
		t.Errorf("队列应该有 256 个元素, 实际 %d", pp.runqtail-pp.runqhead)
	}

	// 再放入一个（应该触发 runqputslow，将一半放入全局队列）
	extraG := newG(func() {})
	runqput(pp, extraG, false)

	// 本地队列应该减少
	localCount := pp.runqtail - pp.runqhead
	if localCount >= 256 {
		t.Errorf("队列应该有一半被转移到全局队列, 本地剩余 %d", localCount)
	}

	// 全局队列应该有数据
	globalCount := len(sched.runq)
	if globalCount == 0 {
		t.Error("全局队列应该有数据")
	}

	t.Logf("本地队列: %d, 全局队列: %d", localCount, globalCount)
}

func TestGlobalQueue(t *testing.T) {
	// 重置全局队列
	sched.runq = nil

	g1 := newG(func() {})
	g2 := newG(func() {})
	g3 := newG(func() {})

	// 放入全局队列
	globrunqput(g1)
	globrunqput(g2)
	globrunqput(g3)

	if len(sched.runq) != 3 {
		t.Errorf("全局队列应该有 3 个 G, 实际 %d", len(sched.runq))
	}

	// 创建一个 P 来获取
	pp := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	// 从全局队列获取
	got := globrunqget(pp, 10)
	if got != g1 {
		t.Errorf("应该获取 g1, 实际 goid=%d", got.goid)
	}

	// 检查是否有其他 G 被转移到本地队列
	t.Logf("本地队列数量: %d", pp.runqtail-pp.runqhead)
}

func TestRunqempty(t *testing.T) {
	pp := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
		runnext:  nil,
	}

	// 空队列
	if !runqempty(pp) {
		t.Error("应该为空")
	}

	// 添加到 runnext
	g1 := newG(func() {})
	pp.runnext = g1

	if runqempty(pp) {
		t.Error("有 runnext 不应该为空")
	}

	// 清空 runnext，添加到队列
	pp.runnext = nil
	runqput(pp, g1, false)

	if runqempty(pp) {
		t.Error("有队列元素不应该为空")
	}
}
