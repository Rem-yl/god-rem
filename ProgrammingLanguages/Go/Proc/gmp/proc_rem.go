package gmp

import (
	"os"
	"runtime"
	"strconv"
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

// initG0M0 初始化 g0 和 m0 的双向绑定
func initG0M0() {
	g0 = &g{
		goid:   0,
		status: _Gidle,
		fn:     nil,
	}

	m0 = &m{
		id:       0,
		g0:       g0,
		curg:     g0,
		spinning: false,
	}

	g0.m = m0
	g0.g0 = g0 // g0 的 g0 指向自己

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

	// 读取 GOMAXPROCS 环境变量
	procs := int32(runtime.NumCPU())
	if v := os.Getenv("GOMAXPROCS"); v != "" {
		if i, err := strconv.ParseInt(v, 10, 32); err == nil {
			procs = int32(i)
		}
	}

	if procresize(procs) != nil {
		panic("unknown runnable goroutine during bootstrap")
	}
}

// ============ Phase 3: 调度器核心逻辑 ============

// procresize 调整 P 的数量
// 返回一个有可运行 G 的 P（如果有的话）
func procresize(nprocs int32) *p {
	old := len(sched.allp)

	// 创建新的 P
	for i := int32(old); i < nprocs; i++ {
		pp := &p{
			id:     int64(i),
			status: _Pidle,
		}
		sched.allp = append(sched.allp, pp)
	}

	// 分配 P 给 m0
	mp := getg().m
	if mp != nil {
		if mp.p != nil && mp.p.id < int64(nprocs) {
			// 继续使用当前的 P
			mp.p.status = _Prunning
		} else {
			// 获取一个 P
			if nprocs > 0 {
				pp := sched.allp[0]
				pp.status = _Prunning
				mp.p = pp
				pp.m = mp
			}
		}
	}

	// 将多余的 P 放入空闲列表
	for i := nprocs; i < int32(len(sched.allp)); i++ {
		pp := sched.allp[i]
		if pp == nil {
			continue
		}
		pp.status = _Pdead
		// 将其本地队列的 G 转移到全局队列
		for !runqempty(pp) {
			gp := runqget(pp)
			if gp != nil {
				globrunqput(gp)
			}
		}
	}

	// 将空闲的 P 放入 pidle 链表
	var pidle *p
	var runningP int32 = 0
	for i := int32(0); i < nprocs; i++ {
		pp := sched.allp[i]
		if mp != nil && mp.p == pp {
			runningP = 1
			continue
		}
		pp.status = _Pidle
		pp.link = pidle
		pidle = pp
	}
	sched.pidle = pidle
	sched.npidle.Store(nprocs - runningP)

	return nil
}

// newproc 创建一个新的 G 来运行 fn
func newproc(fn func()) {
	gp := newG(fn)
	gp.status = _Grunnable

	// 获取当前的 P
	mp := getg().m
	pp := mp.p

	if pp == nil {
		// 没有 P，放入全局队列
		globrunqput(gp)
	} else {
		// 放入 P 的本地队列
		// next=true 使用 runnext 优化
		runqput(pp, gp, true)
	}
}

// findrunnable 查找一个可运行的 G
// 按照以下顺序查找：
// 1. 本地队列
// 2. 全局队列
// 3. 网络轮询器（暂不实现）
// 4. 工作窃取
func findrunnable() *g {
	mp := getg().m
	pp := mp.p

	if pp == nil {
		return nil
	}

	// 1. 从本地队列获取
	if gp := runqget(pp); gp != nil {
		return gp
	}

	// 2. 从全局队列获取
	if gp := globrunqget(pp, 1); gp != nil {
		return gp
	}

	// 3. 尝试从其他 P 窃取
	if gp := runqsteal(pp); gp != nil {
		return gp
	}

	// 没有可运行的 G
	return nil
}

// execute 开始执行 gp
func execute(gp *g) {
	mp := getg().m

	// 设置状态
	gp.status = _Grunning
	gp.m = mp // 设置 g.m 关联
	mp.curg = gp

	// 切换到 gp
	setg(gp)

	// 执行 G 的函数
	if gp.fn != nil {
		gp.fn()
	}

	// G 执行完毕，调用 goexit
	goexit()
}

// goexit G 退出时的清理工作
func goexit() {
	gp := getg()
	mp := gp.m

	// 设置状态为 dead
	gp.status = _Gdead

	// 切换回 g0
	setg(mp.g0)
	mp.curg = nil

	// 继续调度
	schedule()
}

// schedule 调度循环
// 找到一个可运行的 G 并执行它
func schedule() {
	mp := getg().m

	if mp == nil {
		panic("schedule: m is nil")
	}

	// 查找可运行的 G
	gp := findrunnable()

	if gp == nil {
		// 没有可运行的 G
		return
	}

	// 执行找到的 G
	execute(gp)
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

// ============ Phase 5: 工作窃取 ============

// runqsteal 尝试从其他 P 的运行队列窃取 G
// 窃取一半的 G 到 pp 的本地队列
func runqsteal(pp *p) *g {
	// 遍历所有 P
	for _, p2 := range sched.allp {
		if p2 == pp {
			continue // 跳过自己
		}

		// 尝试从 p2 窃取
		if gp := runqstealFromP(pp, p2); gp != nil {
			return gp
		}
	}

	return nil
}

// runqstealFromP 从 p2 窃取一半的 G 到 pp
// 返回第一个窃取到的 G
func runqstealFromP(pp, p2 *p) *g {
	h := p2.runqhead
	t := p2.runqtail
	n := t - h

	if n == 0 {
		return nil // p2 队列为空
	}

	// 窃取一半
	n = n / 2
	if n == 0 {
		n = 1 // 至少窃取一个
	}

	// 获取第一个 G 作为返回值
	var gp *g
	var batch []*g

	for i := uint32(0); i < n; i++ {
		g1 := p2.runq[(h+i)%uint32(len(p2.runq))]
		if g1 == nil {
			continue
		}

		if gp == nil {
			gp = g1 // 第一个 G 作为返回值
		} else {
			batch = append(batch, g1)
		}
	}

	// 更新 p2 的队列头
	p2.runqhead = h + n

	// 将剩余的 G 放入 pp 的本地队列
	for _, g1 := range batch {
		runqput(pp, g1, false)
	}

	return gp
}
