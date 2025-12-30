package gmp

import (
	"testing"
)

// Phase 5: 工作窃取测试

func TestRunqsteal(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil

	// 初始化
	initG0M0()
	procresize(2) // 创建 2 个 P

	// P[0] 是当前 P
	pp := sched.allp[0]
	// P[1] 是被窃取的 P
	p2 := sched.allp[1]

	// 在 P[1] 中添加一些 G
	for i := 0; i < 10; i++ {
		gp := newG(func() {})
		runqput(p2, gp, false)
	}

	// 验证 P[1] 有 10 个 G
	count1 := p2.runqtail - p2.runqhead
	if count1 != 10 {
		t.Errorf("P[1] 应该有 10 个 G, 实际 %d", count1)
	}

	// P[0] 从 P[1] 窃取
	gp := runqsteal(pp)

	if gp == nil {
		t.Fatal("应该窃取到 G")
	}

	// 验证窃取后的状态
	count2 := p2.runqtail - p2.runqhead
	if count2 >= count1 {
		t.Errorf("P[1] 的 G 应该减少, 之前 %d, 现在 %d", count1, count2)
	}

	t.Logf("窃取前 P[1]: %d 个 G, 窃取后: %d 个 G", count1, count2)

	// 验证 P[0] 获得了一些 G
	count0 := pp.runqtail - pp.runqhead
	t.Logf("P[0] 获得了 %d 个 G", count0)
}

func TestRunqstealFromP(t *testing.T) {
	pp := &p{
		id:       0,
		runqhead: 0,
		runqtail: 0,
	}

	p2 := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	// 在 p2 中添加 6 个 G
	for i := 0; i < 6; i++ {
		gp := newG(func() {})
		runqput(p2, gp, false)
	}

	before := p2.runqtail - p2.runqhead
	if before != 6 {
		t.Errorf("p2 应该有 6 个 G, 实际 %d", before)
	}

	// 从 p2 窃取
	gp := runqstealFromP(pp, p2)

	if gp == nil {
		t.Fatal("应该窃取到 G")
	}

	// 验证 p2 减少了一半（3个）
	after := p2.runqtail - p2.runqhead
	expected := before - before/2
	if after != expected {
		t.Errorf("p2 应该剩余 %d 个 G, 实际 %d", expected, after)
	}

	// 验证 pp 获得了窃取的 G（减去返回的那个）
	count := pp.runqtail - pp.runqhead
	t.Logf("窃取了 %d 个 G, 返回 1 个, pp 本地队列获得 %d 个", before/2, count)
}

func TestRunqstealEmpty(t *testing.T) {
	pp := &p{
		id:       0,
		runqhead: 0,
		runqtail: 0,
	}

	p2 := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	// p2 队列为空
	gp := runqstealFromP(pp, p2)

	if gp != nil {
		t.Error("空队列不应该窃取到 G")
	}
}

func TestRunqstealOneG(t *testing.T) {
	pp := &p{
		id:       0,
		runqhead: 0,
		runqtail: 0,
	}

	p2 := &p{
		id:       1,
		runqhead: 0,
		runqtail: 0,
	}

	// p2 只有 1 个 G
	gp1 := newG(func() {})
	runqput(p2, gp1, false)

	// 窃取
	gp := runqstealFromP(pp, p2)

	if gp == nil {
		t.Fatal("应该窃取到 G")
	}

	// p2 应该被窃取了 1 个（至少窃取1个）
	count := p2.runqtail - p2.runqhead
	if count != 0 {
		t.Errorf("p2 应该为空, 实际还有 %d 个 G", count)
	}
}

func TestFindrunableWithSteal(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	initG0M0()
	procresize(3) // 创建 3 个 P

	// P[0] 是当前 P（空）
	// P[1] 和 P[2] 有 G

	// 在 P[1] 中添加 G
	for i := 0; i < 5; i++ {
		gp := newG(func() {})
		runqput(sched.allp[1], gp, false)
	}

	// 在 P[2] 中添加 G
	for i := 0; i < 8; i++ {
		gp := newG(func() {})
		runqput(sched.allp[2], gp, false)
	}

	// P[0] 查找可运行的 G（应该通过工作窃取找到）
	gp := findrunnable()

	if gp == nil {
		t.Fatal("应该通过工作窃取找到 G")
	}

	t.Logf("成功通过工作窃取找到 G (goid=%d)", gp.goid)

	// 验证某个 P 的 G 减少了
	count1 := sched.allp[1].runqtail - sched.allp[1].runqhead
	count2 := sched.allp[2].runqtail - sched.allp[2].runqhead

	t.Logf("P[1]: %d 个 G, P[2]: %d 个 G", count1, count2)

	if count1 == 5 && count2 == 8 {
		t.Error("应该有 P 的 G 被窃取")
	}
}

func TestWorkStealingBalance(t *testing.T) {
	// 重置
	g0 = nil
	m0 = nil
	sched.allp = nil
	sched.runq = nil

	// 初始化
	initG0M0()
	procresize(2) // 创建 2 个 P

	// 在 P[1] 中添加大量 G
	for i := 0; i < 20; i++ {
		gp := newG(func() {})
		runqput(sched.allp[1], gp, false)
	}

	before := sched.allp[1].runqtail - sched.allp[1].runqhead
	t.Logf("窃取前 P[1]: %d 个 G", before)

	// P[0] 窃取多次
	var stolen []*g
	for i := 0; i < 3; i++ {
		gp := runqsteal(sched.allp[0])
		if gp != nil {
			stolen = append(stolen, gp)
		}
	}

	after := sched.allp[1].runqtail - sched.allp[1].runqhead
	t.Logf("窃取后 P[1]: %d 个 G", after)

	if len(stolen) == 0 {
		t.Error("应该窃取到 G")
	}

	t.Logf("成功窃取 %d 轮, 获得 %d 个返回的 G", 3, len(stolen))

	// P[0] 应该也有一些 G
	count0 := sched.allp[0].runqtail - sched.allp[0].runqhead
	t.Logf("P[0] 本地队列: %d 个 G", count0)
}
