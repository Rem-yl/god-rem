package gmp

import (
	"sync/atomic"
)

const (
	_Gidle uint32 = iota
	_Grunnable
	_Grunning
	_Gwaiting
	_Gdead
)

var (
	g0       *g
	m0       *m
	currentG *g
	sched    Schedt
)

type g struct {
	goid   uint64
	status uint32
	fn     func()

	m  *m
	g0 *g
}

func newG(task func()) *g {
	goid := sched.goidgen.Add(1)
	return &g{
		goid:   goid,
		status: _Gidle,
		fn:     task,
	}
}

type m struct {
	id       int64
	p        *p
	curg     *g
	g0       *g
	spinning bool
	link     *m // 用于空闲 M 链表
}

// P 的状态
const (
	_Pidle uint32 = iota
	_Prunning
	_Psyscall
	_Pgcstop
	_Pdead
)

type p struct {
	id       int64
	status   uint32
	runqhead uint32
	runqtail uint32
	runq     [256]*g // 每个p自己的运行队列
	runnext  *g
	m        *m
	link     *p // 用于空闲 P 链表
}

type Schedt struct {
	goidgen   atomic.Uint64
	mnext     int64
	maxmcount int32
	runq      []*g //全局运行队列
	pidle     *p   // 空闲的 P 链表
	midle     *m   // 空间的 M 链表
	npidle    atomic.Int32
	allp      []*p
	allm      []*m
}
