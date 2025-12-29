# 简化版 GMP 模型架构设计

## 项目目标

通过实现一个简化版的 GMP 调度模型，深入理解 Go runtime 的核心调度机制。以**测试驱动**的方式逐步实现，每个功能点都对应 Go runtime 源码位置。

## 核心设计原则

1. **简化但不失真**：保留 GMP 模型的核心逻辑，去掉系统调用、抢占、GC 等复杂特性
2. **可观测性**：添加日志和状态输出，方便调试和理解
3. **测试驱动**：每个功能都先写测试，再实现
4. **渐进式构建**：从最小可运行版本开始，逐步添加特性

---

## 一、整体架构

```
┌─────────────────────────────────────────────┐
│            Scheduler (调度器)                │
│  - 全局运行队列 (Global Runqueue)            │
│  - P 列表管理                                │
│  - M 列表管理                                │
└─────────────────────────────────────────────┘
                    │
        ┌───────────┼───────────┐
        ▼           ▼           ▼
    ┌─────┐     ┌─────┐     ┌─────┐
    │  P0 │     │  P1 │     │  P2 │    P (Processor)
    │ 本地│     │ 本地│     │ 本地│    - 本地运行队列 (256)
    │ 队列│     │ 队列│     │ 队列│    - runnext (优先执行)
    └─────┘     └─────┘     └─────┘
        │           │           │
        ▼           ▼           ▼
    ┌─────┐     ┌─────┐     ┌─────┐
    │  M0 │     │  M1 │     │  M2 │    M (Machine/Thread)
    │     │     │     │     │     │    - 绑定一个 P
    │  G1 │     │  G3 │     │  G5 │    - 执行当前 G
    └─────┘     └─────┘     └─────┘
```

---

## 二、核心数据结构

### 2.1 G (Goroutine)

**对应源码**: `runtime/runtime2.go:407` (`type g struct`)

```go
G 的核心字段：
- goid: goroutine ID (全局递增，int64)
- status: 状态 (idle/runnable/running/waiting/dead)
- fn: 要执行的函数
- result: 执行结果 (可选，用于测试验证)
- m: 当前执行它的 M (running 时)
- waitSince: 等待开始时间 (用于统计)
```

**状态转换**:
```
idle → runnable → running → dead
         ↑          ↓
         └─ waiting ─┘
```

**Go runtime 对应**:
- `runtime.newproc` (proc.go:4239): 创建新 G
- `runtime.goexit` (proc.go:3581): G 执行完毕

---

### 2.2 P (Processor)

**对应源码**: `runtime/runtime2.go:629` (`type p struct`)

```go
P 的核心字段：
- id: P 的 ID (0 到 GOMAXPROCS-1，int32)
- status: 状态 (idle/running/syscall/dead)
- runqhead/runqtail: 本地队列头尾指针 (uint32)
- runq: 本地循环队列 [256]*G
- runnext: 下一个优先执行的 G (用于新创建的 G)
- m: 当前绑定的 M
```

**本地队列操作**:
- `runqput`: 把 G 放入本地队列 (满了则转移一半到全局队列)
- `runqget`: 从本地队列取 G (优先取 runnext)
- `runqsteal`: 被其他 P 窃取一半的 G

**Go runtime 对应**:
- `runtime.runqput` (proc.go:6084): 入队
- `runtime.runqget` (proc.go:6154): 出队
- `runtime.runqsteal` (proc.go:6245): 窃取

---

### 2.3 M (Machine)

**对应源码**: `runtime/runtime2.go:536` (`type m struct`)

```go
M 的核心字段：
- id: M 的 ID (int64)
- p: 绑定的 P
- curg: 当前正在执行的 G
- spinning: 是否处于自旋状态 (正在找工作)
- thread: 底层 OS 线程 (简化版可用 goroutine 模拟)
```

**M 的生命周期**:
1. 创建: `newm()` 创建新 M
2. 启动: `mstart()` 进入调度循环
3. 执行: `execute(G)` 执行具体的 G
4. 调度: `schedule()` 找下一个 G 执行

**Go runtime 对应**:
- `runtime.newm` (proc.go:2385): 创建 M
- `runtime.mstart` (proc.go:1396): M 启动
- `runtime.schedule` (proc.go:3365): 调度循环

---

### 2.4 Sched (全局调度器)

**对应源码**: `runtime/runtime2.go:779` (`type schedt struct`)

```go
Sched 的核心字段：
- goidgen: 全局 G ID 生成器 (int64)
- mnext: 全局 M ID 生成器 (int64)
- runq: 全局运行队列 (链表或切片)
- runqsize: 全局队列长度
- pidle: 空闲 P 链表
- midle: 空闲 M 链表
- npidle: 空闲 P 数量
- nmspinning: 正在自旋的 M 数量
- allp: 所有 P 的数组 []*P
- allm: 所有 M 的链表
```

**Go runtime 对应**:
- `runtime.sched` (proc.go:789): 全局调度器实例

---

## 三、核心流程设计

### 3.1 初始化流程

```
schedinit():
  1. 设置 GOMAXPROCS (P 的数量)
  2. 创建所有 P (allp)
  3. 初始化全局队列
  4. 创建 M0 并绑定 P0
```

**对应源码**: `runtime.schedinit` (proc.go:720)

**测试用例**:
```go
TestSchedInit():
  - 验证创建了正确数量的 P
  - 验证全局队列已初始化
  - 验证 M0-P0 绑定关系
```

---

### 3.2 创建 Goroutine

```
newproc(fn):
  1. 创建新的 G，设置 fn 和 goid
  2. 设置状态为 runnable
  3. 尝试放入当前 P 的 runnext
  4. 如果 runnext 已有 G，则把旧 G 放入本地队列
  5. 如果本地队列满，转移一半到全局队列
```

**对应源码**: `runtime.newproc` (proc.go:4239)

**测试用例**:
```go
TestNewProc():
  - 创建单个 G，验证 goid 递增
  - 创建多个 G，验证 runnext 机制
  - 创建 256+ 个 G，验证队列满时转移到全局队列
```

---

### 3.3 调度核心循环 (schedule)

```
schedule():
  while true:
    1. 检查全局队列 (每 61 次调度检查一次，防止饥饿)
    2. 从当前 P 的本地队列获取 G
    3. 如果没有，调用 findrunnable() 找 G
    4. 执行 execute(G)
```

**对应源码**: `runtime.schedule` (proc.go:3365)

**关键点**:
- **全局队列检查频率**: `schedtick % 61 == 0` (防止全局队列饥饿)
- **调度次数**: `schedtick++` 每次调度递增

---

### 3.4 查找可运行的 G (findrunnable)

```
findrunnable():
  1. 检查当前 P 的 runnext
  2. 检查当前 P 的本地队列
  3. 检查全局队列
  4. 检查网络轮询器 (简化版跳过)
  5. 尝试从其他 P 窃取 (work stealing)
  6. 如果都没有，M 进入自旋或休眠
```

**对应源码**: `runtime.findrunnable` (proc.go:2839)

**工作窃取逻辑**:
```
stealWork():
  随机选一个起始 P
  遍历所有其他 P:
    如果该 P 的本地队列 > 1:
      窃取一半的 G
      返回
```

**对应源码**: `runtime.stealWork` (proc.go:3116)

---

### 3.5 执行 G (execute)

```
execute(G):
  1. 设置 G.status = running
  2. 设置 G.m = 当前 M
  3. 设置 M.curg = G
  4. 调用 G.fn() 执行实际函数
  5. 执行完后调用 goexit()
```

**goexit()**:
```
goexit():
  1. 设置 G.status = dead
  2. 清理 G.m 和 M.curg
  3. 把 G 放入空闲列表 (可复用)
  4. 重新进入 schedule() 循环
```

**对应源码**:
- `runtime.execute` (proc.go:2741)
- `runtime.goexit0` (proc.go:3582)

---

## 四、渐进式实现路径

### Phase 1: 最小可运行版本 (Single-threaded)
```
目标: 单个 M 执行多个 G
组件:
  - G 结构和创建
  - P 的本地队列 (固定大小切片)
  - M 的简单调度循环
测试:
  - 创建 10 个 G，顺序执行
  - 验证所有 G 都执行完毕
```

### Phase 2: 添加全局队列
```
目标: 支持本地队列溢出到全局队列
组件:
  - 全局队列 (Sched.runq)
  - runqput 的溢出逻辑
  - schedule 的全局队列检查
测试:
  - 创建 300 个 G，验证全局队列有 G
  - 验证每 61 次调度会检查全局队列
```

### Phase 3: 多 P 多 M
```
目标: 支持多个 P 和 M 并发调度
组件:
  - 创建多个 P (GOMAXPROCS)
  - 创建多个 M，每个 M 绑定一个 P
  - M 之间独立调度
测试:
  - 创建 1000 个 G，4 个 P
  - 验证负载均衡 (每个 P 执行的 G 数量相近)
```

### Phase 4: 工作窃取
```
目标: 实现 work stealing
组件:
  - findrunnable 的窃取逻辑
  - runqsteal 实现
  - M 的自旋状态管理
测试:
  - P0 有 100 个 G，其他 P 空闲
  - 验证其他 P 能窃取到 G
  - 统计窃取次数
```

### Phase 5: 可观测性和调优
```
目标: 添加监控和统计
组件:
  - 每个 P 的执行统计
  - 全局队列使用率
  - 窃取成功率
  - 类似 GODEBUG=schedtrace 的输出
测试:
  - 运行大量 G，输出调度统计
```

---

## 五、关键算法细节

### 5.1 本地队列的循环数组实现

```
容量: 256
字段: head, tail (uint32)
入队: runq[tail % 256] = g; tail++
出队: g = runq[head % 256]; head++
判满: tail - head >= 256
```

**对应源码**: `runtime/runtime2.go:643-645`

### 5.2 runnext 机制

新创建的 G 优先放入 runnext，提高局部性：
```
if runnext != nil:
  old = runnext
  runnext = newG
  runqput(old)  // 把旧的放入队列
else:
  runnext = newG
```

**对应源码**: `runtime.runqput` (proc.go:6084)

### 5.3 窃取数量计算

窃取目标 P 的一半 G：
```
n = (target.tail - target.head) / 2
从 target.head 开始取 n 个 G
更新 target.head += n
```

**对应源码**: `runtime.runqsteal` (proc.go:6245)

### 5.4 防止全局队列饥饿

每 61 次调度检查一次全局队列：
```
if schedtick % 61 == 0:
  从全局队列取一个 G
  return G
```

**为什么是 61？**: 质数，避免周期性冲突

**对应源码**: `runtime.schedule` (proc.go:3438)

---

## 六、简化掉的部分

为了聚焦核心调度逻辑，以下特性不实现：

1. **系统调用处理**: `entersyscall/exitsyscall` (太复杂)
2. **抢占机制**: `preemptone/preemptall` (需要信号)
3. **网络轮询器**: `netpoll` (需要 epoll/kqueue)
4. **GC 协调**: `gcStart/gcMarkDone` (另一个大主题)
5. **栈增长**: `morestack` (需要汇编)
6. **TLS 管理**: `settls` (平台相关)

### 保留的核心

1. **GMP 绑定关系**
2. **本地队列 + 全局队列**
3. **工作窃取算法**
4. **调度循环逻辑**
5. **G 的状态转换**

---

## 七、项目结构建议

```
src/gmp/
├── types.go          # G/M/P/Sched 结构定义
├── g.go              # G 的创建和管理
├── g_test.go         # G 的测试
├── p.go              # P 的队列操作
├── p_test.go         # P 的测试
├── m.go              # M 的调度循环
├── m_test.go         # M 的测试
├── sched.go          # 全局调度器
├── sched_test.go     # 调度器测试
├── schedule.go       # schedule/findrunnable 实现
├── schedule_test.go  # 调度逻辑测试
├── steal.go          # 工作窃取算法
├── steal_test.go     # 窃取测试
└── stats.go          # 统计和日志

src/docs/
├── README.md         # 文档索引
├── FAQ.md            # 常见问题解答
├── architecture.md   # 架构设计（本文档）
└── tdd-guide.md      # TDD 开发指南
```

---

## 八、对照 Go Runtime 学习要点

### 阅读顺序建议

1. **runtime/runtime2.go**
   - 先看 g, m, p, schedt 结构定义
   - 理解每个字段的作用

2. **runtime/proc.go**
   - `schedinit()`: 初始化流程
   - `newproc()`: 创建 G
   - `schedule()`: 调度循环
   - `findrunnable()`: 查找可运行 G
   - `execute()`: 执行 G

3. **关键常量**
   - `_GOMAXPROCS`: proc.go:720
   - `_p_` 本地队列大小: runtime2.go:644 (256)
   - 全局队列检查频率: proc.go:3438 (61)

---

## 九、学习路径建议

**三周计划**：

### 第一周: 基础结构
```
Day 1-2: 定义 G/M/P/Sched 结构
  测试: TestGCreate, TestPCreate, TestMCreate

Day 3-4: 实现 P 的本地队列
  测试: TestRunqPut, TestRunqGet, TestRunqFull

Day 5-7: 实现单 M 调度循环
  测试: TestSingleMSchedule (10 个简单 G)
```

### 第二周: 多 P 调度
```
Day 8-10: 实现全局队列
  测试: TestGlobalQueue, TestQueueOverflow

Day 11-12: 实现多 M 绑定多 P
  测试: TestMultiP (4P 执行 100 G)

Day 13-14: 负载均衡测试
  测试: TestLoadBalance (验证每个 P 执行数量)
```

### 第三周: 工作窃取
```
Day 15-17: 实现 findrunnable 和 stealWork
  测试: TestWorkStealing (不平衡负载)

Day 18-19: 添加自旋状态管理
  测试: TestSpinning (验证空闲 M 的行为)

Day 20-21: 完整集成测试
  测试: TestFullScheduler (1000 G, 4P, 8M)
```

---

## 十、调试和可视化建议

### 日志输出格式
```
[调度器] 创建 P0-P3
[P0] 入队 G1 (本地队列)
[M0] 绑定 P0
[M0] 执行 G1
[G1] 开始执行: task-1
[G1] 完成
[M0] G1 退出，重新调度
[M1] 从 P0 窃取 50 个 G
```

### 统计信息 (类似 schedtrace)
```
SCHED 1000ms: gomaxprocs=4 idleprocs=0 threads=4 spinningthreads=1
  P0: runqsize=23, exectime=250ms, steals=5
  P1: runqsize=18, exectime=245ms, steals=3
  P2: runqsize=20, exectime=255ms, steals=4
  P3: runqsize=19, exectime=250ms, steals=2
  Global: runqsize=10
```

---

## 参考资料

### Go Runtime 源码位置 (go1.21+)
- `runtime/runtime2.go`: 核心数据结构
- `runtime/proc.go`: 调度器主逻辑
- `runtime/asm_amd64.s`: 汇编入口点

### 推荐阅读
1. Go 语言设计与实现 - 调度器章节
2. GopherCon 2018: Kavya Joshi - The Scheduler Saga
3. Scalable Go Scheduler Design Doc

---

**通过实现这个简化版 GMP，你会对 Go 调度器有深刻的理解！**
