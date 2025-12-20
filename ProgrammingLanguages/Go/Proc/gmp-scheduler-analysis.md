# Go GMP 调度模型深入分析

## 一、GMP 模型概述

### 1.1 三个核心概念

```
G (Goroutine)  - 用户态的轻量级线程
M (Machine)    - OS 线程的抽象
P (Processor)  - 调度的上下文，持有本地运行队列
```

### 1.2 为什么需要 GMP？

**旧的 GM 模型问题**：
1. 全局锁竞争：所有 M 共享一个全局运行队列
2. M 频繁切换：导致额外的延迟
3. 内存缓存局部性差：M 之间传递 G

**GMP 模型优势**：
1. **P 的本地队列**：减少锁竞争（无锁操作）
2. **Work Stealing**：空闲的 M 可以从其他 P 偷取 G
3. **Hand-off**：当 M 阻塞时，将 P 转移给其他 M
4. **内存缓存**：每个 P 有自己的 mcache

---

## 二、核心数据结构详解

### 2.1 g (Goroutine) 结构

**源码位置**: `runtime/runtime2.go:422`

```go
type g struct {
    // 栈信息
    stack       stack   // [stack.lo, stack.hi) 栈内存范围
    stackguard0 uintptr // 栈增长检查点，用于栈溢出检测和抢占
    stackguard1 uintptr // 用于 g0 和 gsignal 栈

    // 链接关系
    m         *m      // 当前运行在哪个 M 上
    sched     gobuf   // 调度上下文（PC、SP、BP 等）

    // 状态
    atomicstatus atomic.Uint32  // G 的状态
    goid         uint64          // goroutine ID
    schedlink    guintptr        // 链表指针（用于队列）

    // 抢占相关
    preempt       bool  // 抢占信号
    preemptStop   bool  // 是否转换到 _Gpreempted
    preemptShrink bool  // 是否在安全点收缩栈

    // 其他
    _panic    *_panic  // panic 链表
    _defer    *_defer  // defer 链表
    lockedm   muintptr // 锁定到特定的 M
}
```

**G 的状态机**：
```
_Gidle        - 刚分配，尚未初始化
_Grunnable    - 在运行队列中，等待执行
_Grunning     - 正在执行
_Gsyscall     - 正在执行系统调用
_Gwaiting     - 被阻塞（channel、锁等）
_Gdead        - 已退出，可以被重用
_Gpreempted   - 被抢占，等待恢复
```

**gobuf 结构**（保存调度上下文）：
```go
type gobuf struct {
    sp   uintptr  // 栈指针
    pc   uintptr  // 程序计数器
    g    guintptr // 指向 g
    ctxt unsafe.Pointer
    ret  uintptr  // 返回值
    lr   uintptr  // 链接寄存器（ARM）
    bp   uintptr  // 基址指针（用于栈回溯）
}
```

---

### 2.2 m (Machine) 结构

**源码位置**: `runtime/runtime2.go:552`

```go
type m struct {
    // 核心 goroutine
    g0      *g     // 调度专用的 goroutine（栈 ~64KB）
    curg    *g     // 当前正在运行的 goroutine

    // 关联的 P
    p       puintptr // 当前关联的 P（执行 Go 代码时）
    nextp   puintptr // 唤醒 M 时关联的 P
    oldp    puintptr // 系统调用前关联的 P

    // 线程本地存储
    tls     [tlsSlots]uintptr  // TLS 槽位

    // 状态
    id          int64
    spinning    bool  // 是否正在寻找可运行的 G
    blocked     bool  // 是否被阻塞在 note 上

    // 信号处理
    gsignal     *g    // 信号处理专用的 g

    // 锁
    locks       int32

    // CGO
    incgo       bool   // 是否在执行 cgo 调用
    ncgocall    uint64 // cgo 调用总数

    // 其他
    park        note       // M 的休眠/唤醒机制
    alllink     *m         // 全局 M 链表
    schedlink   muintptr   // 调度器的 M 链表
}
```

**M 的生命周期**：
```
创建 → 关联P → 执行G → 系统调用/阻塞 → Hand-off P → 休眠/销毁
```

---

### 2.3 p (Processor) 结构

**源码位置**: `runtime/runtime2.go:649`

```go
type p struct {
    id          int32
    status      uint32     // pidle/prunning/psyscall/pgcstop/pdead

    // 关联的 M
    m           muintptr   // 反向链接到 M

    // 本地运行队列（无锁）
    runqhead uint32
    runqtail uint32
    runq     [256]guintptr  // 本地队列（环形缓冲区）
    runnext  guintptr       // 下一个优先运行的 G

    // 内存分配
    mcache      *mcache    // 内存分配缓存

    // GC 相关
    gcBgMarkWorker       guintptr
    gcw                  gcWork

    // 对象池
    deferpool    []*_defer
    gFree struct {
        gList
        n int32
    }

    // 计数器
    schedtick   uint32  // 每次调度递增
    syscalltick uint32  // 每次系统调用递增
}
```

**P 的状态**：
```
_Pidle      - 空闲，在空闲列表中
_Prunning   - 正在运行 G
_Psyscall   - 在系统调用中
_Pgcstop    - GC 暂停
_Pdead      - 不再使用
```

**本地队列结构**：
```
┌─────────────────────────────────┐
│  runnext (优先级最高)             │
└─────────────────────────────────┘
         ↓ (如果 runnext 已满)
┌─────────────────────────────────┐
│  runq[256] 环形队列               │
│  head → [...G...G...G...] ← tail │
└─────────────────────────────────┘
         ↓ (如果本地队列满)
┌─────────────────────────────────┐
│  全局运行队列 (需要加锁)           │
└─────────────────────────────────┘
```

---

### 2.4 全局调度器 sched

**源码位置**: `runtime/runtime2.go`

```go
type schedt struct {
    lock mutex

    // 全局运行队列
    runq     gQueue
    runqsize int32

    // 空闲 P 列表
    pidle      puintptr
    npidle     atomic.Int32
    nmspinning atomic.Int32  // 正在自旋的 M 数量

    // 空闲 M 列表
    midle        muintptr
    nmidle       int32
    nmidlelocked int32

    // M 统计
    mnext  int64  // 下一个 M 的 ID
    maxmcount int32  // 最大 M 数量（默认 10000）

    // GC 等待的 G
    gFree struct {
        lock    mutex
        stack   gList  // 有栈的 dead G
        noStack gList  // 无栈的 dead G
        n       int32
    }

    // 系统监控
    sysmonlock  mutex

    // 时间
    lastpoll    atomic.Int64
    pollUntil   atomic.Int64
}
```

---

## 三、Runtime 初始化流程详解

### 3.1 schedinit() 函数分析

**源码位置**: `runtime/proc.go:782`

```go
func schedinit() {
    // 1. 初始化所有的锁
    lockInit(&sched.lock, lockRankSched)
    lockInit(&sched.sysmonlock, lockRankSysmon)
    // ... 更多锁初始化

    // 2. 获取当前 g (g0)
    gp := getg()

    // 3. Race detector 初始化
    if raceenabled {
        gp.racectx, raceprocctx0 = raceinit()
    }

    // 4. 设置最大 M 数量
    sched.maxmcount = 10000

    // 5. 世界初始处于停止状态
    worldStopped()

    // 6. 核心子系统初始化
    ticks.init()           // 时钟初始化
    moduledataverify()     // 验证模块数据
    stackinit()            // 栈池初始化
    mallocinit()           // 内存分配器初始化
    cpuinit(godebug)       // CPU 特性检测
    randinit()             // 随机数初始化
    alginit()              // 哈希算法初始化
    mcommoninit(gp.m, -1)  // M 初始化（m0）
    modulesinit()          // 模块初始化
    typelinksinit()        // 类型链接初始化
    itabsinit()            // 接口表初始化
    stkobjinit()           // 栈对象初始化

    // 7. 命令行参数和环境变量
    goargs()
    goenvs()
    secure()
    checkfds()
    parsedebugvars()

    // 8. GC 初始化
    gcinit()

    // 9. 确定 P 的数量（GOMAXPROCS）
    procs := ncpu
    if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
        procs = n
    }

    // 10. 创建和初始化 P
    if procresize(procs) != nil {
        throw("unknown runnable goroutine during bootstrap")
    }

    // 11. 世界启动
    worldStarted()
}
```

---

### 3.2 stackinit() - 栈初始化

```go
// 栈大小范围：2KB ~ 1GB
// 小对象栈：使用固定大小的缓存池
// 大对象栈：从堆分配

var stackpool [_NumStackOrders]struct {
    item stackpoolItem
    _    [cpu.CacheLinePadSize - unsafe.Sizeof(stackpoolItem{})%cpu.CacheLinePadSize]byte
}

// stackinit 初始化栈缓存
func stackinit() {
    // 初始化全局栈池
    for i := range stackpool {
        stackpool[i].item.span.init()
    }
    // 初始化大栈缓存
    for i := range stackLarge.free {
        stackLarge.free[i].init()
    }
}
```

---

### 3.3 mallocinit() - 内存分配器初始化

内存分配器使用 **TCMalloc** 思想：

```
┌──────────────────────────────────────┐
│           Central Cache               │  (全局，需要锁)
│   mcentral[_NumSizeClasses]           │
└──────────────────────────────────────┘
                 ↑
                 │ (缓存不足时)
                 │
┌──────────────────────────────────────┐
│      Thread Cache (每个P一个)          │  (本地，无锁)
│      mcache                           │
│   - tiny allocator (< 16B)           │
│   - small allocator (16B - 32KB)     │
│   - large allocator (> 32KB)         │
└──────────────────────────────────────┘
```

**Size Classes**：Go 预定义了 67 个大小类别（8B, 16B, 32B, ...）

---

### 3.4 procresize(nprocs) - 创建 P

**源码位置**: `runtime/proc.go:5671`

```go
func procresize(nprocs int32) *p {
    assertLockHeld(&sched.lock)
    assertWorldStopped()

    old := gomaxprocs

    // 1. 扩展 allp 数组
    if nprocs > int32(len(allp)) {
        lock(&allpLock)
        if nprocs <= int32(cap(allp)) {
            allp = allp[:nprocs]
        } else {
            nallp := make([]*p, nprocs)
            copy(nallp, allp[:cap(allp)])
            allp = nallp
        }
        unlock(&allpLock)
    }

    // 2. 初始化新的 P
    for i := old; i < nprocs; i++ {
        pp := allp[i]
        if pp == nil {
            pp = new(p)
        }
        pp.init(i)
        atomicstorep(unsafe.Pointer(&allp[i]), unsafe.Pointer(pp))
    }

    // 3. 关联 m0 和 allp[0]
    gp := getg()
    if gp.m.p != 0 && gp.m.p.ptr().id < nprocs {
        gp.m.p.ptr().status = _Prunning
    } else {
        gp.m.p = 0
        pp := allp[0]
        pp.m = 0
        pp.status = _Pidle
        acquirep(pp)  // m0 获取 p0
    }

    // 4. 释放多余的 P
    for i := nprocs; i < old; i++ {
        pp := allp[i]
        pp.destroy()
    }

    // 5. 将多余的 P 放入空闲列表
    var runnablePs *p
    for i := nprocs - 1; i >= 0; i-- {
        pp := allp[i]
        if gp.m.p.ptr() == pp {
            continue
        }
        pp.status = _Pidle
        if runqempty(pp) {
            pidleput(pp)  // 放入空闲列表
        } else {
            pp.m.set(mget())
            pp.link.set(runnablePs)
            runnablePs = pp
        }
    }

    return runnablePs
}
```

**p.init(id)** 初始化单个 P：
```go
func (pp *p) init(id int32) {
    pp.id = id
    pp.status = _Pidle
    pp.sudogcache = pp.sudogbuf[:0]
    pp.deferpool = pp.deferpoolbuf[:0]
    pp.wbBuf.reset()

    // 初始化 mcache
    if pp.mcache == nil {
        if id == 0 {
            pp.mcache = getg().m.mcache  // 引导阶段
        } else {
            pp.mcache = allocmcache()
        }
    }
}
```

---

### 3.5 runtime.main() - 主 Goroutine

**源码位置**: `runtime/proc.go:147`

```go
func main() {
    mp := getg().m

    // 1. 设置最大栈大小（64位：1GB，32位：250MB）
    if goarch.PtrSize == 8 {
        maxstacksize = 1000000000
    } else {
        maxstacksize = 250000000
    }

    mainStarted = true

    // 2. 启动系统监控线程（sysmon）
    if haveSysmon {
        systemstack(func() {
            newm(sysmon, nil, -1)
        })
    }

    // 3. 锁定到主 OS 线程
    lockOSThread()

    if mp != &m0 {
        throw("runtime.main not on m0")
    }

    // 4. 记录启动时间
    runtimeInitTime = nanotime()

    // 5. 执行 runtime 包的 init
    doInit(runtime_inittasks)

    defer func() {
        if needUnlock {
            unlockOSThread()
        }
    }()

    // 6. 启用 GC
    gcenable()

    main_init_done = make(chan bool)

    // 7. CGO 初始化（如果需要）
    if iscgo {
        // ... CGO 相关初始化
        startTemplateThread()
        cgocall(_cgo_notify_runtime_init_done, nil)
    }

    // 8. 执行所有包的 init 函数
    for m := &firstmoduledata; m != nil; m = m.next {
        doInit(m.inittasks)
    }

    close(main_init_done)

    needUnlock = false
    unlockOSThread()

    // 9. 调用 main.main
    fn := main_main
    fn()  // 间接调用，链接器处理

    // 10. main.main 返回后的清理
    if raceenabled {
        runExitHooks(0)
        racefini()
    }

    // 11. 等待其他 goroutine 的 panic 完成
    if runningPanicDefers.Load() != 0 {
        for c := 0; c < 1000; c++ {
            if runningPanicDefers.Load() == 0 {
                break
            }
            Gosched()
        }
    }

    if panicking.Load() != 0 {
        gopark(nil, nil, waitReasonPanicWait, traceBlockForever, 1)
    }

    runExitHooks(0)

    // 12. 退出进程
    exit(0)

    // 不应该到达这里
    for {
        var x *int32
        *x = 0
    }
}
```

---

### 3.6 sysmon - 系统监控线程

**源码位置**: `runtime/proc.go`

```go
func sysmon() {
    lock(&sched.lock)
    sched.nmsys++
    checkdead()
    unlock(&sched.lock)

    lasttrace := int64(0)
    idle := 0
    delay := uint32(0)

    for {
        if idle == 0 {
            delay = 20  // 开始时快速检查
        } else if idle > 50 {
            delay = 10 * 1000  // 空闲时慢速检查
        }
        usleep(delay)

        // 1. 检查死锁
        if t := (gcTrigger{kind: gcTriggerTime}); t.test() && forcegc.idle.Load() {
            lock(&forcegc.lock)
            forcegc.idle.Store(false)
            forcegc.g.schedlink = 0
            injectglist(&forcegc.g)
            unlock(&forcegc.lock)
        }

        // 2. 触发定时 GC
        if lastpoll != 0 && lastpoll+10*1000*1000*1000 < now {
            netpoll(0)
        }

        // 3. 抢占长时间运行的 G
        if retake(now) != 0 {
            idle = 0
        } else {
            idle++
        }

        // 4. 归还内存给 OS
        if shouldReleaseMemory() {
            scavenge()
        }
    }
}
```

**sysmon 的职责**：
1. **抢占调度**：抢占运行超过 10ms 的 goroutine
2. **网络轮询**：调用 netpoll 检查网络 I/O
3. **GC 触发**：定时触发 GC
4. **死锁检测**：检测全局死锁
5. **内存回收**：归还内存给操作系统

---

## 四、GMP 调度流程

### 4.1 创建 Goroutine

**go 语句** → `runtime.newproc` → `runtime.newproc1`

```go
// newproc 创建一个新的 goroutine
func newproc(fn *funcval) {
    gp := getg()
    pc := getcallerpc()
    systemstack(func() {
        newg := newproc1(fn, gp, pc, false, waitReasonZero)

        pp := getg().m.p.ptr()
        runqput(pp, newg, true)  // 放入 P 的本地队列

        if mainStarted {
            wakep()  // 唤醒或创建 M 来执行
        }
    })
}
```

**newproc1** 详细步骤：
```go
func newproc1(fn *funcval, callergp *g, callerpc uintptr, parked bool, waitreason waitReason) *g {
    mp := acquirem()  // 禁止抢占
    pp := mp.p.ptr()

    // 1. 尝试从 P 的本地缓存获取 dead G
    newg := gfget(pp)
    if newg == nil {
        // 2. 分配新的 G（初始栈 2KB）
        newg = malg(_StackMin)
        casgstatus(newg, _Gidle, _Gdead)
        allgadd(newg)  // 加入全局 G 列表
    }

    // 3. 计算栈顶
    totalSize := uintptr(4*goarch.PtrSize + sys.MinFrameSize)
    totalSize = alignUp(totalSize, sys.StackAlign)
    sp := newg.stack.hi - totalSize

    // 4. 初始化 gobuf
    memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
    newg.sched.sp = sp
    newg.sched.pc = abi.FuncPCABI0(goexit) + sys.PCQuantum
    newg.sched.g = guintptr(unsafe.Pointer(newg))
    gostartcallfn(&newg.sched, fn)

    // 5. 设置 G 的状态
    newg.gopc = callerpc
    newg.startpc = fn.fn
    casgstatus(newg, _Gdead, _Grunnable)

    // 6. 分配 goid
    newg.goid = pp.goidcache
    pp.goidcache++

    releasem(mp)
    return newg
}
```

---

### 4.2 调度循环

**M 的调度循环**: `runtime.mstart` → `runtime.mstart1` → `schedule`

```go
// schedule 调度循环的核心
func schedule() {
    mp := getg().m
    pp := mp.p.ptr()

top:
    // 1. GC 等待检查
    if sched.gcwaiting.Load() {
        gcstopm()
        goto top
    }

    // 2. 每 61 次调度，从全局队列获取一次
    if pp.schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)
        gp := globrunqget(pp, 1)
        unlock(&sched.lock)
        if gp != nil {
            return gp
        }
    }

    var gp *g
    var inheritTime bool

    // 3. 优先检查 runnext
    if gp == nil {
        gp, inheritTime = runqget(pp)
    }

    // 4. 从全局队列获取
    if gp == nil {
        gp, inheritTime = findRunnable()  // 阻塞直到找到
    }

    // 5. 执行 G
    execute(gp, inheritTime)
}
```

**findRunnable** 查找可运行的 G：
```go
func findRunnable() (gp *g, inheritTime bool, tryWakeP bool) {
    mp := getg().m
    pp := mp.p.ptr()

top:
    // 1. 本地队列
    if gp, inheritTime := runqget(pp); gp != nil {
        return gp, inheritTime, false
    }

    // 2. 全局队列
    if sched.runqsize != 0 {
        lock(&sched.lock)
        gp := globrunqget(pp, 0)
        unlock(&sched.lock)
        if gp != nil {
            return gp, false, false
        }
    }

    // 3. 网络轮询
    if netpollinited() && netpollAnyWaiters() {
        if list, delta := netpoll(0); !list.empty() {
            gp := list.pop()
            injectglist(&list)
            return gp, false, true
        }
    }

    // 4. Work Stealing - 从其他 P 偷取
    procs := gomaxprocs
    if mp.spinning || 2*sched.nmspinning.Load() < procs-sched.npidle.Load() {
        if !mp.spinning {
            mp.becomeSpinning()
        }

        gp, inheritTime, tnow, w, newWork := stealWork(now)
        if gp != nil {
            return gp, inheritTime, false
        }
    }

    // 5. 没有找到，休眠 M
    stopm()
    goto top
}
```

**stealWork** - 工作窃取：
```go
func stealWork(now int64) (gp *g, inheritTime bool, rnow, pollUntil int64, newWork bool) {
    pp := getg().m.p.ptr()

    ranTimer := false
    const stealTries = 4
    for i := 0; i < stealTries; i++ {
        stealTimersOrRunNextG := i == stealTries-1

        for enum := stealOrder.start(fastrand()); !enum.done(); enum.next() {
            if sched.gcwaiting.Load() {
                return nil, false, now, pollUntil, false
            }

            p2 := allp[enum.position()]
            if pp == p2 {
                continue
            }

            // 从 p2 的本地队列偷一半
            if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
                return gp, false, now, pollUntil, ranTimer
            }
        }
    }

    return nil, false, now, pollUntil, ranTimer
}
```

---

### 4.3 执行 Goroutine

```go
func execute(gp *g, inheritTime bool) {
    mp := getg().m

    // 1. 设置 G 的状态
    casgstatus(gp, _Grunnable, _Grunning)
    gp.waitsince = 0
    gp.preempt = false
    gp.stackguard0 = gp.stack.lo + _StackGuard

    if !inheritTime {
        mp.p.ptr().schedtick++
    }

    // 2. 关联 G 和 M
    mp.curg = gp
    gp.m = mp

    // 3. 跳转到 G 的代码（汇编实现）
    gogo(&gp.sched)
}
```

**gogo** (汇编 `runtime/asm_amd64.s`)：
```asm
TEXT runtime·gogo(SB), NOSPLIT, $0-8
    MOVQ    buf+0(FP), BX       // gobuf
    MOVQ    gobuf_g(BX), DX
    MOVQ    gobuf_sp(BX), SP    // 恢复 SP
    MOVQ    gobuf_ret(BX), AX
    MOVQ    gobuf_ctxt(BX), DX
    MOVQ    gobuf_bp(BX), BP
    MOVQ    $0, gobuf_sp(BX)
    MOVQ    $0, gobuf_ret(BX)
    MOVQ    $0, gobuf_ctxt(BX)
    MOVQ    $0, gobuf_bp(BX)
    MOVQ    gobuf_pc(BX), BX
    JMP     BX                  // 跳转到 G 的 PC
```

---

## 五、关键机制

### 5.1 抢占调度

**基于栈增长的协作式抢占**（Go 1.14 之前）：
- 函数调用时检查 `stackguard0`
- 如果被设置为 `stackPreempt`，触发抢占

**基于信号的异步抢占**（Go 1.14+）：
```go
// preemptone 请求抢占一个 G
func preemptone(pp *p) bool {
    mp := pp.m.ptr()
    if mp == nil || mp == getg().m {
        return false
    }

    gp := mp.curg
    if gp == nil || gp == mp.g0 {
        return false
    }

    gp.preempt = true
    gp.stackguard0 = stackPreempt

    // 异步抢占：发送信号
    if preemptMSupported && debug.asyncpreemptoff == 0 {
        pp.preempt = true
        preemptM(mp)
    }

    return true
}
```

### 5.2 系统调用处理

**进入系统调用**：
```go
func entersyscall() {
    reentersyscall(getcallerpc(), getcallersp())
}

func reentersyscall(pc, sp uintptr) {
    gp := getg()

    // 1. 保存调用者的 PC/SP
    gp.syscallsp = sp
    gp.syscallpc = pc

    // 2. 设置 G 的状态为 Gsyscall
    casgstatus(gp, _Grunning, _Gsyscall)

    // 3. 将 P 设置为 Psyscall
    pp := gp.m.p.ptr()
    pp.m = 0
    gp.m.oldp.set(pp)
    gp.m.p = 0
    atomic.Store(&pp.status, _Psyscall)

    // 4. sysmon 会检测长时间的系统调用并 Hand-off P
}
```

**退出系统调用**：
```go
func exitsyscall() {
    gp := getg()

    // 1. 快速路径：尝试重新获取原来的 P
    oldp := gp.m.oldp.ptr()
    if exitsyscallfast(oldp) {
        // 成功获取 P
        casgstatus(gp, _Gsyscall, _Grunning)
        gp.syscallsp = 0
        gp.m.p.ptr().syscalltick++
        return
    }

    // 2. 慢速路径：无法获取 P，调度到其他 M
    mcall(exitsyscall0)
}
```

---

## 六、总结

### 6.1 GMP 模型的核心优势

1. **并发性能**：
   - 本地队列减少锁竞争
   - Work Stealing 实现负载均衡
   - 每个 P 独立的内存缓存

2. **灵活调度**：
   - 协作式 + 抢占式结合
   - Hand-off 机制处理阻塞
   - sysmon 保证调度公平性

3. **资源利用**：
   - M 可以复用
   - P 数量可调（GOMAXPROCS）
   - G 栈动态增长（2KB ~ 1GB）

### 6.2 关键数据流

```
用户代码
  ↓
go 关键字
  ↓
runtime.newproc
  ↓
创建 G 并放入 P 的本地队列
  ↓
wakep / startm (唤醒或创建 M)
  ↓
M 执行调度循环 (schedule)
  ↓
findRunnable (本地队列 → 全局队列 → netpoll → Work Stealing)
  ↓
execute(gp)
  ↓
gogo → 跳转到 G 的代码
  ↓
G 执行完毕或阻塞
  ↓
goexit / park → 回到调度循环
```
