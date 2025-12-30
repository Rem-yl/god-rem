package gmp

import "sync/atomic"

const (
	_Gidle int32 = iota
	_Grunnable
	_Gwaiting
	_Gdead
)

var (
	goidgen  atomic.Uint64
	currentG *g
	g0       *g
	m0       *m
)

// g 代表goroutine
type g struct {
	goid   uint64 // g的数量可能会非常大
	status int32
	fn     func()

	m *m
}

// m 对应线程, 是真正的工作实体
type m struct {
	id   int32 // 负数值的id有特殊作用
	curg *g
}
