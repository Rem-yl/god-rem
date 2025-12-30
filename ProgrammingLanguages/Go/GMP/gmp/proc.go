package gmp

// 模拟runtime前的汇编代码操作
// 1. 初始化g0, m0
// 2. 互相绑定 g0, m0
// 3. 设置当前g为g0
func osinit() {
	g0 = &g{
		goid:   0,
		status: _Gidle,
		fn:     nil, // remtodo: g0.fn初始化为nil?
	}

	m0 = &m{
		id:   0,
		curg: g0,
	}

	g0.m = m0

	currentG = g0 // remtodo: will use setg() to set
}

func newG(task func()) *g {
	goid := goidgen.Add(1)

	return &g{
		goid:   goid,
		status: _Gidle,
		fn:     task,
	}
}

func ExecuteG(g *g) {
	if g.fn == nil {
		panic("g.fn is nil")
	}

	if g.status == _Gdead {
		panic("g is not runnable")
	}

	g.fn()
	g.status = _Gdead
}
