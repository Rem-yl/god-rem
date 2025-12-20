# Goroutine 生命周期与调度实战分析

基于示例程序：
```go
func main() {
    fmt.Println("Hello, world!")

    go func() {
        fmt.Println("Hello, goroutine world!")
        time.Sleep(5 * time.Second)
    }()

    time.Sleep(6 * time.Second)
}
```

---

## 一、编译时分析

### 1.1 逃逸分析

使用 `-gcflags="-m -m"` 查看编译器的逃逸分析：

```bash
$ go build -gcflags="-m -m" main.go
```

**关键发现**：

1. **闭包逃逸到堆**：
   ```
   ./main.go:11:5: func literal escapes to heap
   ```
   - 原因：`go func()` 创建的闭包会被传递给 `runtime.newproc`
   - 闭包必须在堆上分配，因为它的生命周期超过了创建它的函数

2. **字符串逃逸到堆**：
   ```
   ./main.go:12:15: "Hello, goroutine world!" escapes to heap
   ```
   - 原因：字符串被传递给 `fmt.Println`，需要通过接口传递
   - 接口要求值在堆上

3. **main 函数不能内联**：
   ```
   ./main.go:8:6: cannot inline main: unhandled op GO
   ```
   - 原因：包含 `go` 语句（goroutine 创建）

---

## 二、运行时调度追踪

### 2.1 调度器状态（程序启动时 t=0ms）

```
SCHED 0ms:
  gomaxprocs=24        # P 的数量 = CPU 核心数
  idleprocs=23         # 空闲的 P（只有 P0 在运行 main）
  threads=5            # OS 线程总数
  spinningthreads=0    # 没有线程在自旋寻找工作
  runqueue=0           # 全局运行队列为空
```

**P 的状态**：
```
P0: status=1 schedtick=1 syscalltick=0 m=0 runqsize=0
  - status=1: _Prunning（正在运行）
  - m=0: 关联到 M0（主线程）
  - runqsize=0: 本地队列为空

P1-P23: status=0 m=nil runqsize=0
  - status=0: _Pidle（空闲）
  - m=nil: 未关联到任何 M
```

**M 的状态**：
```
M0: p=0 curg=1 mallocing=0 locks=2 spinning=false
  - p=0: 关联到 P0
  - curg=1: 当前运行 G1（main goroutine）
  - locks=2: 持有 2 个锁
  - spinning=false: 不在自旋状态

M1-M4: blocked=true/false, curg=nil
  - 空闲的 M，等待被唤醒
```

**G 的状态**：
```
G1: status=2(chan receive) m=0 lockedm=0
  - status=2: _Grunning（正在运行 main）
  - 实际上这里显示 "chan receive" 可能是快照时刻

G2: status=4(force gc (idle))
  - GC 强制触发的 goroutine

G3: status=4(GC sweep wait)
  - GC 清扫 goroutine

G4: status=4(GC scavenge wait)
  - 内存回收 goroutine

G5: status=1()
  - status=1: _Grunnable（可运行）
  - 这是我们创建的 goroutine！
```

---

## 三、Goroutine 创建详解

### 3.1 源码级别的创建过程

**Go 语句编译**：
```go
go func() { ... }()
```

**编译器生成**（伪代码）：
```go
fn := &funcval{fn: <function address>}
runtime.newproc(fn)
```

---

### 3.2 runtime.newproc 调用链

```
go func() { ... }
  ↓
runtime.newproc(fn)
  ↓
systemstack(func() {
    newg := newproc1(fn, gp, pc, false, waitReasonZero)
    runqput(pp, newg, true)
    wakep()
})
```

**关键步骤**：

#### 步骤 1：newproc1 - 创建 G

**源码位置**: `runtime/proc.go`

```go
func newproc1(fn *funcval, callergp *g, callerpc uintptr, parked bool, waitreason waitReason) *g {
    mp := acquirem()  // 禁止抢占，防止被调度走
    pp := mp.p.ptr()

    // 1. 尝试从 P 的本地缓存获取 dead G
    newg := gfget(pp)
    if newg == nil {
        // 2. 分配新的 G（初始栈大小 2KB）
        newg = malg(_StackMin)  // _StackMin = 2048
        casgstatus(newg, _Gidle, _Gdead)
        allgadd(newg)  // 加入全局 G 列表（用于 GC 扫描）
    }

    // 3. 计算栈空间
    totalSize := uintptr(4*goarch.PtrSize + sys.MinFrameSize)
    totalSize = alignUp(totalSize, sys.StackAlign)
    sp := newg.stack.hi - totalSize  // 栈向下增长

    // 4. 清空 sched 并设置初始值
    memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
    newg.sched.sp = sp
    newg.sched.pc = abi.FuncPCABI0(goexit) + sys.PCQuantum  // 返回地址设为 goexit
    newg.sched.g = guintptr(unsafe.Pointer(newg))

    // 5. 设置函数指针
    gostartcallfn(&newg.sched, fn)

    // 6. 初始化 G 的其他字段
    newg.gopc = callerpc      // 创建者的 PC
    newg.startpc = fn.fn      // G 的入口函数
    newg.parentGoid = callergp.goid

    // 7. 状态转换：_Gdead → _Grunnable
    casgstatus(newg, _Gdead, _Grunnable)

    // 8. 分配 goid
    if pp.goidcache == pp.goidcacheend {
        // 从全局获取一批 goid
        pp.goidcache = sched.goidgen.Add(_GoidCacheBatch)
        pp.goidcacheend = pp.goidcache + _GoidCacheBatch
    }
    newg.goid = pp.goidcache
    pp.goidcache++

    releasem(mp)
    return newg
}
```

**栈布局**：
```
高地址 (stack.hi)
  ┌────────────────┐
  │  未使用空间      │
  ├────────────────┤
  │  返回地址(goexit)│ ← sched.pc
  ├────────────────┤
  │  函数参数        │
  ├────────────────┤
  │  函数地址        │ ← sched.sp (栈指针)
  └────────────────┘
低地址 (stack.lo)
```

#### 步骤 2：gostartcallfn - 设置函数调用

```go
func gostartcallfn(gobuf *gobuf, fv *funcval) {
    var fn unsafe.Pointer
    if fv != nil {
        fn = unsafe.Pointer(fv.fn)
    } else {
        fn = unsafe.Pointer(abi.FuncPCABIInternal(nilfunc))
    }
    gostartcall(gobuf, fn, unsafe.Pointer(fv))
}

func gostartcall(buf *gobuf, fn, ctxt unsafe.Pointer) {
    sp := buf.sp
    sp -= goarch.PtrSize
    *(*uintptr)(unsafe.Pointer(sp)) = buf.pc  // 保存返回地址
    buf.sp = sp
    buf.pc = uintptr(fn)   // 设置 PC 为函数入口
    buf.ctxt = ctxt
}
```

#### 步骤 3：runqput - 将 G 放入运行队列

**源码位置**: `runtime/proc.go`

```go
func runqput(pp *p, gp *g, next bool) {
    // 1. 如果 next=true，尝试放入 runnext
    if next {
    retryNext:
        oldnext := pp.runnext
        if !pp.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {
            goto retryNext
        }
        if oldnext == 0 {
            return  // runnext 之前为空，直接返回
        }
        // runnext 之前有值，把旧值放入本地队列
        gp = oldnext.ptr()
    }

retry:
    h := atomic.LoadAcq(&pp.runqhead)
    t := pp.runqtail

    // 2. 本地队列未满，放入队列尾部
    if t-h < uint32(len(pp.runq)) {
        pp.runq[t%uint32(len(pp.runq))].set(gp)
        atomic.StoreRel(&pp.runqtail, t+1)
        return
    }

    // 3. 本地队列满了，转移一半到全局队列
    if runqputslow(pp, gp, h, t) {
        return
    }
    goto retry
}
```

**队列优先级**：
```
1. runnext (1个槽位，优先级最高)
   ↓
2. 本地队列 runq (256个槽位)
   ↓
3. 全局队列 sched.runq (无限大小，需要加锁)
```

#### 步骤 4：wakep - 唤醒或创建 M

```go
func wakep() {
    // 1. 如果没有空闲的 P，直接返回
    if sched.npidle.Load() == 0 {
        return
    }

    // 2. 如果已经有 M 在自旋寻找工作，返回
    if sched.nmspinning.Load() != 0 || !sched.nmspinning.CompareAndSwap(0, 1) {
        return
    }

    // 3. 启动 M
    startm(nil, true, false)
}

func startm(pp *p, spinning, lockedg bool) {
    lock(&sched.lock)

    // 1. 如果没有指定 P，从空闲列表获取
    if pp == nil {
        pp = pidleget()
        if pp == nil {
            unlock(&sched.lock)
            if spinning {
                sched.nmspinning.Add(-1)
            }
            return
        }
    }

    // 2. 获取空闲的 M
    nmp := mget()
    if nmp == nil {
        // 没有空闲 M，创建新的
        id := mReserveID()
        unlock(&sched.lock)

        var fn func()
        if spinning {
            fn = mspinning
        }
        newm(fn, pp, id)  // 创建新的 M
        return
    }

    unlock(&sched.lock)

    // 3. 唤醒 M
    nmp.nextp.set(pp)
    nmp.park.note.signal()
}
```

---

## 四、Goroutine 调度执行

### 4.1 M 的调度循环

**入口**: `runtime.mstart` → `runtime.mstart1` → `schedule`

```go
func schedule() {
    mp := getg().m
    pp := mp.p.ptr()

top:
    // 1. GC 等待检查
    if sched.gcwaiting.Load() {
        gcstopm()
        goto top
    }

    // 2. 检查 timer
    checkTimers(pp, 0)

    var gp *g
    var inheritTime bool

    // 3. 每 61 次调度从全局队列获取（公平性保证）
    if pp.schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)
        gp = globrunqget(pp, 1)
        unlock(&sched.lock)
    }

    // 4. 从本地队列获取
    if gp == nil {
        gp, inheritTime = runqget(pp)
    }

    // 5. 如果本地队列为空，寻找可运行的 G
    if gp == nil {
        gp, inheritTime = findrunnable()  // 阻塞直到找到
    }

    // 6. 执行 G
    execute(gp, inheritTime)
}
```

### 4.2 runqget - 从本地队列获取

```go
func runqget(pp *p) (gp *g, inheritTime bool) {
    // 1. 优先检查 runnext
    next := pp.runnext
    if next != 0 && pp.runnext.cas(next, 0) {
        return next.ptr(), true
    }

    // 2. 从本地队列头部获取
    for {
        h := atomic.LoadAcq(&pp.runqhead)
        t := atomic.LoadAcq(&pp.runqtail)
        if t == h {
            return nil, false  // 队列为空
        }
        gp := pp.runq[h%uint32(len(pp.runq))].ptr()
        if atomic.CasRel(&pp.runqhead, h, h+1) {
            return gp, false
        }
    }
}
```

### 4.3 execute - 执行 Goroutine

```go
func execute(gp *g, inheritTime bool) {
    mp := getg().m

    // 1. 状态转换：_Grunnable → _Grunning
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

    // 3. 切换到 G 的上下文执行
    gogo(&gp.sched)  // 汇编实现，不会返回
}
```

### 4.4 gogo - 上下文切换（汇编）

**源码位置**: `runtime/asm_amd64.s`

```asm
TEXT runtime·gogo(SB), NOSPLIT, $0-8
    MOVQ    buf+0(FP), BX       // gobuf
    MOVQ    gobuf_g(BX), DX     // G

    // 更新 TLS 中的 g
    get_tls(CX)
    MOVQ    DX, g(CX)
    MOVQ    DX, R14             // R14 = g (调用约定)

    // 恢复寄存器
    MOVQ    gobuf_sp(BX), SP    // 恢复栈指针
    MOVQ    gobuf_ret(BX), AX
    MOVQ    gobuf_ctxt(BX), DX
    MOVQ    gobuf_bp(BX), BP

    // 清空 gobuf
    MOVQ    $0, gobuf_sp(BX)
    MOVQ    $0, gobuf_ret(BX)
    MOVQ    $0, gobuf_ctxt(BX)
    MOVQ    $0, gobuf_bp(BX)

    // 跳转到 PC
    MOVQ    gobuf_pc(BX), BX
    JMP     BX                  // 开始执行 G 的代码
```

---

## 五、Goroutine 退出

### 5.1 正常退出流程

当 goroutine 的函数返回时：

```
goroutine 函数 return
  ↓
goexit1()  (在栈上预设的返回地址)
  ↓
mcall(goexit0)
  ↓
goexit0() - 清理并回到调度
```

**goexit0 实现**：
```go
func goexit0(gp *g) {
    mp := getg().m
    pp := mp.p.ptr()

    // 1. 状态转换：_Grunning → _Gdead
    casgstatus(gp, _Grunning, _Gdead)

    // 2. 清理 G 的数据
    gp.m = nil
    gp.lockedm = 0
    mp.curg = nil

    // 3. 清空栈指针（GC 扫描优化）
    gp.sched.sp = 0
    gp.sched.pc = 0
    gp.sched.g = 0

    // 4. 放入 P 的本地缓存（用于复用）
    gfput(pp, gp)

    // 5. 继续调度
    schedule()
}
```

---

## 六、系统调用处理

### 6.1 进入系统调用

**示例**: `time.Sleep(5 * time.Second)`

```
time.Sleep()
  ↓
runtime.timeSleep()
  ↓
gopark() - 阻塞当前 G
  ↓
或者是真正的系统调用 (如 read/write)
  ↓
entersyscall()
```

**entersyscall 实现**：
```go
func entersyscall() {
    reentersyscall(getcallerpc(), getcallersp())
}

func reentersyscall(pc, sp uintptr) {
    gp := getg()

    // 1. 保存调用者上下文
    save(pc, sp)
    gp.syscallsp = sp
    gp.syscallpc = pc

    // 2. 状态转换：_Grunning → _Gsyscall
    casgstatus(gp, _Grunning, _Gsyscall)

    // 3. 解除 P 和 M 的关联
    pp := gp.m.p.ptr()
    pp.m = 0
    gp.m.oldp.set(pp)
    gp.m.p = 0
    atomic.Store(&pp.status, _Psyscall)

    // 4. sysmon 会监控长时间的系统调用
    //    如果超过 10ms，会将 P Hand-off 给其他 M
}
```

**P 的状态变化**：
```
_Prunning (执行 Go 代码)
  ↓ entersyscall
_Psyscall (系统调用中，P 暂时闲置)
  ↓ sysmon retake (如果超时)
_Pidle (P 被其他 M 获取)
```

### 6.2 退出系统调用

```go
func exitsyscall() {
    gp := getg()

    // 1. 快速路径：尝试重新获取原来的 P
    oldp := gp.m.oldp.ptr()
    gp.m.oldp = 0

    if exitsyscallfast(oldp) {
        // 成功重新获取 P
        casgstatus(gp, _Gsyscall, _Grunning)
        gp.syscallsp = 0
        gp.m.p.ptr().syscalltick++
        return
    }

    // 2. 慢速路径：无法获取 P
    //    将 G 放回全局队列，M 休眠
    mcall(exitsyscall0)
}

func exitsyscallfast(oldp *p) bool {
    // 1. 尝试 CAS oldp 的状态：_Psyscall → _Prunning
    if oldp != nil && oldp.status == _Psyscall && oldp.m == 0 {
        if atomic.Cas(&oldp.status, _Psyscall, _Prunning) {
            wirep(oldp)
            return true
        }
    }

    // 2. 尝试获取空闲 P
    if sched.pidle != 0 {
        if acquirep(pidleget()) {
            return true
        }
    }

    return false
}
```

---

## 七、实际运行分析

### 7.1 我们的程序执行时间线

```
t=0ms:
  M0-P0 执行 G1 (main goroutine)
  ├─ fmt.Println("Hello, world!")
  ├─ 创建 G5 (go func)
  │  └─ runqput(P0, G5, true) → 放入 P0.runnext
  └─ time.Sleep(6s) → gopark → G1 阻塞

t=1ms:
  M0-P0 从队列获取 G5
  ├─ execute(G5)
  ├─ fmt.Println("Hello, goroutine world!")
  └─ time.Sleep(5s) → gopark → G5 阻塞

t=5001ms:
  G5 的 timer 到期
  └─ G5 状态：_Gwaiting → _Grunnable
  └─ G5 执行完毕 → goexit0 → G5: _Grunning → _Gdead

t=6001ms:
  G1 的 timer 到期
  └─ G1 状态：_Gwaiting → _Grunnable
  └─ G1 执行完毕（main返回）→ runtime.main 调用 exit(0)

程序退出
```

### 7.2 调度器快照（t=1002ms）

```
SCHED 1002ms: gomaxprocs=24 idleprocs=24 threads=5
  P0: status=0 syscalltick=2 runqsize=0 timerslen=2
    - 所有 P 都是 idle
    - P0 有 2 个 timer（G1 和 G5 的 sleep）

  G1: status=4(timer goroutine (idle))
    - 在等待 timer

  G5: (已经执行完毕或在等待)
```

---

## 八、总结

### 8.1 Goroutine 生命周期

```
创建 → 可运行 → 运行 → 阻塞/完成 → 死亡 → 复用
  ↓      ↓       ↓        ↓          ↓      ↓
newproc runqput execute  gopark    goexit  gfput
_Gdead  _Grunnable _Grunning _Gwaiting _Gdead (复用池)
```

### 8.2 关键机制

1. **G 的复用**: Dead G 被放入本地缓存，避免频繁分配
2. **栈管理**: 初始 2KB，动态增长到 1GB
3. **调度公平性**: 每 61 次从全局队列获取
4. **Work Stealing**: 空闲 M 从其他 P 偷取工作
5. **系统调用优化**: Hand-off 机制，P 不被阻塞
