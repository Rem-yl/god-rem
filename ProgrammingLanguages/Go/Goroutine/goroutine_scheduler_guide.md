# Go 调度器源码学习指南

## 为什么 Go 调度器看起来很复杂？

- 涉及操作系统概念（线程、上下文切换）
- 涉及汇编代码（栈切换、寄存器操作）
- 涉及并发控制（原子操作、内存屏障）
- 代码分散在多个文件中
- 优化代码较多，不够直观

**好消息**：核心思想其实很简单，我们可以分层理解。

## 学习路线图

```
第一层：理解基本概念
    ↓
第二层：理解 GMP 模型
    ↓
第三层：阅读核心数据结构
    ↓
第四层：跟踪关键流程
    ↓
第五层：深入优化细节
```

---

## 第一层：理解基本概念（预备知识）

在看源码前，必须先理解这些概念：

### 1. 为什么需要调度器？

**问题**：操作系统线程太重
- 创建慢（几 KB 栈 + 系统调用）
- 切换慢（用户态↔内核态）
- 数量有限（几千个）

**Go 的方案**：用户态调度
- Goroutine 很轻（2KB 初始栈）
- 在用户态切换（无需系统调用）
- 可以创建几十万个

### 2. M:N 调度模型

```
M 个 Goroutine 映射到 N 个 OS 线程

例如：
100,000 个 Goroutine → 运行在 → 8 个 OS 线程上
```

**核心挑战**：
- 如何分配 goroutine 到线程？
- 如何实现抢占？
- 如何负载均衡？

### 3. 协作式 vs 抢占式

**协作式调度**（早期 Go）：
- Goroutine 主动让出 CPU（如函数调用、channel 操作）
- 问题：死循环会霸占 CPU

**抢占式调度**（Go 1.14+）：
- 运行时可以强制暂停 goroutine
- 基于信号实现

---

## 第二层：GMP 模型（核心架构）

这是理解 Go 调度器的**关键**！

### G - Goroutine

```
G = 用户代码的执行单元

包含：
├─ 栈空间（初始 2KB，可扩展）
├─ 执行状态（PC、SP 等寄存器）
├─ 调度信息（优先级、等待原因）
└─ 关联的 M（正在执行它的线程）
```

**状态机**：
```
_Gidle → _Grunnable → _Grunning → _Gwaiting → _Grunnable
         (可运行)     (运行中)     (等待中)
```

### M - Machine（OS 线程）

```
M = 操作系统线程的抽象

包含：
├─ 关联的 G（当前正在执行的 goroutine）
├─ 关联的 P（需要 P 才能执行 G）
├─ g0（特殊的调度栈）
└─ 线程本地存储
```

**关键点**：
- M 代表真正的 OS 线程
- M 必须绑定 P 才能执行 G
- M 的数量可以动态增长（如遇阻塞调用）

### P - Processor（逻辑处理器）

```
P = 执行 Go 代码的上下文

包含：
├─ 本地 G 队列（最多 256 个）
├─ 当前正在执行的 G
├─ runnext（下一个优先执行的 G）
├─ mcache（内存分配缓存）
└─ 状态（running/idle/syscall）
```

**关键点**：
- P 的数量 = GOMAXPROCS（默认等于 CPU 核心数）
- P 持有 G 的本地队列，减少竞争
- P 是连接 M 和 G 的桥梁

### GMP 协作关系

```
全局视图：

全局 G 队列（Global Queue）
    ↓ 窃取
P1 本地队列 → M1 → 执行 G1
P2 本地队列 → M2 → 执行 G2
P3 本地队列 → M3 → 执行 G3
P4 本地队列 → 空闲
```

**工作流程**：
1. M 必须持有 P 才能执行 G
2. M 从 P 的本地队列获取 G
3. 如果 P 本地队列空了，从全局队列或其他 P 窃取
4. 如果 M 发生阻塞（如系统调用），P 会被传递给其他 M

---

## 第三层：核心数据结构（源码位置）

开始看源码！建议顺序：

### 1. `runtime/runtime2.go` - 核心结构定义

**从这里开始！** 这个文件定义了所有核心数据结构。

#### 重点阅读：

**g 结构体**（约 400 行）：
```go
// 位置：runtime/runtime2.go:400+
type g struct {
    stack       stack   // 栈信息
    stackguard0 uintptr // 栈保护

    sched gobuf         // 调度信息（保存的寄存器）

    atomicstatus uint32 // 状态
    goid         int64  // goroutine ID

    m            *m     // 当前使用的 M

    // ... 还有很多字段
}
```

**阅读建议**：
- 先看注释，理解每个字段的作用
- 关注 `stack`、`sched`、`m` 这几个关键字段
- 暂时忽略性能优化相关的字段

**m 结构体**（约 200 行）：
```go
// 位置：runtime/runtime2.go:500+
type m struct {
    g0      *g     // 调度栈
    curg    *g     // 当前执行的 G
    p       puintptr // 关联的 P

    // ...
}
```

**p 结构体**（约 100 行）：
```go
// 位置：runtime/runtime2.go:600+
type p struct {
    status      uint32  // 状态

    runqhead uint32     // 本地队列头
    runqtail uint32     // 本地队列尾
    runq     [256]guintptr // 本地队列
    runnext  guintptr    // 下一个要运行的 G

    // ...
}
```

**学习任务**：
- [ ] 阅读 g、m、p 结构体定义
- [ ] 理解每个关键字段的作用
- [ ] 画出三者的关系图

### 2. `runtime/proc.go` - 调度核心逻辑

这是**最重要**的文件（约 7000 行），包含所有调度逻辑。

**不要从头到尾读！** 按功能模块读：

---

## 第四层：跟踪关键流程（推荐阅读顺序）

### 流程 1：创建 Goroutine

**入口**：`go` 关键字 → `newproc()`

```
用户代码：go func() { ... }
    ↓
编译器转换为：runtime.newproc()
    ↓
runtime/proc.go:4249
```

**阅读路径**：
```
1. newproc()         // proc.go:4249 - 入口
   ├─ 获取参数
   └─ 调用 newproc1()

2. newproc1()        // proc.go:4280 - 核心
   ├─ 从 P 的缓存获取 G（或分配新的）
   ├─ 初始化 G 的栈和调度信息
   ├─ 设置 G 的状态为 _Grunnable
   └─ 放入 P 的本地队列或全局队列

3. runqput()         // proc.go:5900 - 入队
   ├─ 尝试放入 runnext
   └─ 否则放入本地队列
```

**学习任务**：
- [ ] 阅读 `newproc()` 和 `newproc1()` 函数
- [ ] 理解 G 是如何被创建和初始化的
- [ ] 理解 G 如何被放入队列

**关键代码片段**：
```go
// proc.go:4280
func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, callergp *g, callerpc uintptr) *g {
    _g_ := getg()  // 获取当前 G
    _p_ := _g_.m.p.ptr()  // 获取当前 P

    newg := gfget(_p_)  // 尝试从 P 的缓存获取 G
    if newg == nil {
        newg = malg(_StackMin)  // 分配新 G
    }

    // 初始化 newg 的栈和寄存器
    // ...

    runqput(_p_, newg, true)  // 放入队列
    return newg
}
```

### 流程 2：调度循环

**核心**：`schedule()` 函数 - 调度器的心脏

```
M 的执行循环：
    ↓
schedule() - 选择下一个 G
    ↓
execute() - 执行 G
    ↓
gogo() - 切换到 G 的栈（汇编）
    ↓
G 的用户代码执行
    ↓
goexit() - G 执行完毕
    ↓
回到 schedule()
```

**阅读路径**：
```
1. schedule()        // proc.go:3300 - 调度核心
   ├─ 每 61 次从全局队列获取
   ├─ 检查 P 的 runnext
   ├─ 从 P 的本地队列获取
   ├─ findRunnable() - 如果本地没有
   └─ execute(G) - 执行选中的 G

2. findRunnable()    // proc.go:2800 - 查找可运行的 G
   ├─ 检查本地队列
   ├─ 检查全局队列
   ├─ 检查网络轮询器（netpoller）
   ├─ 工作窃取（stealWork）
   └─ 如果都没有，进入休眠

3. execute()         // proc.go:2700 - 执行 G
   ├─ 设置 G 的状态为 _Grunning
   ├─ 关联 G 和 M
   └─ gogo(&g.sched) - 跳转到 G（汇编）

4. gogo()            // asm_amd64.s - 汇编代码
   └─ 恢复 G 的寄存器，跳转到 G 的 PC
```

**学习任务**：
- [ ] 阅读 `schedule()` 函数
- [ ] 理解 G 的选择策略
- [ ] 阅读 `findRunnable()` - 理解工作窃取
- [ ] 了解 `execute()` 和 `gogo()`（汇编可以先跳过）

**关键代码片段**：
```go
// proc.go:3300
func schedule() {
    _g_ := getg()  // 获取当前 g0（调度栈）

top:
    var gp *g

    // 每 61 次从全局队列获取一次
    if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
        gp = globrunqget(_g_.m.p.ptr(), 1)
    }

    if gp == nil {
        gp = runqget(_g_.m.p.ptr())  // 从本地队列获取
    }

    if gp == nil {
        gp = findRunnable()  // 全局查找
    }

    execute(gp, false)  // 执行
}
```

### 流程 3：系统调用处理

**问题**：如果 G 执行系统调用（如读文件），会阻塞 M，怎么办？

**方案**：P/M 分离

```
G1 执行系统调用
    ↓
M1 被阻塞（进入内核态）
    ↓
P1 从 M1 分离
    ↓
P1 寻找新的 M2（或创建）
    ↓
M2 绑定 P1，继续执行其他 G
    ↓
系统调用返回后，M1 尝试重新获取 P
```

**阅读路径**：
```
1. entersyscall()    // proc.go:3600 - 进入系统调用
   ├─ 保存 G 的状态
   ├─ 设置 P 的状态为 _Psyscall
   └─ P 可能被其他 M 抢走

2. exitsyscall()     // proc.go:3700 - 退出系统调用
   ├─ 尝试重新获取原来的 P
   ├─ 或者尝试获取空闲的 P
   └─ 如果都失败，放入全局队列
```

**学习任务**：
- [ ] 阅读 `entersyscall()` 和 `exitsyscall()`
- [ ] 理解 P/M 分离的机制

### 流程 4：抢占调度

**Go 1.14 之前**：协作式抢占
- 依赖函数调用时的栈检查

**Go 1.14+**：异步抢占
- 基于信号（SIGURG）

**阅读路径**：
```
1. retake()          // proc.go:5000 - 监控线程定期调用
   ├─ 遍历所有 P
   ├─ 检查是否运行过长
   └─ preemptone(G) - 抢占

2. preemptone()      // proc.go:5100
   └─ 发送抢占信号

3. asyncPreempt()    // preempt_amd64.s - 信号处理
   └─ 保存 G 的状态，切换到调度循环
```

**学习任务**：
- [ ] 阅读 `retake()` 函数
- [ ] 了解抢占的触发条件

### 流程 5：工作窃取

**负载均衡机制**：空闲的 P 从忙碌的 P 窃取 G

```
P1: [G1 G2 G3 G4 G5] - 忙碌
P2: [G6]             - 空闲
P3: []               - 空闲
    ↓
P3 从 P1 窃取一半：
P1: [G1 G2 G3]
P3: [G4 G5]
```

**阅读路径**：
```
1. stealWork()       // proc.go:2900 - 窃取工作
   ├─ 随机选择一个 P
   ├─ runqsteal() - 窃取一半的 G
   └─ 返回窃取到的 G

2. runqsteal()       // proc.go:6000
   └─ 从目标 P 的本地队列窃取一半
```

---

## 第五层：深入优化细节（进阶）

掌握核心流程后，可以研究这些优化：

### 1. 本地缓存

```
为什么需要本地队列？
- 避免全局锁竞争
- 利用 CPU 缓存局部性

结构：
├─ runnext：下一个优先执行（无锁）
├─ 本地队列：256 个 G（无锁环形队列）
└─ 全局队列：所有 G（需要锁）
```

**阅读**：`runqput()`、`runqget()` 的实现

### 2. g0 栈

```
每个 M 有两个栈：
├─ g0 栈：用于调度代码
└─ 用户 G 栈：用于用户代码

为什么需要 g0？
- 调度代码需要稳定的栈
- 用户 G 栈可能很小或被换出
```

**阅读**：`mcall()` - 切换到 g0 执行函数

### 3. Sysmon 线程

```
特殊的监控线程（不绑定 P）：
├─ 抢占长时间运行的 G
├─ 释放长时间 syscall 的 P
├─ 触发垃圾回收
└─ 网络轮询器
```

**阅读**：`sysmon()` 函数 - proc.go:4800

### 4. 内存分配集成

```
P 持有 mcache：
└─ 每个 P 有独立的内存缓存
   └─ 分配小对象时无需锁
```

**阅读**：`runtime/mcache.go`

---

## 推荐的实践学习方法

### 1. 边读边画图

每个流程都画出来：
```
[示例]
创建 Goroutine：
  用户代码
     ↓
  newproc()
     ↓
  gfget() → 获取 G
     ↓
  初始化 G.sched
     ↓
  runqput() → 放入队列
```

### 2. 边读边写注释

在源码中加入自己的理解：
```go
// 我的理解：这里是从 P 的缓存获取空闲的 G
newg := gfget(_p_)
```

### 3. 使用 Delve 调试器跟踪

```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试程序
dlv debug your_program.go

# 设置断点
(dlv) break runtime.newproc1
(dlv) break runtime.schedule

# 运行并观察调用栈
(dlv) continue
(dlv) stack
```

### 4. 写测试代码验证理解

```go
// 测试 goroutine 调度
func TestScheduler() {
    runtime.GOMAXPROCS(2)

    for i := 0; i < 10; i++ {
        go func(id int) {
            fmt.Printf("G %d on M %d\n", id, runtime.NumGoroutine())
        }(i)
    }

    time.Sleep(time.Second)
}
```

### 5. 参考资料配合阅读

- **官方设计文档**：
  - https://go.dev/src/runtime/HACKING.md
  - https://go.dev/s/go11sched （经典论文）

- **优秀博客**：
  - "Scalable Go Scheduler Design Doc"
  - "The Go scheduler" by Morsing

- **视频**：
  - GopherCon 关于调度器的演讲

---

## 具体的阅读计划（4周）

### 第 1 周：理论基础
- [ ] 复习进程/线程概念
- [ ] 理解 GMP 模型
- [ ] 阅读 `runtime2.go` 中的结构体定义
- [ ] 画出 GMP 关系图

### 第 2 周：核心流程
- [ ] 跟踪 goroutine 创建流程
- [ ] 理解 `schedule()` 调度循环
- [ ] 理解本地队列和全局队列
- [ ] 写测试代码验证理解

### 第 3 周：高级特性
- [ ] 系统调用处理
- [ ] 抢占调度
- [ ] 工作窃取
- [ ] 使用 delve 调试验证

### 第 4 周：深入优化
- [ ] g0 栈机制
- [ ] sysmon 线程
- [ ] 与内存分配器的集成
- [ ] 阅读汇编代码（可选）

---

## 核心源码文件清单

### 必读（按顺序）
1. `runtime/runtime2.go` - 数据结构定义 ⭐⭐⭐⭐⭐
2. `runtime/proc.go` - 调度核心逻辑 ⭐⭐⭐⭐⭐
3. `runtime/stack.go` - 栈管理 ⭐⭐⭐

### 重要
4. `runtime/asm_amd64.s` - 汇编实现（栈切换）⭐⭐⭐
5. `runtime/netpoll.go` - 网络轮询器 ⭐⭐⭐
6. `runtime/preempt.go` - 抢占实现 ⭐⭐⭐

### 进阶
7. `runtime/mgc.go` - 垃圾回收与调度的交互 ⭐⭐
8. `runtime/mheap.go` - 堆内存管理 ⭐⭐
9. `runtime/sys_linux_amd64.s` - 系统调用包装 ⭐

---

## 常见困惑解答

### Q1: 为什么有这么多汇编代码？
**A**: 汇编用于：
- 栈切换（需要直接操作 SP、PC 寄存器）
- 系统调用包装
- 性能关键路径

**建议**：初期可以跳过汇编，只需理解其功能。

### Q2: g0 是什么？
**A**: 每个 M 的调度栈。调度代码在 g0 上执行，用户代码在普通 G 上执行。

### Q3: 为什么需要 P？
**A**: P 的设计是关键优化：
- 解耦 M 和 G（M 可以阻塞，P 可以转移）
- 提供本地队列（减少锁竞争）
- 提供本地缓存（内存分配）

### Q4: 如何验证自己的理解？
**A**:
- 写测试代码观察行为
- 使用 `GODEBUG=schedtrace=1000` 查看调度信息
- 使用 delve 设置断点跟踪

---

## 总结：核心要点

1. **GMP 模型是核心**：
   - G = goroutine
   - M = OS 线程
   - P = 逻辑处理器（连接 M 和 G）

2. **调度循环是心脏**：
   - `schedule()` 选择 G
   - `execute()` 执行 G
   - `gogo()` 切换栈

3. **优化的关键**：
   - 本地队列减少竞争
   - 工作窃取负载均衡
   - P/M 分离应对阻塞

4. **不要纠结细节**：
   - 先理解大框架
   - 再深入具体实现
   - 汇编可以最后看

**记住**：调度器的目标是让 CPU 忙起来，让 goroutine 快起来！

祝学习顺利！
