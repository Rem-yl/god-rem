package gmp

import (
	"os"
	"strconv"
	"time"
)

func getg() *g {
	return currentG
}

func setg(g *g) {
	currentG = g
}

func ExecuteG(g *g) {
	g.fn()
	g.status = _Gdead
}

// remtodo: g0, m0的初始化和绑定
func initG0M0() {
	g0 = &g{
		goid:   0,
		status: _Gidle,
		fn:     nil,
	}

	m0 = &m{
		id:       0,
		curg:     g0,
		spinning: false,
	}

	g0.m = m0

	setg(g0)
}

// 模拟汇编代码的初始化操作
func osinit() {
	initG0M0()
}

func schedinit() {
	osinit()

	gp := getg()
	if gp != g0 {
		panic("schedinit must run on g0")
	}

	sched.maxmcount = 10000

	var procs int32
	v := os.Getenv("GOMAXPROCS")

	if i, err := strconv.ParseInt(v, 10, 32); err == nil {
		procs = int32(i)
	} else {
		procs = 1
	}

	if procresize(procs) != nil {
		panic("unknown runnable goroutine during bootstrap")
	}
}

func procresize(procs int32) *p {
	// TODO: 实现 P 的创建和初始化
	_ = time.Now() // 占位，后续实现时会用到
	return nil
}

// ============ Phase 2: P 的本地队列操作 ============

// runqput 将 gp 放入 pp 的本地可运行队列
// 如果队列满了，将一半的 G 放入全局队列
// 参数 next 为 true 时，将 gp 放入 pp.runnext
func runqput(pp *p, gp *g, next bool) {
	if next {
		// 优先放入 runnext
		oldnext := pp.runnext
		pp.runnext = gp
		gp.status = _Grunnable

		if oldnext == nil {
			return
		}
		// runnext 被占用，将旧的 G 放入队列
		gp = oldnext
	}

	// 尝试放入本地队列
retry:
	h := pp.runqhead
	t := pp.runqtail

	// 队列未满
	if t-h < uint32(len(pp.runq)) {
		pp.runq[t%uint32(len(pp.runq))] = gp
		pp.runqtail = t + 1
		gp.status = _Grunnable
		return
	}

	// 队列满了，将一半放入全局队列
	if runqputslow(pp, gp) {
		return
	}
	// 全局队列操作失败，重试
	goto retry
}

// runqputslow 将 pp 的本地队列的一半 G 和 gp 一起放入全局队列
func runqputslow(pp *p, gp *g) bool {
	var batch [len(pp.runq)/2 + 1]*g

	// 获取本地队列的一半
	h := pp.runqhead
	t := pp.runqtail
	n := t - h
	n = n / 2

	if n != uint32(len(pp.runq)/2) {
		panic("runqputslow: queue size mismatch")
	}

	for i := uint32(0); i < n; i++ {
		batch[i] = pp.runq[(h+i)%uint32(len(pp.runq))]
	}
	batch[n] = gp

	// 更新队列头
	pp.runqhead = h + n

	// 放入全局队列
	globrunqputbatch(batch[:n+1])
	return true
}

// runqget 从 pp 的本地可运行队列获取一个 G
// 如果 inheritTime 为 true，gp 应该继承当前时间片
func runqget(pp *p) *g {
	// 先检查 runnext
	next := pp.runnext
	if next != nil {
		pp.runnext = nil
		return next
	}

	// 从本地队列获取
	h := pp.runqhead
	t := pp.runqtail

	if t == h {
		return nil // 队列为空
	}

	gp := pp.runq[h%uint32(len(pp.runq))]
	pp.runqhead = h + 1
	return gp
}

// runqempty 检查 pp 的本地队列是否为空
func runqempty(pp *p) bool {
	return pp.runnext == nil && pp.runqhead == pp.runqtail
}

// ============ 全局队列操作（简化版）============

// globrunqputbatch 将一批 G 放入全局队列
func globrunqputbatch(batch []*g) {
	for _, gp := range batch {
		if gp != nil {
			gp.status = _Grunnable
			sched.runq = append(sched.runq, gp)
		}
	}
}

// globrunqput 将 gp 放入全局队列
func globrunqput(gp *g) {
	gp.status = _Grunnable
	sched.runq = append(sched.runq, gp)
}

// globrunqget 从全局队列获取一个 G
// 尝试从全局队列获取一批 G
func globrunqget(pp *p, max int32) *g {
	if len(sched.runq) == 0 {
		return nil
	}

	// 获取一个 G
	gp := sched.runq[0]
	sched.runq = sched.runq[1:]

	// 尝试获取更多 G 到本地队列（负载均衡）
	n := int32(len(sched.runq))
	if n > max {
		n = max
	}

	for i := int32(0); i < n && len(sched.runq) > 0; i++ {
		g1 := sched.runq[0]
		sched.runq = sched.runq[1:]
		runqput(pp, g1, false)
	}

	return gp
}
