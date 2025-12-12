# Go Runtime schedule() 函数深度解析

## 概述

`schedule()` 函数是 Go 运行时调度器的核心函数，位于 `runtime/proc.go:4110-4206`。该函数负责选择下一个要执行的 goroutine 并开始执行它。这个函数**永不返回**，而是通过 `execute()` 跳转到选中的 goroutine 执行，形成持续的调度循环。

## 函数签名

```go
func schedule()
```

## 完整执行流程

### 1. 初始化和前置安全检查

#### 1.1 获取当前 M 并检查锁状态

```go
mp := getg().m

if mp.locks != 0 {
    throw("schedule: holding locks")
}
```

**关键点：**
- 获取当前 M (machine/线程)
- **检查是否持有锁**：调度时不能持有任何锁，否则会导致死锁
- 这是一个重要的安全检查，确保调度器不会在持有锁的情况下切换 goroutine

#### 1.2 处理锁定的 goroutine

```go
if mp.lockedg != 0 {
    stoplockedm()
    execute(mp.lockedg.ptr(), false) // Never returns.
}
```

**关键点：**
- 处理通过 `runtime.LockOSThread()` 锁定到特定 OS 线程的 goroutine
- 如果当前 M 被某个 goroutine 锁定，则必须执行那个特定的 goroutine
- `stoplockedm()` 会停止当前 M 并等待被锁定的 goroutine
- `execute()` 永不返回，直接跳转执行

#### 1.3 CGO 调用检查

```go
if mp.incgo {
    throw("schedule: in cgo")
}
```

**关键点：**
- 不能在 CGO 调用期间进行调度
- 因为 CGO 正在使用 M 的 g0 栈，此时切换会导致栈错误

---

### 2. 主调度循环入口

```go
top:
    pp := mp.p.ptr()
    pp.preempt = false
```

**关键点：**
- **`top` 标签**：这是一个重要的跳转点，多处会 `goto top` 重新开始调度流程
- 获取当前的 P (processor，逻辑处理器)
- 清除 P 的抢占标记，准备开始新的调度周期

#### 2.1 自旋状态一致性检查

```go
if mp.spinning && (pp.runnext != 0 || pp.runqhead != pp.runqtail) {
    throw("schedule: spinning with local work")
}
```

**关键点：**
- **自旋 (spinning)** 状态表示 M 正在积极寻找工作
- 如果 M 处于自旋状态，则本地运行队列应该为空
- 这是一个重要的不变量检查：自旋的 M 不应该有本地工作
- `pp.runnext`：P 的下一个待运行 goroutine
- `pp.runqhead` 和 `pp.runqtail`：P 的本地运行队列头尾指针

---

### 3. 查找可运行的 goroutine

```go
gp, inheritTime, tryWakeP := findRunnable() // blocks until work is available
```

**这是最核心的调用！**

#### 3.1 findRunnable() 的工作流程

`findRunnable()` 会按以下顺序尝试获取可运行的 goroutine：

1. **本地运行队列 (local run queue)**
   - 首先检查 P 的本地队列，这是最快的路径

2. **全局运行队列 (global run queue)**
   - 检查全局队列，保证公平性

3. **网络轮询器 (netpoller)**
   - 检查是否有网络 I/O 就绪的 goroutine

4. **工作窃取 (work stealing)**
   - 从其他 P 的本地队列中窃取一半的 goroutine

5. **再次检查全局队列和 netpoller**
   - 确保不会遗漏新添加的工作

6. **如果都没有找到，会阻塞等待**

#### 3.2 返回值说明

- **`gp`**: 找到的可运行 goroutine
- **`inheritTime`**: 是否继承当前的时间片
  - `true`: goroutine 可以继续使用当前时间片（例如从 channel 操作唤醒）
  - `false`: goroutine 需要一个新的时间片（例如从定时器唤醒）
- **`tryWakeP`**: 是否需要唤醒额外的 P
  - `true`: 找到的是特殊 goroutine（GC worker、trace reader 等），需要唤醒另一个 P 处理普通工作

---

### 4. findRunnable() 返回后的处理

#### 4.1 重新获取 P 和清理快照

```go
pp = mp.p.ptr()
mp.clearAllpSnapshot()
```

**关键点：**
- 重新获取 P 指针：`findRunnable()` 可能导致 M 与 P 解绑和重新绑定
- 清理 allp 快照：`findRunnable()` 可能创建了所有 P 的快照，现在不再需要，清理以便 GC

#### 4.2 释放 GC Mark Worker

```go
gcController.releaseNextGCMarkWorker(pp)
```

**关键点：**
- 如果 P 原本被分配了一个 GC mark worker，但 `findRunnable()` 选择了其他 goroutine
- 需要释放该 worker，让其他 P 可以运行它
- 如果 `tryWakeP` 为真，会唤醒另一个 P 来运行该 worker

---

### 5. 处理冻结世界 (Freeze the World)

```go
if debug.dontfreezetheworld > 0 && freezing.Load() {
    // 死锁在这里而不是在 findRunnable 循环中
    lock(&deadlock)
    lock(&deadlock)
}
```

**关键点：**
- 当运行时正在"冻结世界"（通常在 panic 或崩溃时）
- `debug.dontfreezetheworld > 0` 是一个调试标志
- 通过重复锁定同一个 mutex 造成死锁，阻止 goroutine 继续执行
- 目的是保持调度器状态不变，便于调试

---

### 6. 重置自旋状态

```go
if mp.spinning {
    resetspinning()
}
```

**关键点：**
- 如果 M 之前在自旋寻找工作，现在找到了 goroutine
- 需要调用 `resetspinning()` 来：
  1. 将 M 的 spinning 标志设为 false
  2. 减少全局自旋 M 计数器 `sched.nmspinning`
  3. **可能启动新的自旋 M** 以维持系统并行度
- 这是调度器保持活性的关键机制

---

### 7. 处理禁用调度

```go
if sched.disable.user && !schedEnabled(gp) {
    lock(&sched.lock)
    if schedEnabled(gp) {
        // 在获取锁期间，调度可能被重新启用
        unlock(&sched.lock)
    } else {
        // 将 goroutine 放入待运行队列
        sched.disable.runnable.pushBack(gp)
        unlock(&sched.lock)
        goto top  // 重新开始调度
    }
}
```

**关键点：**
- 某些情况下用户调度可能被禁用（例如 STW - Stop The World）
- 如果当前 goroutine 不允许调度，将其放入等待队列
- **`goto top`** 重新开始调度流程，寻找其他可运行的 goroutine
- 双重检查机制：获取锁后再次检查，避免竞态条件

---

### 8. 唤醒额外的 P

```go
if tryWakeP {
    wakep()
}
```

**关键点：**
- 如果要调度的是特殊 goroutine（GC worker 或 trace reader）
- 调用 `wakep()` 唤醒一个空闲的 P 来处理普通工作
- 确保系统的并行度和响应性

#### wakep() 的逻辑
- 检查是否有空闲的 P 和是否需要更多自旋的 M
- 如果需要，唤醒或创建一个新的 M 来运行 P

---

### 9. 处理锁定到特定 M 的 goroutine

```go
if gp.lockedm != 0 {
    // 将自己的 P 交给锁定的 M，然后阻塞等待新的 P
    startlockedm(gp)
    goto top  // 重新开始调度
}
```

**关键点：**
- 如果 goroutine 通过 `LockOSThread()` 锁定到特定的 M
- 当前 M 会将 goroutine 和 P 交给那个特定的 M
- 当前 M **`goto top`** 重新查找其他工作
- 这确保了锁定的 goroutine 在正确的 OS 线程上运行

---

### 10. 执行选中的 goroutine

```go
execute(gp, inheritTime)
```

**最终步骤！**

#### execute() 函数的工作
1. 将 goroutine 的状态从 `_Grunnable` 改为 `_Grunning`
2. 将 goroutine 与当前 M 关联
3. 如果 `inheritTime` 为 false，增加调度计数
4. **调用 `gogo()` 切换到 goroutine 的栈并开始执行**

**关键点：**
- `execute()` **永不返回**
- 通过汇编实现的栈切换跳转到 goroutine 代码
- goroutine 结束后会调用 `goexit()`，最终回到 `schedule()`
- 形成闭环：`schedule() -> execute() -> gogo() -> goroutine 执行 -> goexit() -> schedule()`

---

## 调度流程图

```
┌─────────────────────────────────────────────────┐
│              schedule() 开始                     │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  1. 安全检查                                     │
│     - 不能持有锁                                 │
│     - 检查 lockedg                              │
│     - 检查 CGO 状态                             │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  top: 调度循环入口                               │
│     - 清除抢占标记                               │
│     - 检查自旋状态                               │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  2. findRunnable() - 核心查找                   │
│     - 本地队列                                   │
│     - 全局队列                                   │
│     - netpoller                                 │
│     - 工作窃取                                   │
│     【可能阻塞直到找到工作】                      │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  3. 后续处理                                     │
│     - 重新获取 P                                 │
│     - 清理快照                                   │
│     - 释放 GC worker                            │
└────────────────┬────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  4. 状态管理                                     │
│     - 检查冻结状态                               │
│     - 重置自旋状态                               │
│     - 处理禁用调度 ──────────┐                   │
└────────────────┬────────────┘│                   │
                 │             └───> goto top      │
                 ▼                                 │
┌─────────────────────────────────────────────────┐│
│  5. 特殊情况处理                                 ││
│     - 唤醒额外 P (tryWakeP)                     ││
│     - 处理 lockedm ────────────┘                ││
└────────────────┬─────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────┐
│  6. execute(gp, inheritTime)                    │
│     【永不返回】                                 │
│     - 切换到 goroutine 栈                       │
│     - 开始执行 goroutine                        │
└─────────────────────────────────────────────────┘
                 │
                 ▼
        Goroutine 执行完毕
                 │
                 ▼
         goexit() -> schedule()
         【回到调度循环】
```

---

## 关键数据结构

### M (Machine - OS 线程)
```go
type m struct {
    g0          *g        // 用于执行调度代码的 goroutine
    curg        *g        // 当前运行的 goroutine
    p           puintptr  // 当前关联的 P
    nextp       puintptr  // 下一个要关联的 P
    spinning    bool      // M 是否在自旋寻找工作
    lockedg     guintptr  // 锁定到这个 M 的 goroutine
    locks       int32     // 当前持有的锁数量
    incgo       bool      // 是否在 CGO 调用中
    // ... 更多字段
}
```

### P (Processor - 逻辑处理器)
```go
type p struct {
    status      uint32    // P 的状态
    runqhead    uint32    // 本地运行队列头
    runqtail    uint32    // 本地运行队列尾
    runq        [256]guintptr  // 本地运行队列
    runnext     guintptr  // 下一个运行的 goroutine
    preempt     bool      // 抢占标志
    // ... 更多字段
}
```

### G (Goroutine)
```go
type g struct {
    stack       stack     // goroutine 栈
    sched       gobuf     // 调度信息（PC, SP 等）
    atomicstatus uint32   // goroutine 状态
    m            *m       // 当前运行在哪个 M 上
    lockedm     muintptr  // 锁定到哪个 M
    waitreason  waitReason // 等待原因
    // ... 更多字段
}
```

---

## 关键设计思想

### 1. 永不返回的调度循环
- `schedule()` 通过 `execute()` 和 `goto top` 形成循环
- 每个 M 都在 `schedule() -> goroutine 执行 -> schedule()` 之间循环
- 避免了传统的函数调用返回，减少栈开销

### 2. 工作查找的公平性
- 本地队列优先：减少锁竞争
- 定期检查全局队列：防止全局队列中的 goroutine 饥饿
- 工作窃取：负载均衡，提高 CPU 利用率

### 3. 自旋机制 (Spinning)
- 自旋的 M 积极寻找工作，减少延迟
- 限制自旋 M 的数量，避免浪费 CPU
- `resetspinning()` 确保总有足够的自旋 M

### 4. 特殊情况的细致处理
- **lockedm/lockedg**: 支持 `LockOSThread()` 语义
- **CGO**: 确保 CGO 调用的安全性
- **禁用调度**: 支持 STW 等特殊操作
- **冻结世界**: 便于调试和错误处理

### 5. 并发控制和安全性
- 多重检查防止竞态条件
- 使用原子操作和锁保护共享数据
- 双重检查模式（check-lock-check）

### 6. 性能优化
- 本地队列减少锁竞争
- `runnext` 优化：立即运行最近就绪的 goroutine（局部性）
- 自适应的自旋和阻塞策略

---

## 常见调用路径

### 1. 主动让出 CPU
```
runtime.Gosched()
  -> gosched_m(gp)
    -> goschedImpl(gp, false)
      -> dropg()
      -> globrunqput(gp)  // 放入全局队列
      -> schedule()
```

### 2. 阻塞操作（channel、锁等）
```
chansend/chanrecv/lock
  -> gopark()
    -> park_m()
      -> casgstatus(gp, _Grunning, _Gwaiting)
      -> dropg()
      -> schedule()
```

### 3. 被抢占
```
preemption signal
  -> asyncPreempt()
    -> gopreempt_m(gp)
      -> goschedImpl(gp, true)  // preempted = true
        -> schedule()
```

### 4. Goroutine 退出
```
goroutine 函数返回
  -> goexit()
    -> goexit1()
      -> mcall(goexit0)
        -> goexit0()
          -> casgstatus(gp, _Grunning, _Gdead)
          -> dropg()
          -> schedule()
```

---

## 与其他调度函数的关系

### findRunnable()
- 被 `schedule()` 调用
- 负责查找可运行的 goroutine
- 可能阻塞直到找到工作

### execute()
- 被 `schedule()` 调用
- 执行选中的 goroutine
- 永不返回

### wakep()
- 唤醒一个空闲的 P
- 可能创建新的 M

### resetspinning()
- 重置当前 M 的自旋状态
- 可能启动新的自旋 M

### startlockedm()
- 启动锁定的 M
- 将 P 和 goroutine 交给它

---

## 性能考量

### 快速路径优化
1. **本地队列优先**: 无锁访问，最快
2. **runnext 优化**: 缓存局部性，减少延迟
3. **自旋减少唤醒延迟**: 有工作时立即处理

### 负载均衡
1. **工作窃取**: 从忙碌的 P 窃取工作
2. **全局队列**: 定期检查，防止饥饿
3. **动态 P 管理**: 根据负载调整 GOMAXPROCS

### 避免过度竞争
1. **每个 P 有独立的本地队列**: 减少锁竞争
2. **批量操作**: 一次窃取多个 goroutine
3. **限制自旋 M 数量**: 避免浪费 CPU

---

## 调试和监控

### 相关环境变量和调试标志
```go
GODEBUG=schedtrace=1000  // 每秒打印调度信息
GODEBUG=scheddetail=1    // 打印详细调度信息
```

### 关键指标
- `sched.nmspinning`: 当前自旋的 M 数量
- `sched.npidle`: 空闲的 P 数量
- `sched.nmidle`: 空闲的 M 数量
- `sched.runqsize`: 全局队列长度

---

## 总结

`schedule()` 函数是 Go 调度器的核心，它通过精心设计的算法和机制实现了：

1. **高效的 goroutine 调度**: 快速查找和执行可运行的 goroutine
2. **公平性**: 通过全局队列和工作窃取避免饥饿
3. **低延迟**: 通过自旋机制减少唤醒延迟
4. **高并发性**: 通过 P 的本地队列减少锁竞争
5. **灵活性**: 支持各种特殊情况（locked goroutine、CGO、STW 等）

这个函数体现了 Go 调度器的设计哲学：**在保证公平性的前提下，最大化性能和并发性**。每个 M 都在这个函数和 goroutine 执行之间循环，共同构成了 Go 高效的并发运行时。

---

## 参考资料

- [Go Scheduler Design Doc](https://golang.org/s/go11sched)
- [Analysis of the Go runtime scheduler](http://www.cs.columbia.edu/~aho/cs6998/reports/12-12-11_DeshpandeSponslerWeiss_GO.pdf)
- [Scalable Go Scheduler Design Doc](https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw)
- Go Runtime Source Code: `runtime/proc.go`

---

**文档版本**: 1.0
**最后更新**: 2025-12-11
**适用 Go 版本**: Go 1.23+
