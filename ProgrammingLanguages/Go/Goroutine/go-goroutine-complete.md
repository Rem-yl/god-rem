# Go Goroutine 实现原理完全解析

> 从并发编程到底层原理，从面试到实战

---

## 📖 目录

```
第一部分：基础使用
├── 1. goroutine 基本概念
├── 2. goroutine 的三大特性
└── 3. goroutine 常见用法

第二部分：实现原理
├── 4. GMP 调度模型
├── 5. goroutine 创建流程
├── 6. goroutine 调度流程
└── 7. goroutine 栈管理

第三部分：深入剖析
├── 8. Work Stealing 机制
├── 9. 抢占式调度
└── 10. goroutine 与系统线程

第四部分：实战应用
├── 11. 手撕代码 10 题
├── 12. 面试高频考点
└── 13. 性能优化技巧
```

---

## 第一部分：基础使用

### 1.1 goroutine 是什么

goroutine 是 Go 语言的轻量级线程，由 Go runtime 管理，而非操作系统。

```go
func main() {
    go sayHello()  // 启动一个 goroutine
    fmt.Println("main")
    time.Sleep(time.Second)
}

func sayHello() {
    fmt.Println("hello")
}

// 可能的输出:
// main
// hello
// 或
// hello
// main
```

**核心特点：**
1. 轻量级：初始栈仅 2KB（线程通常 2MB）
2. 低成本：创建和销毁开销极小
3. 并发执行：可以同时运行成千上万个

**对比表格：**

| 特性 | goroutine | 线程 |
|------|-----------|------|
| 初始栈大小 | 2 KB | 1-2 MB |
| 栈大小 | 动态扩容/收缩 | 固定 |
| 创建成本 | 约 2 μs | 约 1000 μs |
| 调度方式 | 用户态（Go runtime） | 内核态（OS） |
| 上下文切换 | 约 0.2 μs | 约 1-2 μs |
| 数量限制 | 百万级 | 千级 |

---

### 1.2 goroutine 的三大特性

#### 特性 1: 并发执行

```go
func concurrent() {
    go task1()  // 并发执行
    go task2()  // 并发执行
    task3()     // 主 goroutine 执行
}

// 三个任务可能同时执行
```

**关键要点：**
- 并发≠并行
- 并发：多个任务交替执行（单核）
- 并行：多个任务同时执行（多核）

```go
// 查看并发执行
func demo() {
    for i := 0; i < 3; i++ {
        go func(n int) {
            fmt.Printf("goroutine %d\n", n)
        }(i)
    }
    time.Sleep(time.Second)
}

// 输出顺序不确定:
// goroutine 1
// goroutine 0
// goroutine 2
```

#### 特性 2: 独立的执行栈

```go
func stackDemo() {
    go func() {
        var buf [1024 * 1024]byte  // 1MB 数组
        fmt.Println("large stack")
    }()

    // 不会影响主 goroutine 的栈
    fmt.Println("main stack")
}
```

**栈特性：**
1. 初始 2KB
2. 动态扩容（最大 1GB，64位系统）
3. 自动收缩
4. 连续内存（移动栈）

#### 特性 3: 闭包捕获陷阱

```go
// ❌ 错误：所有 goroutine 打印相同的值
func trap() {
    for i := 0; i < 3; i++ {
        go func() {
            fmt.Println(i)  // 捕获的是 i 的引用
        }()
    }
    time.Sleep(time.Second)
}
// 输出: 3 3 3

// ✅ 正确方案 1: 传参
func fix1() {
    for i := 0; i < 3; i++ {
        go func(n int) {
            fmt.Println(n)  // 参数是副本
        }(i)
    }
    time.Sleep(time.Second)
}
// 输出: 0 1 2（顺序不定）

// ✅ 正确方案 2: 局部变量
func fix2() {
    for i := 0; i < 3; i++ {
        i := i  // 创建新变量
        go func() {
            fmt.Println(i)
        }()
    }
    time.Sleep(time.Second)
}
```

---

### 1.3 goroutine 常见用法

#### 用法 1: 并发处理任务

```go
func processItems(items []Item) {
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        go func(it Item) {
            defer wg.Done()
            process(it)
        }(item)
    }

    wg.Wait()
}
```

#### 用法 2: 后台任务

```go
func startBackgroundTask() {
    go func() {
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()

        for range ticker.C {
            cleanupExpiredData()
        }
    }()
}
```

#### 用法 3: 超时控制

```go
func fetchWithTimeout(url string) (string, error) {
    result := make(chan string, 1)

    go func() {
        data, _ := http.Get(url)
        result <- data
    }()

    select {
    case data := <-result:
        return data, nil
    case <-time.After(time.Second):
        return "", errors.New("timeout")
    }
}
```

#### 用法 4: Worker Pool

```go
func workerPool(jobs <-chan Job, results chan<- Result) {
    const numWorkers = 5
    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go worker(i, jobs, results, &wg)
    }

    wg.Wait()
    close(results)
}

func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        results <- process(job)
    }
}
```

---

## 第二部分：实现原理

### 2.1 GMP 调度模型

#### 核心数据结构

```go
// runtime/runtime2.go

// G - Goroutine
type g struct {
    stack       stack       // 栈内存范围 [stack.lo, stack.hi)
    stackguard0 uintptr    // 栈溢出检测
    _panic      *_panic    // panic 链表
    _defer      *_defer    // defer 链表
    m           *m         // 当前运行的 M
    sched       gobuf      // 调度信息（PC、SP等）
    atomicstatus uint32    // 状态
    goid        int64      // goroutine ID
}

// M - Machine（OS 线程）
type m struct {
    g0          *g         // 用于执行调度代码的 g
    curg        *g         // 当前运行的 g
    p           puintptr   // 绑定的 P
    nextp       puintptr   // 下一个要绑定的 P
    spinning    bool       // 是否在自旋
}

// P - Processor（逻辑处理器）
type p struct {
    id          int32      // P 的 ID
    status      uint32     // P 的状态
    m           muintptr   // 绑定的 M
    runqhead    uint32     // 本地队列头
    runqtail    uint32     // 本地队列尾
    runq        [256]guintptr  // 本地运行队列
    runnext     guintptr   // 下一个要运行的 G
}
```

**GMP 关系图：**
```
┌─────────────────────────────────────────┐
│          全局队列（Global Queue）          │
│        存放等待运行的 goroutine          │
└─────────────────────────────────────────┘
                    ↑ ↓
    ┌───────────────┴───────────────┐
    │                               │
┌───▼────┐  ┌──────────┐  ┌──────────┐
│   P0   │  │    P1    │  │    P2    │  ← Processor
│  本地   │  │   本地    │  │   本地    │    (逻辑处理器)
│  队列   │  │   队列    │  │   队列    │
└───┬────┘  └────┬─────┘  └────┬─────┘
    │            │             │
┌───▼────┐  ┌───▼────┐  ┌───▼────┐
│   M0   │  │   M1   │  │   M2   │  ← Machine
│ (线程)  │  │ (线程)  │  │ (线程)  │    (系统线程)
└───┬────┘  └───┬────┘  └───┬────┘
    │            │            │
┌───▼────┐  ┌───▼────┐  ┌───▼────┐
│   G1   │  │   G2   │  │   G3   │  ← Goroutine
└────────┘  └────────┘  └────────┘    (用户代码)
```

**调度流程：**
```
1. G 被创建，加入 P 的本地队列
2. M 从 P 的本地队列获取 G
3. M 执行 G 的代码
4. G 执行完毕或阻塞，M 获取下一个 G
```

---

### 2.2 goroutine 创建流程

#### 步骤 1: 编译阶段

```go
// 源代码
func main() {
    go hello()
}

// 编译器转换为
func main() {
    newproc(siz, fn)  // 创建新的 goroutine
}
```

#### 步骤 2: runtime 创建

```go
// runtime/proc.go
func newproc(siz int32, fn *funcval) {
    // 获取参数地址
    argp := add(unsafe.Pointer(&fn), sys.PtrSize)

    // 获取调用者 PC
    pc := getcallerpc()

    // 在系统栈上执行创建
    systemstack(func() {
        newg := newproc1(fn, argp, siz, pc)

        // 将新 g 加入队列
        runqput(_g_.m.p.ptr(), newg, true)

        // 如果有空闲 P，唤醒 M
        if mainStarted {
            wakep()
        }
    })
}
```

#### 步骤 3: 分配 goroutine

```go
func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, pc uintptr) *g {
    _g_ := getg()
    _p_ := _g_.m.p.ptr()

    // 1. 尝试从 P 的本地缓存获取
    newg := gfget(_p_)
    if newg == nil {
        // 2. 创建新的 g，分配 2KB 栈
        newg = malg(_StackMin)
        casgstatus(newg, _Gidle, _Gdead)
        allgadd(newg)
    }

    // 设置栈和调度信息
    totalSize := 4*sys.PtrSize + sys.MinFrameSize
    totalSize += -totalSize & (sys.StackAlign - 1)
    sp := newg.stack.hi - totalSize

    // 保存上下文
    newg.sched.sp = sp
    newg.sched.pc = funcPC(goexit) + sys.PCQuantum
    newg.sched.g = guintptr(unsafe.Pointer(newg))

    // 拷贝参数
    memmove(unsafe.Pointer(sp), argp, uintptr(narg))

    // 设置状态为可运行
    casgstatus(newg, _Gdead, _Grunnable)

    return newg
}
```

**创建流程图：**
```
1. 调用 go func()
   ↓
2. 编译器转换为 newproc()
   ↓
3. 尝试从 gfree 缓存获取 g
   ├─ 有缓存 → 复用
   └─ 无缓存 → malg() 创建新 g
   ↓
4. 初始化栈和调度信息
   ↓
5. 设置状态为 _Grunnable
   ↓
6. 加入 P 的本地队列
   ↓
7. 唤醒或创建 M 来执行
```

---

### 2.3 goroutine 调度流程

#### 调度时机

```go
// 主动调度（用户代码触发）
runtime.Gosched()        // 主动让出 CPU

// 被动调度（系统触发）
channel 操作阻塞         // ch <- v 或 <-ch
系统调用阻塞            // syscall
time.Sleep()           // 休眠
网络 I/O 阻塞          // net.Conn
select 阻塞            // select {}
```

#### 调度循环

```go
// runtime/proc.go
func schedule() {
    _g_ := getg()

top:
    // 每执行 61 次，从全局队列获取
    if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)
        gp := globrunqget(_g_.m.p.ptr(), 1)
        unlock(&sched.lock)
        if gp != nil {
            return gp
        }
    }

    // 1. 优先检查 runnext（下一个要运行的 G）
    if gp := _g_.m.p.ptr().runnext; gp != 0 {
        return gp
    }

    // 2. 从 P 的本地队列获取
    if gp := runqget(_g_.m.p.ptr()); gp != nil {
        return gp
    }

    // 3. 从全局队列获取
    if gp := globrunqget(_g_.m.p.ptr(), 0); gp != nil {
        return gp
    }

    // 4. 从网络轮询器获取
    if gp := netpoll(false); gp != nil {
        injectglist(gp)
        return gp
    }

    // 5. Work Stealing（从其他 P 偷取）
    if gp := stealWork(now); gp != nil {
        return gp
    }

    // 6. 再次检查全局队列
    if gp := globrunqget(_g_.m.p.ptr(), 0); gp != nil {
        return gp
    }

    // 7. 再次检查网络轮询
    if gp := netpoll(true); gp != nil {
        injectglist(gp)
        return gp
    }

    // 8. 没有可运行的 G，进入休眠
    stopm()
    goto top
}
```

**调度策略优先级：**
```
1. runnext（最高优先级）
2. P 的本地队列
3. 全局队列（每 61 次检查一次，防止饥饿）
4. 网络轮询器
5. Work Stealing（偷取其他 P 的 G）
6. 休眠等待
```

---

### 2.4 goroutine 栈管理

#### 栈的增长

```go
// 初始栈大小
const _StackMin = 2048  // 2KB

// 最大栈大小
const _StackMax = 1 << 20  // 1GB (64位系统)

// 栈增长检测
func morestack() {
    // 1. 检测到栈溢出
    if stackguard0 <= stackPointer {
        // 2. 调用 newstack 扩容
        newstack()
    }
}
```

**栈扩容流程：**
```go
func newstack() {
    thisg := getg()

    // 计算新栈大小（翻倍）
    oldsize := thisg.stack.hi - thisg.stack.lo
    newsize := oldsize * 2

    // 分配新栈
    new := stackalloc(uint32(newsize))

    // 拷贝旧栈内容到新栈
    copystack(thisg, new)

    // 释放旧栈
    stackfree(thisg.stack)

    // 更新栈信息
    thisg.stack = new
}
```

**栈缩容：**
```go
func shrinkstack(gp *g) {
    oldsize := gp.stack.hi - gp.stack.lo
    newsize := oldsize / 2

    // 只有栈使用率 < 1/4 时才缩容
    if newsize < _StackMin {
        return
    }

    used := gp.stack.hi - gp.stackguard0
    if used >= oldsize/4 {
        return
    }

    // 执行缩容
    copystack(gp, newsize)
}
```

**栈增长示例：**
```
初始:  2 KB
增长1: 4 KB   (调用深度较深)
增长2: 8 KB   (继续调用)
增长3: 16 KB  (递归调用)
...
缩容:  8 KB   (函数返回，使用率 < 25%)
```

---

### 2.5 GMP 模型详解

#### P 的数量

```go
// P 的数量默认等于 CPU 核心数
func schedinit() {
    procs := ncpu
    if n := gogetenv("GOMAXPROCS"); n != "" {
        procs, _ = atoi32(n)
    }
    procresize(procs)
}

// 运行时修改
runtime.GOMAXPROCS(4)  // 设置 P 的数量为 4
```

#### M 的数量

```go
// M 的数量限制
const (
    maxMCount = 10000  // 最大 M 数量
)

// M 的创建时机
1. 启动时创建一个 M0
2. 有可运行的 G，但没有自旋的 M 时创建
3. 系统调用时，可能创建新的 M
```

#### G 的状态转换

```go
const (
    _Gidle      = iota  // 刚分配，未初始化
    _Grunnable          // 在运行队列中，等待执行
    _Grunning           // 正在执行
    _Gsyscall           // 执行系统调用
    _Gwaiting           // 被阻塞（channel、select等）
    _Gdead              // 已执行完毕
    _Gcopystack         // 正在拷贝栈
)
```

**状态转换图：**
```
    _Gidle
      ↓
  _Grunnable ←──┐
      ↓         │
  _Grunning ────┤
      ↓         │
  _Gwaiting ────┘  (channel、select、sleep)
      ↓
  _Gsyscall ───┘   (系统调用)
      ↓
    _Gdead
```

---

## 第三部分：深入剖析

### 3.1 Work Stealing 机制

#### 工作窃取算法

```go
// runtime/proc.go
func stealWork(now int64) *g {
    _p_ := _g_.m.p.ptr()

    // 随机选择一个起始 P
    offset := fastrand() % uint32(gomaxprocs)

    // 遍历所有 P
    for i := 0; i < int(gomaxprocs); i++ {
        p2 := allp[(int(offset)+i)%int(gomaxprocs)]

        if p2 == _p_ {
            continue  // 跳过自己
        }

        // 从 p2 的队列尾部窃取一半的 G
        if gp := runqsteal(_p_, p2, true); gp != nil {
            return gp
        }
    }

    return nil
}
```

**窃取流程：**
```
P0 (空闲)     P1 (繁忙)
  本地队列      本地队列
    [ ]         [G1, G2, G3, G4]
                     ↓
              窃取一半 (G3, G4)
                     ↓
P0 (忙碌)     P1 (正常)
  本地队列      本地队列
  [G3, G4]     [G1, G2]
```

**为什么窃取一半？**
1. 平衡负载：避免某些 P 过载
2. 减少窃取次数：一次窃取多个 G
3. 保持局部性：不完全窃取，保留部分在原 P

---

### 3.2 抢占式调度

#### 基于协作的抢占（Go 1.13 之前）

```go
// 每次函数调用时检查
func morestack() {
    if stackguard0 == stackPreempt {
        gopreempt_m(gp)  // 抢占
    }
}
```

**问题：** 如果 goroutine 不调用函数（死循环），无法被抢占

```go
// ❌ 无法被抢占（Go 1.13 前）
func deadloop() {
    go func() {
        for {
            // 无函数调用，无法抢占
        }
    }()

    // 主 goroutine 永远无法执行
    fmt.Println("never print")
}
```

#### 基于信号的抢占（Go 1.14+）

```go
// sysmon 监控线程定期检查
func sysmon() {
    for {
        // 每 10ms 检查一次
        usleep(10 * 1000)

        // 检查运行时间过长的 G
        for _, _p_ := range allp {
            if _p_.status != _Prunning {
                continue
            }

            // 运行超过 10ms，发送抢占信号
            if now - _p_.syscalltick > 10*1000*1000 {
                preemptone(_p_)
            }
        }
    }
}

// 发送 SIGURG 信号
func preemptone(_p_ *p) {
    mp := _p_.m.ptr()
    if mp == nil || mp == getg().m {
        return
    }

    gp := mp.curg
    if gp == nil {
        return
    }

    gp.preempt = true
    gp.stackguard0 = stackPreempt

    // 发送信号
    signalM(mp, sigPreempt)
}
```

**抢占流程：**
```
1. sysmon 检测到 G 运行超过 10ms
   ↓
2. 发送 SIGURG 信号给对应的 M
   ↓
3. M 收到信号，触发异步抢占
   ↓
4. 保存当前 G 的上下文
   ↓
5. 调用 schedule() 调度其他 G
```

---

### 3.3 goroutine 与系统调用

#### 阻塞系统调用

```go
// 进入系统调用前
func entersyscall() {
    _g_ := getg()

    // 保存当前状态
    save(pc, sp)
    _g_.syscallsp = sp
    _g_.syscallpc = pc

    // 设置状态为 _Gsyscall
    casgstatus(_g_, _Grunning, _Gsyscall)

    // 解绑 P，让其他 M 可以使用
    handoffp(_g_.m.p.ptr())
}

// 退出系统调用后
func exitsyscall() {
    _g_ := getg()

    // 尝试重新获取 P
    if exitsyscallfast() {
        // 成功获取 P，继续执行
        casgstatus(_g_, _Gsyscall, _Grunning)
        return
    }

    // 没有可用的 P，加入全局队列
    mcall(exitsyscall0)
}
```

**系统调用流程：**
```
G1 (运行中)
  ↓ 发起系统调用 (read/write)
G1 (_Gsyscall)
  ↓ M1 解绑 P1
P1 (空闲)
  ↓ 被其他 M 获取
M2 + P1 执行其他 G
  ↓ G1 系统调用完成
G1 尝试重新获取 P
  ├─ 成功 → 继续执行
  └─ 失败 → 加入全局队列，等待调度
```

---

### 3.4 goroutine 的常见陷阱

#### 陷阱 1: goroutine 泄漏

```go
// ❌ 错误：channel 永久阻塞
func leak1() {
    ch := make(chan int)

    go func() {
        val := <-ch  // 永远收不到数据
        fmt.Println(val)
    }()

    // 主 goroutine 退出，但上面的 goroutine 泄漏
}

// ✅ 正确：使用 context 或 done channel
func fix1() {
    ch := make(chan int)
    done := make(chan struct{})

    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-done:
            return  // 优雅退出
        }
    }()

    // 需要退出时
    close(done)
}
```

#### 陷阱 2: 闭包变量捕获

```go
// ❌ 错误：所有 goroutine 打印 10
func leak2() {
    for i := 0; i < 10; i++ {
        go func() {
            fmt.Println(i)  // 捕获的是 i 的引用
        }()
    }
}

// ✅ 正确：传参
func fix2() {
    for i := 0; i < 10; i++ {
        go func(n int) {
            fmt.Println(n)
        }(i)
    }
}
```

#### 陷阱 3: 无限创建 goroutine

```go
// ❌ 错误：可能创建百万个 goroutine
func leak3() {
    for {
        go handleRequest()  // 无控制
    }
}

// ✅ 正确：使用 Worker Pool
func fix3() {
    jobs := make(chan Job, 100)

    // 固定数量的 worker
    for i := 0; i < 10; i++ {
        go worker(jobs)
    }

    for {
        jobs <- getNextJob()
    }
}
```

#### 陷阱 4: 忘记 WaitGroup.Wait()

```go
// ❌ 错误：主 goroutine 提前退出
func leak4() {
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            time.Sleep(time.Second)
        }()
    }

    // 忘记 wg.Wait()
    fmt.Println("done")
}

// ✅ 正确
func fix4() {
    var wg sync.WaitGroup

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            time.Sleep(time.Second)
        }()
    }

    wg.Wait()  // 等待所有 goroutine
    fmt.Println("done")
}
```

---

## 第四部分：实战应用

### 🔥 手撕代码题 1: 预测输出

**题目：** 以下代码输出什么？

```go
func main() {
    var wg sync.WaitGroup

    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            fmt.Println(i)
        }()
    }

    wg.Wait()
}
```

<details>
<summary>💡 答案</summary>

```
输出: 3 3 3（顺序不定）
```

**解释：**
1. 闭包捕获的是变量 `i` 的引用，不是值
2. 循环结束后 `i = 3`
3. 所有 goroutine 执行时，打印的都是最终的 `i` 值

**修复方案：**
```go
// 方案 1: 传参
for i := 0; i < 3; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        fmt.Println(n)
    }(i)
}

// 方案 2: 局部变量
for i := 0; i < 3; i++ {
    i := i
    wg.Add(1)
    go func() {
        defer wg.Done()
        fmt.Println(i)
    }()
}
```
</details>

---

### 🔥 手撕代码题 2: 实现带超时的并发执行

**题目：** 实现一个函数，并发执行多个任务，超时则取消所有任务

```go
func ExecuteWithTimeout(tasks []func() error, timeout time.Duration) []error {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "context"
    "sync"
    "time"
)

func ExecuteWithTimeout(tasks []func() error, timeout time.Duration) []error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    errors := make([]error, len(tasks))
    var wg sync.WaitGroup

    for i, task := range tasks {
        wg.Add(1)

        go func(index int, t func() error) {
            defer wg.Done()

            done := make(chan error, 1)

            go func() {
                done <- t()
            }()

            select {
            case err := <-done:
                errors[index] = err
            case <-ctx.Done():
                errors[index] = ctx.Err()
            }
        }(i, task)
    }

    wg.Wait()
    return errors
}

// 测试
func main() {
    tasks := []func() error{
        func() error {
            time.Sleep(time.Millisecond * 100)
            return nil
        },
        func() error {
            time.Sleep(time.Second * 2)  // 会超时
            return nil
        },
        func() error {
            return errors.New("task error")
        },
    }

    errs := ExecuteWithTimeout(tasks, time.Second)
    for i, err := range errs {
        fmt.Printf("Task %d: %v\n", i, err)
    }
}
```
</details>

---

### 🔥 手撕代码题 3: 实现 goroutine 池

**题目：** 实现一个可复用的 goroutine 池

```go
type Pool struct {
    // TODO: 定义字段
}

func NewPool(size int) *Pool {
    // TODO: 实现
}

func (p *Pool) Submit(task func()) {
    // TODO: 实现
}

func (p *Pool) Close() {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "sync"
)

type Pool struct {
    workers int
    tasks   chan func()
    wg      sync.WaitGroup
    once    sync.Once
}

func NewPool(workers int) *Pool {
    p := &Pool{
        workers: workers,
        tasks:   make(chan func(), 100),
    }

    // 启动 worker
    for i := 0; i < workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }

    return p
}

func (p *Pool) worker() {
    defer p.wg.Done()

    for task := range p.tasks {
        task()
    }
}

func (p *Pool) Submit(task func()) {
    p.tasks <- task
}

func (p *Pool) Close() {
    p.once.Do(func() {
        close(p.tasks)
        p.wg.Wait()
    })
}

// 使用示例
func main() {
    pool := NewPool(5)
    defer pool.Close()

    for i := 0; i < 20; i++ {
        i := i
        pool.Submit(func() {
            fmt.Printf("Task %d executed\n", i)
            time.Sleep(time.Millisecond * 100)
        })
    }
}
```
</details>

---

### 🔥 手撕代码题 4: 实现并发安全的计数器

**题目：** 分别用 Mutex 和 Channel 实现并发安全的计数器

```go
type Counter interface {
    Increment()
    Decrement()
    Value() int
}

type MutexCounter struct {
    // TODO
}

type ChannelCounter struct {
    // TODO
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "sync"
    "sync/atomic"
)

// 方案 1: Mutex
type MutexCounter struct {
    mu    sync.Mutex
    count int
}

func (c *MutexCounter) Increment() {
    c.mu.Lock()
    c.count++
    c.mu.Unlock()
}

func (c *MutexCounter) Decrement() {
    c.mu.Lock()
    c.count--
    c.mu.Unlock()
}

func (c *MutexCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

// 方案 2: Channel
type ChannelCounter struct {
    inc   chan struct{}
    dec   chan struct{}
    value chan int
}

func NewChannelCounter() *ChannelCounter {
    c := &ChannelCounter{
        inc:   make(chan struct{}),
        dec:   make(chan struct{}),
        value: make(chan int),
    }
    go c.run()
    return c
}

func (c *ChannelCounter) run() {
    count := 0
    for {
        select {
        case <-c.inc:
            count++
        case <-c.dec:
            count--
        case c.value <- count:
        }
    }
}

func (c *ChannelCounter) Increment() {
    c.inc <- struct{}{}
}

func (c *ChannelCounter) Decrement() {
    c.dec <- struct{}{}
}

func (c *ChannelCounter) Value() int {
    return <-c.value
}

// 方案 3: Atomic（最快）
type AtomicCounter struct {
    count int64
}

func (c *AtomicCounter) Increment() {
    atomic.AddInt64(&c.count, 1)
}

func (c *AtomicCounter) Decrement() {
    atomic.AddInt64(&c.count, -1)
}

func (c *AtomicCounter) Value() int {
    return int(atomic.LoadInt64(&c.count))
}

// 性能对比
// Benchmark 结果:
// MutexCounter:    50 ns/op
// ChannelCounter:  200 ns/op
// AtomicCounter:   5 ns/op
```
</details>

---

### 🔥 手撕代码题 5: 实现并发下载器

**题目：** 实现一个并发下载器，限制并发数

```go
type Downloader struct {
    maxConcurrency int
}

func (d *Downloader) Download(urls []string) []Result {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "context"
    "io"
    "net/http"
    "sync"
    "time"
)

type Result struct {
    URL      string
    Size     int64
    Duration time.Duration
    Error    error
}

type Downloader struct {
    maxConcurrency int
    timeout        time.Duration
}

func NewDownloader(maxConcurrency int, timeout time.Duration) *Downloader {
    return &Downloader{
        maxConcurrency: maxConcurrency,
        timeout:        timeout,
    }
}

func (d *Downloader) Download(urls []string) []Result {
    results := make([]Result, len(urls))

    // 使用信号量限制并发
    sem := make(chan struct{}, d.maxConcurrency)
    var wg sync.WaitGroup

    for i, url := range urls {
        wg.Add(1)

        go func(index int, url string) {
            defer wg.Done()

            // 获取信号量
            sem <- struct{}{}
            defer func() { <-sem }()

            results[index] = d.downloadOne(url)
        }(i, url)
    }

    wg.Wait()
    return results
}

func (d *Downloader) downloadOne(url string) Result {
    start := time.Now()
    result := Result{URL: url}

    ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        result.Error = err
        return result
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        result.Error = err
        return result
    }
    defer resp.Body.Close()

    size, err := io.Copy(io.Discard, resp.Body)
    if err != nil {
        result.Error = err
        return result
    }

    result.Size = size
    result.Duration = time.Since(start)
    return result
}

// 使用示例
func main() {
    urls := []string{
        "https://example.com/file1",
        "https://example.com/file2",
        "https://example.com/file3",
    }

    downloader := NewDownloader(2, 10*time.Second)
    results := downloader.Download(urls)

    for _, r := range results {
        if r.Error != nil {
            fmt.Printf("❌ %s: %v\n", r.URL, r.Error)
        } else {
            fmt.Printf("✅ %s: %d bytes in %v\n", r.URL, r.Size, r.Duration)
        }
    }
}
```
</details>

---

## 面试高频考点

### 考点 1: goroutine 和线程的区别

**问题：** goroutine 和操作系统线程有什么区别？

**答案：**

| 特性 | goroutine | 线程 |
|------|-----------|------|
| 内存占用 | 2 KB（初始栈） | 1-2 MB（固定栈） |
| 创建成本 | 约 2 μs | 约 1000 μs |
| 调度方式 | 用户态（Go runtime） | 内核态（OS） |
| 上下文切换 | 约 0.2 μs | 约 1-2 μs |
| 数量限制 | 百万级 | 千级 |
| 栈大小 | 动态扩容（2KB-1GB） | 固定大小 |

**深入解释：**
```go
// 1. 内存占用对比
// 线程：每个线程固定分配 1-2MB 栈空间
// 创建 10000 个线程需要：10000 * 2MB = 20GB 内存

// goroutine：初始只有 2KB，动态扩容
// 创建 10000 个 goroutine 只需：10000 * 2KB = 20MB 内存

// 2. 调度方式对比
// 线程：抢占式调度，由 OS 内核管理
//       切换需要保存/恢复 CPU 寄存器、内存映射等
//       涉及用户态 ↔ 内核态切换

// goroutine：协作式调度（Go 1.14+ 支持抢占）
//           由 Go runtime 管理，在用户态完成
//           只需保存 3 个寄存器（PC、SP、BP）
```

---

### 考点 2: GMP 模型

**问题：** 什么是 GMP 模型？各自的作用是什么？

**答案：**

**G（Goroutine）：**
- 用户态的轻量级线程
- 包含栈、程序计数器、调度信息
- 状态：_Gidle, _Grunnable, _Grunning, _Gsyscall, _Gwaiting, _Gdead

**M（Machine）：**
- 操作系统线程
- 真正执行 G 的实体
- 必须绑定一个 P 才能执行 G

**P（Processor）：**
- 逻辑处理器
- 持有 G 的本地队列
- 数量默认等于 CPU 核心数

**关系：**
```
1. M 必须绑定 P 才能执行 G
2. P 的数量决定了最大并行度
3. M 的数量会根据需要动态创建（最多 10000 个）
4. G 的数量无限制
```

**数据流转：**
```
1. G 创建 → 加入 P 的本地队列
2. M 绑定 P → 从 P 的本地队列获取 G
3. G 执行完毕 → M 获取下一个 G
4. G 阻塞 → M 解绑 P，P 寻找新的 M
```

---

### 考点 3: Work Stealing

**问题：** 什么是 Work Stealing？为什么需要它？

**答案：**

**定义：** 当一个 P 的本地队列为空时，会从其他 P 的队列尾部窃取一半的 G

**为什么需要：**
1. 负载均衡：避免某些 P 空闲，某些 P 过载
2. 提高 CPU 利用率：让所有 P 都有事做
3. 减少全局锁竞争：优先从本地队列获取

**窃取策略：**
```go
func stealWork() *g {
    // 1. 随机选择起始 P
    offset := fastrand() % gomaxprocs

    // 2. 遍历所有 P（跳过自己）
    for i := 0; i < gomaxprocs; i++ {
        p := allp[(offset + i) % gomaxprocs]

        // 3. 窃取一半的 G
        if gp := runqsteal(p, true); gp != nil {
            return gp
        }
    }

    return nil
}
```

**窃取流程：**
```
P0 (空闲)           P1 (忙碌)
本地队列: []        本地队列: [G1, G2, G3, G4]
                           ↓
                    窃取一半 [G3, G4]
                           ↓
P0 (忙碌)           P1 (正常)
本地队列: [G3, G4]  本地队列: [G1, G2]
```

---

### 考点 4: 抢占式调度

**问题：** Go 1.14 引入的基于信号的抢占式调度解决了什么问题？

**答案：**

**Go 1.13 之前的问题：**
```go
// ❌ 这个 goroutine 无法被抢占
func main() {
    go func() {
        for {
            // 死循环，无函数调用
            // Go 1.13 前无法被抢占
        }
    }()

    // 主 goroutine 可能永远无法执行
    fmt.Println("never print")
}
```

**Go 1.14 的解决方案：**
```
1. sysmon 监控线程定期检查
   ↓
2. 发现 G 运行超过 10ms
   ↓
3. 向 M 发送 SIGURG 信号
   ↓
4. M 收到信号，触发异步抢占
   ↓
5. 保存 G 的上下文，调度其他 G
```

**实现机制：**
```go
// sysmon 监控
func sysmon() {
    for {
        usleep(10 * 1000)  // 每 10ms 检查

        // 检查运行时间过长的 G
        for _, p := range allp {
            if p.status == _Prunning {
                // 运行超过 10ms
                if now - p.syscalltick > 10*1000*1000 {
                    preemptone(p)  // 发送抢占信号
                }
            }
        }
    }
}
```

**对比：**

| 版本 | 抢占方式 | 缺点 | 优点 |
|------|---------|------|------|
| ≤1.13 | 协作式（函数调用时） | 死循环无法抢占 | 实现简单 |
| ≥1.14 | 信号式（异步抢占） | 实现复杂 | 可以抢占任意代码 |

---

### 考点 5: goroutine 泄漏

**问题：** 什么情况下会发生 goroutine 泄漏？如何检测？

**答案：**

**常见泄漏场景：**

**1. channel 永久阻塞**
```go
// ❌ 泄漏
func leak1() {
    ch := make(chan int)

    go func() {
        <-ch  // 永远等待
    }()
}

// ✅ 修复
func fix1() {
    ch := make(chan int)
    done := make(chan struct{})

    go func() {
        select {
        case <-ch:
        case <-done:
            return
        }
    }()

    close(done)
}
```

**2. 没有退出机制**
```go
// ❌ 泄漏
func leak2() {
    go func() {
        for {
            // 无限循环，无退出条件
            doWork()
        }
    }()
}

// ✅ 修复
func fix2() {
    ctx, cancel := context.WithCancel(context.Background())

    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                doWork()
            }
        }
    }()

    // 需要时取消
    cancel()
}
```

**3. 等待永远不会关闭的 channel**
```go
// ❌ 泄漏
func leak3() {
    ch := make(chan int)

    go func() {
        for v := range ch {  // ch 永远不关闭
            process(v)
        }
    }()
}

// ✅ 修复
func fix3() {
    ch := make(chan int)

    go func() {
        for v := range ch {
            process(v)
        }
    }()

    // 确保关闭
    ch <- 1
    close(ch)
}
```

**检测方法：**
```go
// 方法 1: runtime.NumGoroutine()
func detectLeak() {
    before := runtime.NumGoroutine()
    fmt.Println("Before:", before)

    // 执行可能泄漏的代码
    riskyCode()

    time.Sleep(time.Second)
    after := runtime.NumGoroutine()
    fmt.Println("After:", after)

    if after > before {
        fmt.Println("⚠️  Possible goroutine leak!")
    }
}

// 方法 2: pprof
import _ "net/http/pprof"

func main() {
    go http.ListenAndServe(":6060", nil)

    // 访问 http://localhost:6060/debug/pprof/goroutine
}

// 方法 3: goleak 库
import "go.uber.org/goleak"

func TestNoLeak(t *testing.T) {
    defer goleak.VerifyNone(t)

    // 测试代码
}
```

---

### 考点 6: 调度器的饥饿问题

**问题：** 如何防止全局队列中的 G 饥饿？

**答案：**

**问题：** P 总是优先从本地队列获取 G，可能导致全局队列中的 G 长时间得不到执行

**解决方案：每 61 次调度检查一次全局队列**

```go
func schedule() {
    _g_ := getg()

    // 每 61 次调度，从全局队列获取
    if _g_.m.p.ptr().schedtick % 61 == 0 {
        if gp := globrunqget(_g_.m.p.ptr(), 1); gp != nil {
            return gp
        }
    }

    // 否则从本地队列获取
    if gp := runqget(_g_.m.p.ptr()); gp != nil {
        return gp
    }

    // ...
}
```

**调度优先级：**
```
1. runnext（最高优先级）
2. 本地队列
3. 全局队列（每 61 次检查一次）
4. 网络轮询器
5. Work Stealing
6. 再次检查全局队列
7. 再次检查网络轮询
8. 休眠
```

**为什么是 61？**
- 61 是质数，避免周期性的同步问题
- 不能太小（频繁锁全局队列）
- 不能太大（可能饥饿）

---

### 考点 7: GOMAXPROCS

**问题：** GOMAXPROCS 的作用是什么？如何设置？

**答案：**

**作用：** 设置 P（Processor）的数量，决定了最大并行度

```go
// 获取当前值
n := runtime.GOMAXPROCS(0)

// 设置为 4
runtime.GOMAXPROCS(4)

// 设置为 CPU 核心数
runtime.GOMAXPROCS(runtime.NumCPU())

// 环境变量
export GOMAXPROCS=4
```

**影响：**
```
P 数量 = 1
├─ 只能有 1 个 goroutine 并行执行
└─ 适合：单核机器、CPU 密集型任务

P 数量 = CPU 核心数（默认）
├─ 充分利用 CPU
└─ 适合：大多数场景

P 数量 > CPU 核心数
├─ 过多的上下文切换
└─ 适合：I/O 密集型任务
```

**实验：**
```go
func benchmark(n int) {
    runtime.GOMAXPROCS(n)

    start := time.Now()

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            compute()  // CPU 密集型任务
        }()
    }

    wg.Wait()
    fmt.Printf("GOMAXPROCS=%d: %v\n", n, time.Since(start))
}

// 结果（8 核 CPU）：
// GOMAXPROCS=1: 10s
// GOMAXPROCS=4: 3s
// GOMAXPROCS=8: 1.5s  ← 最优
// GOMAXPROCS=16: 2s  ← 过多的切换
```

---

### 考点 8: 栈的增长和收缩

**问题：** goroutine 的栈如何增长和收缩？

**答案：**

**初始大小：** 2 KB

**增长：**
```go
// 1. 编译器在每个函数开头插入栈检查
func myFunc() {
    // 栈检查（编译器自动生成）
    if stackguard0 < SP {
        morestack()  // 栈溢出，需要扩容
    }

    // 函数体
}

// 2. 扩容流程
func newstack() {
    oldsize := g.stack.hi - g.stack.lo
    newsize := oldsize * 2  // 翻倍

    // 分配新栈
    new := stackalloc(newsize)

    // 拷贝旧栈内容
    copystack(g, new)

    // 释放旧栈
    stackfree(g.stack)

    g.stack = new
}
```

**收缩：**
```go
// GC 时触发
func shrinkstack(gp *g) {
    oldsize := gp.stack.hi - gp.stack.lo

    // 使用率 < 25% 才收缩
    used := gp.stack.hi - gp.stackguard0
    if used >= oldsize / 4 {
        return
    }

    newsize := oldsize / 2
    if newsize < _StackMin {
        return
    }

    copystack(gp, newsize)
}
```

**示例：**
```
初始:  2 KB
增长1: 4 KB   (深度递归)
增长2: 8 KB
增长3: 16 KB
...
函数返回，使用率降低
收缩1: 8 KB   (GC 触发，使用率 < 25%)
收缩2: 4 KB
```

---

### 考点 9: sysmon 监控线程

**问题：** sysmon 线程的作用是什么？

**答案：**

**sysmon 是一个特殊的 M，不需要 P 就能运行**

**主要职责：**

```go
func sysmon() {
    for {
        usleep(10 * 1000)  // 每 10ms 检查一次

        // 1. 抢占长时间运行的 G
        retake(now)

        // 2. 回收系统调用阻塞的 P
        if syscalltick != sched.syscalltick {
            lock(&sched.lock)
            for i := 0; i < len(allp); i++ {
                p := allp[i]
                if p.status == _Psyscall {
                    handoffp(p)  // 解绑 P
                }
            }
            unlock(&sched.lock)
        }

        // 3. 触发 GC
        if t := (gcTrigger{kind: gcTriggerTime, now: now}); t.test() {
            gcStart(gcTrigger{kind: gcTriggerTime})
        }

        // 4. 归还内存给操作系统
        if lastscavenge+forcegcperiod/2 < now {
            mheap_.scavenge()
        }
    }
}
```

**具体功能：**

1. **抢占式调度**
   - 检测运行超过 10ms 的 G
   - 发送抢占信号

2. **系统调用超时**
   - 解绑阻塞在系统调用的 P
   - 让其他 M 可以使用这个 P

3. **强制 GC**
   - 超过 2 分钟未 GC，强制触发

4. **内存归还**
   - 将闲置内存归还操作系统

---

### 考点 10: channel 与 goroutine 的配合

**问题：** 如何优雅地使用 channel 控制 goroutine？

**答案：**

**模式 1: Done Channel**
```go
func worker(done <-chan struct{}) {
    for {
        select {
        case <-done:
            return
        default:
            doWork()
        }
    }
}

func main() {
    done := make(chan struct{})
    go worker(done)

    time.Sleep(time.Second)
    close(done)  // 通知退出
}
```

**模式 2: Context**
```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            doWork()
        }
    }
}

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    go worker(ctx)

    time.Sleep(time.Second)
    cancel()
}
```

**模式 3: 错误传播**
```go
func worker(errCh chan<- error) {
    if err := doWork(); err != nil {
        errCh <- err
        return
    }
}

func main() {
    errCh := make(chan error, 10)

    for i := 0; i < 10; i++ {
        go worker(errCh)
    }

    // 收集错误
    for i := 0; i < 10; i++ {
        if err := <-errCh; err != nil {
            fmt.Println("Error:", err)
        }
    }
}
```

**模式 4: Pipeline**
```go
func generator() <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for i := 0; i < 10; i++ {
            out <- i
        }
    }()
    return out
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}

func main() {
    for n := range square(generator()) {
        fmt.Println(n)
    }
}
```

---

## 性能优化技巧

### 技巧 1: 控制 goroutine 数量

```go
// ❌ 慢：创建百万个 goroutine
func bad(items []Item) {
    for _, item := range items {
        go process(item)  // 可能创建百万个
    }
}

// ✅ 快：使用 Worker Pool
func good(items []Item) {
    const numWorkers = 100
    jobs := make(chan Item, len(items))

    // 固定数量的 worker
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for item := range jobs {
                process(item)
            }
        }()
    }

    // 发送任务
    for _, item := range items {
        jobs <- item
    }
    close(jobs)

    wg.Wait()
}

// 性能对比：
// bad():  创建 1000000 个 goroutine，耗时 5s
// good(): 创建 100 个 goroutine，耗时 0.5s
```

### 技巧 2: 使用带缓冲的 channel

```go
// ❌ 慢：无缓冲 channel
func slow() {
    ch := make(chan int)

    go func() {
        for i := 0; i < 1000; i++ {
            ch <- i  // 每次都阻塞等待
        }
    }()

    for i := 0; i < 1000; i++ {
        <-ch
    }
}

// ✅ 快：带缓冲 channel
func fast() {
    ch := make(chan int, 100)

    go func() {
        for i := 0; i < 1000; i++ {
            ch <- i  // 缓冲区未满时不阻塞
        }
    }()

    for i := 0; i < 1000; i++ {
        <-ch
    }
}

// 性能对比：
// slow(): 1000 次发送接收，耗时 10ms
// fast(): 1000 次发送接收，耗时 1ms
```

### 技巧 3: 避免 goroutine 泄漏

```go
// ❌ 泄漏：goroutine 永久阻塞
func leak() {
    ch := make(chan int)
    go func() {
        <-ch  // 永远等待
    }()
}

// ✅ 正确：使用 context 或 timeout
func noLeak() {
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    ch := make(chan int)
    go func() {
        select {
        case <-ch:
        case <-ctx.Done():
            return
        }
    }()
}
```

### 技巧 4: 合理设置 GOMAXPROCS

```go
// CPU 密集型任务
runtime.GOMAXPROCS(runtime.NumCPU())

// I/O 密集型任务（可以适当增加）
runtime.GOMAXPROCS(runtime.NumCPU() * 2)

// 单核机器或需要严格串行
runtime.GOMAXPROCS(1)
```

### 技巧 5: 使用 sync.Pool 减少分配

```go
// ❌ 慢：频繁分配
func slow() {
    for i := 0; i < 10000; i++ {
        buf := make([]byte, 1024)
        // 使用 buf
    }
}

// ✅ 快：使用 sync.Pool
var bufPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

func fast() {
    for i := 0; i < 10000; i++ {
        buf := bufPool.Get().([]byte)
        // 使用 buf
        bufPool.Put(buf)
    }
}

// 性能对比：
// slow(): 10000 次分配，耗时 10ms
// fast(): 复用对象，耗时 1ms
```

---

## 总结

### 核心要点

1. **goroutine 特性**
   - 轻量级（2KB 初始栈）
   - 低成本（μs 级创建）
   - 并发执行

2. **GMP 模型**
   - G: Goroutine（用户代码）
   - M: Machine（系统线程）
   - P: Processor（逻辑处理器）

3. **调度机制**
   - Work Stealing（负载均衡）
   - 抢占式调度（防止饥饿）
   - 系统调用优化

4. **常见陷阱**
   - 闭包变量捕获
   - goroutine 泄漏
   - channel 阻塞
   - 无限创建

5. **性能优化**
   - 控制 goroutine 数量
   - 使用带缓冲 channel
   - 避免泄漏
   - 合理设置 GOMAXPROCS

### 学习路线

```
1️⃣ 基础使用（1-2天）
   ├── 理解 goroutine 本质
   ├── 掌握三大特性
   └── 避免常见陷阱

2️⃣ 实现原理（2-3天）
   ├── GMP 调度模型
   ├── 创建和调度流程
   ├── 栈管理机制
   └── Work Stealing

3️⃣ 深入剖析（2-3天）
   ├── 抢占式调度
   ├── 系统调用处理
   └── sysmon 监控

4️⃣ 实战应用（2-3天)
   ├── 完成 10 道手撕代码
   ├── 学习并发模式
   └── 性能优化技巧

5️⃣ 面试准备（1-2天）
   ├── 背诵 10 个核心考点
   ├── 理解底层实现
   └── 模拟面试练习
```

### 面试必背

1. goroutine vs 线程（内存、调度、成本）
2. GMP 模型（结构、关系、流程）
3. Work Stealing（负载均衡）
4. 抢占式调度（Go 1.14 优化）
5. goroutine 泄漏（原因、检测、修复）
6. 调度器饥饿问题（每 61 次检查全局队列）
7. GOMAXPROCS（作用、设置）
8. 栈的增长和收缩（2KB 起步，动态扩容）
9. sysmon 监控线程（4 大职责）
10. channel 与 goroutine 配合（4 种模式）

---

**掌握 goroutine，你就掌握了 Go 并发编程的精髓！** 🚀
