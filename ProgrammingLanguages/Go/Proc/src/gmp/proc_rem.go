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
