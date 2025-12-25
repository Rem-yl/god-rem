# GMP 模型源码深度解析

基于 Go 1.23.0 源码分析 GMP 调度模型的数据结构、理论和调度流程。

源码位置：`$GOROOT/src/runtime/`
- `runtime2.go` - 核心数据结构定义
- `proc.go` - 调度器实现 (约7000行代码)

---

## 一、GMP 模型概述

### 1.1 什么是 GMP

GMP 是 Go 运行时调度器的核心模型，由三个关键组件构成：

- **G (Goroutine)** - 用户级轻量线程，代表一个并发执行单元
- **M (Machine)** - 操作系统线程，真正执行代码的实体
- **P (Processor)** - 逻辑处理器，调度上下文，持有 G 的运行队列

### 1.2 设计理念

```
用户代码
   ↓
G (goroutine) ← 成千上万个，栈2KB起步
   ↓
P (processor) ← GOMAXPROCS个（默认=CPU核心数）
   ↓
M (thread) ← 按需创建，最多10000个
   ↓
OS 内核线程
```

**核心思想**：
1. **M:N 模型** - N个goroutine运行在M个OS线程上
2. **本地队列** - 每个P有本地runnable队列（256容量），减少全局锁竞争
3. **工作窃取** - 空闲的P可以从其他P的队列偷取goroutine
4. **抢占调度** - 防止单个goroutine长时间占用CPU
5. **Hand-off机制** - 系统调用时P可以移交给其他M

---

## 二、数据结构详解

### 2.1 G 结构体 (runtime2.go:422)

`type g struct` - Goroutine 的完整表示

#### 核心字段分类

**1. 栈管理字段**
```go
type g struct {
    // 栈内存范围: [stack.lo, stack.hi)
    stack       stack   // 栈的实际内存区域
    stackguard0 uintptr // 栈增长检测点 (用户栈)
    stackguard1 uintptr // 栈增长检测点 (系统栈g0)

    // 系统调用时的栈信息
    syscallsp uintptr // 系统调用时的SP
    syscallpc uintptr // 系统调用时的PC
    syscallbp uintptr // 系统调用时的BP
```

**关键点**：
- 初始栈大小：2KB
- 最大栈大小：1GB (Linux amd64)
- `stackguard0` 通常设为 `stack.lo + StackGuard(928字节)`
- 当 `SP < stackguard0` 时触发栈增长
- 抢占时可将 `stackguard0` 设为 `StackPreempt` (0xfffffffffffffade)

**2. 调度相关字段**
```go
    m         *m      // 当前运行这个g的M
    sched     gobuf   // 调度上下文（保存寄存器）

    atomicstatus atomic.Uint32 // G的状态
    goid         uint64         // goroutine ID (唯一)
    schedlink    guintptr       // 链表指针（用于队列）

    preempt       bool // 抢占信号
    preemptStop   bool // 是否需要停止到_Gpreempted
    preemptShrink bool // 是否需要收缩栈
```

**gobuf 结构 (runtime2.go:324)** - 保存调度上下文
```go
type gobuf struct {
    sp   uintptr // 栈指针
    pc   uintptr // 程序计数器
    g    guintptr // 指向g自己
    ctxt unsafe.Pointer // 上下文
    ret  uintptr // 返回值
    bp   uintptr // 基址指针(frame pointer)
}
```

**3. G 的状态** (runtime2.go:16-106)

```go
const (
    _Gidle      = 0 // 刚分配，未初始化
    _Grunnable  = 1 // 在运行队列中，等待执行
    _Grunning   = 2 // 正在执行，拥有M和P
    _Gsyscall   = 3 // 正在执行系统调用
    _Gwaiting   = 4 // 阻塞等待（如channel、sleep）
    _Gdead      = 6 // 已退出或在空闲列表
    _Gcopystack = 8 // 栈正在复制
    _Gpreempted = 9 // 被抢占暂停

    // GC扫描状态（与上述状态组合）
    _Gscan = 0x1000
    _Gscanrunnable = _Gscan + _Grunnable
    // ...
)
```

**状态转换示例**：
```
newproc → _Grunnable → (schedule) → _Grunning → (gopark) → _Gwaiting
                                              ↘ (goexit) → _Gdead
```

**4. 等待和阻塞字段**
```go
    waitsince    int64      // 阻塞开始时间
    waitreason   waitReason // 阻塞原因
    waiting      *sudog     // 等待列表（channel等）
```

**5. 调试和追踪字段**
```go
    goid         uint64  // goroutine ID
    parentGoid   uint64  // 创建者的goid
    gopc         uintptr // 创建此goroutine的PC
    startpc      uintptr // goroutine函数的PC
```

**6. 其他重要字段**
```go
    _panic    *_panic // panic链表
    _defer    *_defer // defer链表
    lockedm   muintptr // 锁定到特定M (LockOSThread)
    timer     *timer   // time.Sleep等使用的定时器
```

---

### 2.2 M 结构体 (runtime2.go:552)

`type m struct` - Machine (OS Thread) 的表示

#### 核心字段分类

**1. 关联的 G**
```go
type m struct {
    g0      *g     // 特殊g：调度栈（用于执行调度代码）
    gsignal *g     // 信号处理栈
    curg    *g     // 当前正在运行的用户goroutine
```

**关键点**：
- `g0`：每个M都有一个g0，栈大小8KB（Linux），用于执行运行时代码
- `curg`：用户goroutine，执行用户代码时 m.curg != nil
- `gsignal`：处理信号的特殊g

**2. 关联的 P**
```go
    p      puintptr // 当前关联的P (执行Go代码时)
    nextp  puintptr // 唤醒M时的目标P
    oldp   puintptr // 系统调用前的P
```

**3. M 的身份和状态**
```go
    id            int64    // M的ID
    spinning      bool     // 是否处于自旋状态（找工作中）
    blocked       bool     // 是否在note上阻塞
    locks         int32    // 持有锁的计数
    mallocing     int32    // 当前是否在分配内存
    throwing      throwType
    preemptoff    string   // 如果非空，禁止抢占curg
```

**spinning 状态详解**：
```go
// M 处于spinning状态表示：
// 1. M没有运行用户代码
// 2. M正在积极寻找可运行的goroutine
// 3. nmspinning计数器会增加
//
// spinning → non-spinning 的时机：
// - 找到了可运行的goroutine
// - 需要进入睡眠等待
```

**4. 线程本地存储**
```go
    tls [6]uintptr // 线程本地存储 (TLS)
```

**5. 系统调用相关**
```go
    syscalltick uint32 // 系统调用计数
    incgo       bool   // 是否在执行cgo调用
```

**6. M 链表**
```go
    alllink   *m       // 全局M链表 (runtime.allm)
    schedlink muintptr // 调度器空闲M链表
    freelink  *m       // 等待释放的M链表
```

**7. 创建和销毁**
```go
    mstartfn  func()         // M启动时执行的函数
    freeWait  atomic.Uint32  // 是否可以释放g0和删除m
```

---

### 2.3 P 结构体 (runtime2.go:649)

`type p struct` - Processor (逻辑处理器)

#### 核心字段分类

**1. P 的身份和状态**
```go
type p struct {
    id          int32  // P的ID (0 到 GOMAXPROCS-1)
    status      uint32 // P的状态
    link        puintptr // 空闲P链表
    m           muintptr // 关联的M (nil表示空闲)
```

**P 的状态** (runtime2.go:108-156)
```go
const (
    _Pidle    = 0 // 空闲，在空闲列表或状态转换中
    _Prunning = 1 // 被M拥有，执行用户代码或调度器
    _Psyscall = 2 // 不运行用户代码，但与syscall中的M关联
    _Pgcstop  = 3 // 为STW停止，由stopTheWorld的M拥有
    _Pdead    = 4 // 不再使用 (GOMAXPROCS缩小)
)
```

**状态转换示例**：
```
_Pidle → (acquirep) → _Prunning → (releasep) → _Pidle
                              ↘ (entersyscall) → _Psyscall
                                                      ↘ (exitsyscall) → _Prunning
```

**2. 本地运行队列** - 核心！
```go
    // 本地runnable goroutine队列 - 无锁访问
    runqhead uint32       // 队列头
    runqtail uint32       // 队列尾
    runq     [256]guintptr // 循环队列，容量256

    // runnext：优先运行的goroutine
    // 如果当前goroutine创建了新goroutine，新g会放到runnext
    // 这样可以保持局部性，减少延迟
    runnext guintptr
```

**本地队列工作原理**：
```
newproc创建g → runqput
                ↓
           优先放入runnext (如果为空)
                ↓
           如果runnext已有g，踢出到runq队列尾
                ↓
           如果runq满了(256个)，一半移到全局队列
```

**3. 调度计数器**
```go
    schedtick   uint32 // 每次调度器调用递增
    syscalltick uint32 // 每次系统调用递增
    sysmontick  sysmontick // sysmon观察到的最后tick
```

**4. 缓存和对象池**
```go
    mcache *mcache // 内存分配缓存

    // defer pool
    deferpool    []*_defer
    deferpoolbuf [32]*_defer

    // 空闲G列表 (status == Gdead)
    gFree struct {
        gList
        n int32
    }

    // goroutine ID缓存
    goidcache    uint64
    goidcacheend uint64
```

**5. GC 相关**
```go
    gcAssistTime         int64 // GC辅助时间
    gcFractionalMarkTime int64 // 分数标记时间
    gcMarkWorkerMode     gcMarkWorkerMode // GC标记worker模式
    gcw                  gcWork // GC工作缓冲
    wbBuf                wbBuf  // 写屏障缓冲
```

**6. 定时器**
```go
    timers timers // P拥有的定时器堆
```

---

### 2.4 全局调度器 schedt (runtime2.go:775)

`type schedt struct` - 全局调度器状态

```go
type schedt struct {
    lock mutex // 全局调度器锁

    // M 相关
    midle        muintptr // 空闲M链表
    nmidle       int32    // 空闲M数量
    nmidlelocked int32    // 锁定的空闲M数量
    mnext        int64    // 下一个M的ID
    maxmcount    int32    // M的最大数量 (默认10000)

    // P 相关
    pidle  puintptr     // 空闲P链表
    npidle atomic.Int32 // 空闲P数量
    nmspinning atomic.Int32 // 自旋M的数量

    // 全局runnable队列
    runq     gQueue // 全局队列 (需要lock保护)
    runqsize int32  // 全局队列长度

    // goroutine ID生成器
    goidgen atomic.Uint64

    // 空闲G缓存池
    gFree struct {
        lock    mutex
        stack   gList // 有栈的G
        noStack gList // 无栈的G
        n       int32
    }

    // GC 相关
    gcwaiting  atomic.Bool // GC等待运行
    stopwait   int32       // STW等待的P数量

    // 系统监控
    lastpoll  atomic.Int64  // 最后一次网络轮询时间
    pollUntil atomic.Int64  // 轮询睡眠到的时间
}
```

**全局变量**：
```go
// runtime/proc.go 中定义的全局变量
var (
    allm       *m     // 所有M的链表
    gomaxprocs int32  // P的数量
    ncpu       int32  // CPU核心数
    sched      schedt // 全局调度器
    allp       []*p   // 所有P的数组 [GOMAXPROCS]
)
```

---

## 三、GMP 调度理论

### 3.1 调度器初始化 (schedinit)

**源码位置**: `runtime/proc.go:782`

```go
func schedinit() {
    // 1. 获取当前g (g0)
    gp := getg()

    // 2. 设置最大M数量
    sched.maxmcount = 10000

    // 3. 初始化栈、内存分配器
    stackinit()
    mallocinit()

    // 4. 初始化当前M (m0)
    mcommoninit(gp.m, -1)

    // 5. 初始化CPU个数
    cpuinit()       // 必须在alginit之前
    alginit()       // maps不能在此前使用

    // 6. 模块初始化
    modulesinit()   // 提供activeModules
    typelinksinit() // 使用maps、activeModules
    itabsinit()     // 使用activeModules

    // 7. 初始化命令行参数、环境变量
    goargs()
    goenvs()

    // 8. 处理GODEBUG、GOTRACEBACK等
    parsedebugvars()

    // 9. GC初始化
    gcinit()

    // 10. 设置GOMAXPROCS，创建P
    // 读取环境变量GOMAXPROCS，默认为CPU核心数
    procs := ncpu
    if n := atoi(gogetenv("GOMAXPROCS")); n > 0 {
        procs = n
    }

    // 调整P的数量，创建P数组
    if procresize(procs) != nil {
        throw("unknown runnable goroutine during bootstrap")
    }
}
```

**procresize 函数** - 创建/调整P数组
```go
func procresize(nprocs int32) *p {
    // 1. 获取旧的P数量
    old := gomaxprocs

    // 2. 调整allp切片大小
    if nprocs > int32(len(allp)) {
        // 扩容allp
        allp = append(allp, make([]*p, nprocs-old)...)
    }

    // 3. 初始化新的P
    for i := old; i < nprocs; i++ {
        pp := allp[i]
        if pp == nil {
            pp = new(p)
        }
        pp.init(i) // 初始化P的ID、状态、队列等
        allp[i] = pp
    }

    // 4. 释放多余的P
    for i := nprocs; i < old; i++ {
        pp := allp[i]
        pp.destroy() // 清理P的资源
    }

    // 5. 更新gomaxprocs
    gomaxprocs = nprocs

    // 6. 将所有P放入空闲列表，除了当前P
    var runnablePs *p
    for i := nprocs - 1; i >= 0; i-- {
        pp := allp[i]
        if gp.m.p.ptr() == pp {
            pp.status = _Prunning // 当前P保持运行
        } else {
            pp.status = _Pidle
            pidleput(pp) // 放入空闲列表
        }
    }

    return runnablePs
}
```

**初始化后的状态**：
```
┌─────────┐
│  allp   │ → [P0, P1, P2, ..., P(GOMAXPROCS-1)]
└─────────┘
     ↓
   P0: status=_Prunning, m=m0
   P1: status=_Pidle, m=nil
   P2: status=_Pidle, m=nil
   ...

sched.pidle → P1 → P2 → P3 → ... → nil
```

---

### 3.2 M 的启动 (mstart)

**源码位置**: `runtime/proc.go:1463`

```go
// mstart 是新M的入口点 (汇编代码)
// 调用 mstart0
func mstart0() {
    gp := getg()

    // 初始化g0的栈边界
    osStack := gp.stack.lo == 0
    if osStack {
        // 从系统栈初始化
        size := gp.stack.hi
        if size == 0 {
            size = 8192 * sys.StackGuardMultiplier
        }
        gp.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
        gp.stack.lo = gp.stack.hi - size + 1024
    }

    // 设置栈保护
    gp.stackguard0 = gp.stack.lo + stackGuard
    gp.stackguard1 = gp.stackguard0

    // 调用mstart1
    mstart1()
}

func mstart1() {
    gp := getg()

    // 记录调用者信息
    if gp != gp.m.g0 {
        throw("bad runtime·mstart")
    }

    // 保存调用者，用于mcall
    gp.sched.g = guintptr(unsafe.Pointer(gp))
    gp.sched.pc = getcallerpc()
    gp.sched.sp = getcallersp()

    // 执行启动函数 (如果有)
    if fn := gp.m.mstartfn; fn != nil {
        fn()
    }

    // 如果M不是m0，需要acquirep
    if gp.m != &m0 {
        acquirep(gp.m.nextp.ptr())
        gp.m.nextp = 0
    }

    // 进入调度循环 - 永不返回！
    schedule()
}
```

---

### 3.3 调度循环 (schedule)

**源码位置**: `runtime/proc.go:3966`

```go
// schedule的核心：找到一个runnable的goroutine并执行
// 永不返回
func schedule() {
    mp := getg().m

    // 每61次调度，检查全局队列
    // 防止全局队列饥饿
    if mp.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)
        gp := globrunqget(mp.p.ptr(), 1)
        unlock(&sched.lock)
        if gp != nil {
            return gp
        }
    }

    // 查找runnable goroutine
    gp, inheritTime, tryWakeP := findRunnable()

    // 执行goroutine
    execute(gp, inheritTime)
}
```

**findRunnable 函数** - 查找可运行的G

**源码位置**: `runtime/proc.go:3249`

```go
// findRunnable 尝试找到一个可运行的goroutine来执行
// 不能拆分栈，因为它可能会移动P
func findRunnable() (gp *g, inheritTime, tryWakeP bool) {
    mp := getg().m

top:
    pp := mp.p.ptr()

    // 1. 尝试从本地队列获取
    if gp, inheritTime := runqget(pp); gp != nil {
        return gp, inheritTime, false
    }

    // 2. 尝试从全局队列获取
    if sched.runqsize != 0 {
        lock(&sched.lock)
        gp := globrunqget(pp, 0)
        unlock(&sched.lock)
        if gp != nil {
            return gp, false, false
        }
    }

    // 3. 轮询网络 (非阻塞)
    if netpollinited() && netpollWaiters.Load() > 0 {
        if list, delta := netpoll(0); !list.empty() {
            gp := list.pop()
            // 将其他网络就绪的G放入本地队列
            injectglist(&list)
            return gp, false, true
        }
    }

    // 4. 尝试从其他P窃取工作 (Work Stealing)
    procs := gomaxprocs
    if mp.spinning || 2*sched.nmspinning.Load() < procs-sched.npidle.Load() {
        if !mp.spinning {
            mp.becomeSpinning()
        }

        // 随机选择一个P开始窃取
        offset := fastrand() % uint32(procs)
        for i := 0; i < procs; i++ {
            targetP := allp[(offset+i)%procs]

            // 从targetP窃取一半的G
            if gp := runqsteal(pp, targetP, true); gp != nil {
                return gp, false, false
            }
        }
    }

    // 5. 再次检查全局队列
    if sched.runqsize != 0 {
        lock(&sched.lock)
        gp := globrunqget(pp, 0)
        unlock(&sched.lock)
        if gp != nil {
            return gp, false, false
        }
    }

    // 6. 检查netpoll (可能阻塞)
    if netpollinited() && (netpollWaiters.Load() > 0 || pollUntil != 0) {
        if list, delta := netpoll(delay); !list.empty() {
            gp := list.pop()
            injectglist(&list)
            return gp, false, true
        }
    }

    // 7. 没有工作了，停止spinning，准备休眠
    stopm()
    goto top
}
```

**工作窃取算法** (runqsteal):
```go
// 从p2偷取一半的goroutine到p1
func runqsteal(pp, p2 *p, stealRunNextG bool) *g {
    t := pp.runqtail
    n := runqgrab(p2, &pp.runq, t, stealRunNextG)
    if n == 0 {
        return nil
    }
    n--
    gp := pp.runq[(t+n)%uint32(len(pp.runq))].ptr()
    if n == 0 {
        return gp
    }
    pp.runqtail = t + n
    return gp
}

// runqgrab 从p2批量窃取到batch数组
func runqgrab(pp *p, batch *[256]guintptr, batchHead uint32, stealRunNextG bool) uint32 {
    for {
        h := atomic.LoadAcq(&pp.runqhead)
        t := atomic.LoadAcq(&pp.runqtail)
        n := t - h
        n = n - n/2 // 窃取一半

        if n == 0 {
            // 尝试窃取runnext
            if stealRunNextG {
                if gp := pp.runnext.ptr(); gp != nil {
                    if pp.status == _Prunning {
                        // 只有在P运行时才窃取runnext
                        if !pp.runnext.cas(gp, nil) {
                            continue
                        }
                        batch[batchHead%uint32(len(batch))] = guintptr(unsafe.Pointer(gp))
                        return 1
                    }
                }
            }
            return 0
        }

        // 复制goroutine到batch
        for i := uint32(0); i < n; i++ {
            g := pp.runq[(h+i)%uint32(len(pp.runq))].ptr()
            batch[(batchHead+i)%uint32(len(batch))] = guintptr(unsafe.Pointer(g))
        }

        if atomic.CasRel(&pp.runqhead, h, h+n) {
            return n
        }
    }
}
```

---

### 3.4 执行 G (execute)

**源码位置**: `runtime/proc.go:3221`

```go
// execute 标记gp为running并开始执行
func execute(gp *g, inheritTime bool) {
    mp := getg().m

    // 将gp与M关联
    mp.curg = gp
    gp.m = mp

    // 设置g的状态为running
    casgstatus(gp, _Grunnable, _Grunning)
    gp.waitsince = 0
    gp.preempt = false
    gp.stackguard0 = gp.stack.lo + stackGuard

    if !inheritTime {
        // 新时间片
        mp.p.ptr().schedtick++
    }

    // 检查是否禁用了抢占
    if gp.m.locks != 0 {
        throw("execute: m has locks")
    }

    // gogo在汇编中实现，切换到gp的栈并执行
    // 从gp.sched恢复寄存器：SP, PC等
    gogo(&gp.sched)
}
```

**gogo 函数** (runtime/asm_amd64.s) - 上下文切换
```asm
// func gogo(buf *gobuf)
// 从gobuf恢复状态并执行
TEXT runtime·gogo(SB), NOSPLIT, $0-8
    MOVQ buf+0(FP), BX      // BX = gobuf
    MOVQ gobuf_g(BX), DX    // DX = gp
    MOVQ 0(DX), CX          // 确保gp != nil

    // 保存当前g0的上下文 (已在schedule中完成)

    // 切换到gp的栈
    MOVQ gobuf_sp(BX), SP   // SP = gp.sched.sp
    MOVQ gobuf_bp(BX), BP   // BP = gp.sched.bp
    MOVQ gobuf_ret(BX), AX  // 返回值
    MOVQ gobuf_ctxt(BX), DX // 上下文

    // 清零gobuf
    MOVQ $0, gobuf_sp(BX)
    MOVQ $0, gobuf_bp(BX)
    MOVQ $0, gobuf_ret(BX)
    MOVQ $0, gobuf_ctxt(BX)

    // 切换g到TLS
    MOVQ DX, g(CX)
    MOVQ DX, TLS

    // 跳转到gp.sched.pc执行
    MOVQ gobuf_pc(BX), BX
    JMP BX
```

---

### 3.5 G 的创建 (newproc)

**源码位置**: `runtime/proc.go:4754`

```go
// newproc 创建一个新的goroutine
// 编译器将 go func() 转换为 newproc 调用
func newproc(fn *funcval) {
    gp := getg()
    pc := getcallerpc()

    // 在系统栈上执行newproc1
    systemstack(func() {
        newg := newproc1(fn, gp, pc, false, waitReasonZero)

        pp := getg().m.p.ptr()
        runqput(pp, newg, true) // 放入P的本地队列

        if mainStarted {
            wakep() // 唤醒一个P来执行新的goroutine
        }
    })
}

// newproc1 创建新的goroutine结构
func newproc1(fn *funcval, callergp *g, callerpc uintptr, parked bool, waitreason waitReason) *g {
    mp := getg().m
    pp := mp.p.ptr()

    // 1. 获取一个g结构（从空闲列表或新分配）
    newg := gfget(pp)
    if newg == nil {
        newg = malg(_StackMin) // 2KB栈
        casgstatus(newg, _Gidle, _Gdead)
        allgadd(newg) // 添加到allgs
    }

    // 2. 计算栈顶和参数位置
    totalSize := uintptr(4*goarch.PtrSize) // 4个字
    totalSize = alignUp(totalSize, sys.StackAlign)
    sp := newg.stack.hi - totalSize

    // 3. 初始化g的sched
    memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
    newg.sched.sp = sp
    newg.stktopsp = sp
    newg.sched.pc = abi.FuncPCABI0(goexit) + sys.PCQuantum
    newg.sched.g = guintptr(unsafe.Pointer(newg))

    // 调整PC，实际从fn开始执行
    gostartcallfn(&newg.sched, fn)

    // 4. 设置goroutine属性
    newg.parentGoid = callergp.goid
    newg.gopc = callerpc
    newg.startpc = fn.fn

    // 5. 设置状态为runnable
    casgstatus(newg, _Gdead, _Grunnable)

    // 6. 分配goid
    if pp.goidcache == pp.goidcacheend {
        // 从sched.goidgen批量获取
        pp.goidcache = sched.goidgen.Add(_GoidCacheBatch)
        pp.goidcacheend = pp.goidcache + _GoidCacheBatch
    }
    newg.goid = pp.goidcache
    pp.goidcache++

    return newg
}
```

**runqput** - 将G放入P的本地队列
```go
func runqput(pp *p, gp *g, next bool) {
    if next {
    retryNext:
        oldnext := pp.runnext
        if !pp.runnext.cas(oldnext, gp) {
            goto retryNext
        }
        if oldnext == 0 {
            return
        }
        // 踢出旧的runnext
        gp = oldnext.ptr()
    }

retry:
    h := atomic.LoadAcq(&pp.runqhead)
    t := pp.runqtail
    if t-h < uint32(len(pp.runq)) {
        // 队列未满，放入队尾
        pp.runq[t%uint32(len(pp.runq))].set(gp)
        atomic.StoreRel(&pp.runqtail, t+1)
        return
    }

    // 队列满了，放一半到全局队列
    if runqputslow(pp, gp, h, t) {
        return
    }
    goto retry
}
```

---

### 3.6 系统调用处理

**entersyscall** - 进入系统调用
```go
func entersyscall() {
    // 禁止抢占
    reentersyscall(getcallerpc(), getcallersp())
}

func reentersyscall(pc, sp uintptr) {
    gp := getg()

    // 保存当前PC和SP，GC时使用
    save(pc, sp)

    gp.syscallsp = sp
    gp.syscallpc = pc

    // 设置状态为Gsyscall
    casgstatus(gp, _Grunning, _Gsyscall)

    pp := gp.m.p.ptr()
    pp.m = 0 // 解除M和P的关联
    gp.m.oldp.set(pp) // 保存P
    gp.m.p = 0

    // 设置P为syscall状态
    atomic.Store(&pp.status, _Psyscall)

    // sysmon会检测长时间syscall并抢夺P
}
```

**exitsyscall** - 退出系统调用
```go
func exitsyscall() {
    gp := getg()

    gp.waitsince = 0
    oldp := gp.m.oldp.ptr()
    gp.m.oldp = 0

    // 尝试快速重新获取同一个P
    if exitsyscallfast(oldp) {
        // 成功获取P
        casgstatus(gp, _Gsyscall, _Grunning)
        return
    }

    // 慢路径：没有空闲P，需要调度
    mcall(exitsyscall0)
}

func exitsyscallfast(oldp *p) bool {
    // 1. 尝试重新获取oldp
    if oldp != nil && oldp.status == _Psyscall && atomic.Cas(&oldp.status, _Psyscall, _Pidle) {
        wirep(oldp)
        return true
    }

    // 2. 尝试获取任意空闲P
    if sched.pidle != 0 {
        lock(&sched.lock)
        pp := pidleget()
        unlock(&sched.lock)
        if pp != nil {
            wirep(pp)
            return true
        }
    }

    return false
}
```

---

## 四、调度时机

### 4.1 主动调度

**1. runtime.Gosched()**
```go
func Gosched() {
    // 主动让出CPU
    mcall(gosched_m)
}

func gosched_m(gp *g) {
    // 将gp放回全局队列
    goschedImpl(gp, false)
}

func goschedImpl(gp *g, preempted bool) {
    status := readgstatus(gp)
    casgstatus(gp, status, _Grunnable)
    dropg() // 解除M和gp的关联

    lock(&sched.lock)
    globrunqput(gp) // 放入全局队列
    unlock(&sched.lock)

    schedule() // 重新调度
}
```

**2. channel 操作阻塞**
```go
// gopark 阻塞当前goroutine
func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int) {
    mp := acquirem()
    gp := mp.curg

    mp.waitlock = lock
    mp.waitunlockf = unlockf
    gp.waitreason = reason
    mp.waitTraceBlockReason = traceReason
    mp.waitTraceSkip = traceskip

    releasem(mp)

    mcall(park_m)
}

func park_m(gp *g) {
    mp := getg().m

    // 设置状态为Gwaiting
    casgstatus(gp, _Grunning, _Gwaiting)
    dropg()

    // 执行unlock函数
    if fn := mp.waitunlockf; fn != nil {
        ok := fn(gp, mp.waitlock)
        mp.waitunlockf = nil
        mp.waitlock = nil
        if !ok {
            casgstatus(gp, _Gwaiting, _Grunnable)
            execute(gp, true)
        }
    }

    schedule()
}
```

### 4.2 被动调度（抢占）

**1. 基于协作的抢占**
```go
// 在函数调用时检查stackguard0
// 编译器在函数入口插入栈检查代码
func morestack() {
    if getg().stackguard0 == stackPreempt {
        // 抢占信号
        gopreempt_m()
    } else {
        // 栈增长
        newstack()
    }
}
```

**2. 基于信号的抢占 (Go 1.14+)**
```go
// preemptone 向M发送抢占信号
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

    // 发送SIGURG信号
    if preemptMSupported && debug.asyncpreemptoff == 0 {
        preemptM(mp)
        return true
    }

    return false
}
```

**sysmon** - 系统监控线程
```go
func sysmon() {
    // sysmon运行在M但不需要P
    // 它是一个死循环，永不返回

    for {
        // 休眠一段时间
        usleep(delay)

        // 1. 检查是否需要强制GC
        if forcegchelper.load() != 0 {
            forcegchelper.set(0)
            sysmon_forcegc()
        }

        // 2. 网络轮询
        lastpoll := sched.lastpoll.Load()
        if netpollinited() && lastpoll != 0 && now-lastpoll > 10*1000*1000 {
            sched.lastpoll.CompareAndSwap(lastpoll, now)
            list, delta := netpoll(0)
            if !list.empty() {
                incidlelocked(-1)
                injectglist(&list)
                incidlelocked(1)
            }
        }

        // 3. 抢占长时间运行的G
        // 检查每个P
        for i := 0; i < len(allp); i++ {
            pp := allp[i]
            if pp == nil {
                continue
            }

            pd := &pp.sysmontick
            s := pp.status

            if s == _Prunning || s == _Psyscall {
                // 检查是否运行过长
                t := now - pd.schedwhen
                if t > forcePreemptNS {
                    preemptone(pp)
                }
            }

            // 4. 从长时间syscall中retake P
            if s == _Psyscall {
                // 如果P在syscall中超过10ms，抢夺它
                t := now - pd.syscallwhen
                if t > 10*1000*1000 {
                    if atomic.Cas(&pp.status, s, _Pidle) {
                        incidlelocked(-1)
                        pp.syscalltick++
                        handoffp(pp)
                        incidlelocked(1)
                    }
                }
            }
        }
    }
}
```

---

## 五、核心调度流程总结

### 5.1 完整调度流程图

```
程序启动
    ↓
runtime·rt0_go
    ↓
schedinit() ──────┐
    ↓              │ 创建P数组 [GOMAXPROCS]
procresize()  ←───┘
    ↓
newproc(runtime.main) ← 创建main goroutine
    ↓
mstart() ──────┐
    ↓           │
mstart1()      │ M的启动
    ↓           │
schedule() ←───┘
    ↓
findRunnable() ──────┐
    ├─ 1. 本地队列    │
    ├─ 2. 全局队列    │ 查找G的顺序
    ├─ 3. netpoll     │
    ├─ 4. 工作窃取    │
    └─ 5. 休眠   ←────┘
    ↓
execute(gp)
    ↓
gogo(&gp.sched) ← 切换到G的栈
    ↓
┌───────────────┐
│  用户代码执行  │
└───────────────┘
    ↓
    ├─→ goexit() → goexit1() → goexit0() → schedule()
    ├─→ gopark() → park_m() → schedule()
    ├─→ Gosched() → gosched_m() → schedule()
    └─→ morestack() → newstack()/gopreempt_m() → schedule()
```

### 5.2 关键函数调用链

**创建goroutine**:
```
go func() → newproc → newproc1 → runqput → wakep
```

**调度循环**:
```
schedule → findRunnable → execute → gogo → 用户代码
```

**主动让出**:
```
runtime.Gosched → mcall → gosched_m → globrunqput → schedule
```

**阻塞等待**:
```
gopark → mcall → park_m → casgstatus(_Gwaiting) → schedule
```

**系统调用**:
```
entersyscall → casgstatus(_Gsyscall) → 执行syscall
exitsyscall → exitsyscallfast / mcall(exitsyscall0)
```

---

## 六、源码阅读建议

### 6.1 阅读顺序

1. **数据结构** (runtime2.go)
   - g, m, p 结构体
   - gobuf, schedt
   - 理解各字段的含义

2. **初始化** (proc.go)
   - schedinit() - 第782行
   - procresize() - 创建P
   - newproc() - 第4754行

3. **调度核心** (proc.go)
   - schedule() - 第3966行
   - findRunnable() - 第3249行
   - execute() - 第3221行

4. **上下文切换** (asm_amd64.s)
   - gogo() - 汇编实现
   - mcall() - 从G切换到g0

5. **系统调用** (proc.go)
   - entersyscall/exitsyscall
   - sysmon() - 系统监控

### 6.2 调试技巧

**1. 使用 GODEBUG**
```bash
# 打印调度器trace
GODEBUG=schedtrace=1000 ./program

# 输出示例
SCHED 1000ms: gomaxprocs=4 idleprocs=2 threads=6 spinningthreads=0 idlethreads=1 runqueue=0 [0 0 0 0]
```

**2. 使用 delve 调试**
```bash
dlv debug main.go
(dlv) b runtime.schedule
(dlv) c
(dlv) print getg().m.p.ptr().runqhead
```

**3. 查看汇编代码**
```bash
go build -gcflags="-S" main.go 2>&1 | less
```

### 6.3 重要源码文件

```
$GOROOT/src/runtime/
├── runtime2.go     # 数据结构定义 (约1000行)
├── proc.go         # 调度器实现 (约7000行) ⭐⭐⭐
├── asm_amd64.s     # 汇编实现 (约1600行)
├── stack.go        # 栈管理
├── mgc.go          # 垃圾回收
├── malloc.go       # 内存分配
├── chan.go         # channel实现
├── select.go       # select实现
├── netpoll*.go     # 网络轮询
└── os_linux.go     # OS相关实现
```

---

## 七、常见问题

### Q1: GOMAXPROCS 应该设置为多少？

**A**: 默认值（CPU核心数）通常是最优的。
- CPU密集型：GOMAXPROCS = CPU核心数
- I/O密集型：可以适当增大
- 容器环境：注意 CPU quota 限制

### Q2: 为什么需要P？

**A**: P的引入解决了几个问题：
1. **减少锁竞争**：每个P有本地队列，无锁访问
2. **缓存局部性**：P持有mcache等本地资源
3. **可扩展性**：M和P的数量可以独立调整

### Q3: M和P的数量关系？

**A**:
- P的数量固定：`GOMAXPROCS`
- M的数量动态：按需创建，最多10000个
- 活跃M数量 ≈ P的数量
- 空闲M会休眠，等待唤醒

### Q4: goroutine什么时候会被抢占？

**A**:
1. **协作式抢占**：函数调用时检查 stackguard0
2. **信号式抢占**：运行超过10ms，sysmon发送SIGURG
3. **系统调用抢占**：syscall超过10ms，P被retake

### Q5: 为什么goroutine如此轻量？

**A**:
1. **小栈**：初始只有2KB，按需增长
2. **用户态调度**：无需系统调用
3. **对象复用**：gFree列表复用g结构体
4. **本地队列**：减少同步开销

---

## 八、总结

### 核心要点

1. **GMP模型** = Goroutine + Machine + Processor
   - G：用户级线程，2KB栈
   - M：OS线程，承载G的执行
   - P：调度上下文，持有runnable队列

2. **两级队列**
   - 本地队列：P的runq[256]，无锁访问
   - 全局队列：sched.runq，需要加锁

3. **调度策略**
   - 优先本地队列
   - 周期性检查全局队列（防止饥饿）
   - 工作窃取（负载均衡）
   - 抢占调度（公平性）

4. **关键函数**
   - `schedule()` - 调度主循环
   - `findRunnable()` - 查找可运行的G
   - `execute()` - 执行G
   - `gogo()` - 上下文切换

5. **系统调用处理**
   - entersyscall：P进入_Psyscall状态
   - exitsyscall：尝试快速获取P
   - sysmon：监控并retake长时间syscall的P

### 进一步学习

- 阅读 proc.go 完整源码
- 理解抢占机制的演进
- 学习netpoll的实现
- 研究GC如何与调度器交互

---

**源码版本**: Go 1.23.0
**文档作者**: AI Assistant
**最后更新**: 2025-12-22
