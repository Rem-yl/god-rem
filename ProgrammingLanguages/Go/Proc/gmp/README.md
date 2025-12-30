# 简化版 GMP 调度器实现

## 项目概述

这是一个简化版的 Go Runtime GMP（Goroutine-Machine-Processor）调度模型实现，用于学习和理解 Go 的调度器原理。

## 实现功能

### ✅ Phase 1: 基础数据结构
- **G (Goroutine)**: 用户级协程
  - goid: 唯一标识
  - status: 状态（Idle, Runnable, Running, Waiting, Dead）
  - fn: 要执行的函数
  - m: 关联的 M
  - g0: 指向 g0（对于普通 G，指向关联 M 的 g0）

- **M (Machine)**: 工作线程
  - id: 唯一标识
  - g0: 调度器专用的 g
  - curg: 当前正在运行的 g
  - p: 关联的 P

- **P (Processor)**: 调度上下文
  - id: 唯一标识
  - status: 状态（Idle, Running, Syscall, GCstop, Dead）
  - runq: 本地运行队列（容量 256）
  - runnext: 下一个优先运行的 G
  - m: 关联的 M

- **Schedt**: 全局调度器
  - goidgen: G ID 生成器
  - maxmcount: 最大 M 数量
  - runq: 全局运行队列
  - allp: 所有 P 的列表
  - pidle: 空闲 P 链表

- **getg() / setg()**: 获取和设置当前 G（简化实现使用全局变量）

### ✅ Phase 2: 队列操作
- **本地队列**：
  - `runqput()`: 放入 P 的本地队列
  - `runqget()`: 从 P 的本地队列取出
  - `runnext` 优化：优先执行最近创建的 G

- **全局队列**：
  - `globrunqput()`: 放入全局队列
  - `globrunqget()`: 从全局队列获取
  - 队列满时自动分流到全局队列

### ✅ Phase 3: 调度器核心
- **procresize()**: 创建和调整 P 的数量
- **schedinit()**: 调度器初始化
- **newproc()**: 创建新的 G 并加入队列
- **findrunnable()**: 查找可运行的 G
  1. 本地队列
  2. 全局队列
  3. 工作窃取
- **schedule()**: 调度循环
- **execute()**: 执行 G
- **goexit()**: G 退出后的清理

### ✅ Phase 5: 工作窃取
- **runqsteal()**: 从其他 P 窃取 G
- **runqstealFromP()**: 从指定 P 窃取一半的 G
- 自动负载均衡

## 核心流程

### 1. 初始化流程
```
osinit()
  └─> initG0M0()
      ├─> 创建 g0 和 m0
      ├─> 设置双向关联 (g0.m = m0, m0.g0 = g0)
      └─> setg(g0)  // 设置当前 G 为 g0

schedinit()
  ├─> osinit()
  ├─> 设置 maxmcount = 10000
  └─> procresize(GOMAXPROCS)
      ├─> 创建 P
      ├─> 将 P[0] 分配给 m0
      └─> 其余 P 放入空闲链表
```

### 2. G 的创建和调度
```
newproc(fn)
  └─> newG(fn)
      ├─> 生成唯一 goid
      ├─> 设置状态为 Grunnable
      └─> runqput(pp, gp, true)  // 放入 P 的 runnext

schedule()  // 调度循环
  └─> findrunnable()
      ├─> 1. runqget(pp)          // 本地队列
      ├─> 2. globrunqget(pp, 1)   // 全局队列
      └─> 3. runqsteal(pp)        // 工作窃取
  └─> execute(gp)
      ├─> gp.m = mp
      ├─> setg(gp)                // 切换到用户 G
      ├─> gp.fn()                 // 执行用户函数
      └─> goexit()
          ├─> gp.status = Gdead
          ├─> setg(mp.g0)         // 切换回 g0
          └─> schedule()          // 继续调度
```

### 3. 工作窃取
```
runqsteal(pp)
  └─> 遍历所有 P
      └─> runqstealFromP(pp, p2)
          ├─> 计算窃取数量 n = (p2.runqtail - p2.runqhead) / 2
          ├─> 至少窃取 1 个（如果有）
          ├─> 返回第一个 G
          └─> 其余 G 放入 pp 的本地队列
```

## 测试覆盖

### Phase 1 测试 (7 个)
- [x] TestCreateG - G 的创建
- [x] TestGoidUnique - goid 唯一性
- [x] TestGoidConcurrent - 并发创建 G
- [x] TestExecuteG - G 的执行
- [x] TestGetgSetg - getg/setg 切换
- [x] TestInitG0M0 - g0 和 m0 初始化
- [x] TestSchedinit - 调度器初始化

### Phase 2 测试 (5 个)
- [x] TestRunqPutGet - 基本队列操作
- [x] TestRunqRunnext - runnext 优先级
- [x] TestRunqFull - 队列满时分流
- [x] TestGlobalQueue - 全局队列操作
- [x] TestRunqempty - 队列空检查

### Phase 3 测试 (8 个)
- [x] TestProcresize - P 的创建和分配
- [x] TestNewproc - G 的创建和入队
- [x] TestScheduleBasic - 基础调度
- [x] TestFindrunnable - G 查找逻辑
- [x] TestExecuteAndGoexit - G 执行和退出
- [x] TestScheduleMultipleGs - 多 G 调度
- [x] TestProcresizeExpand - P 数量扩展
- [x] TestProcresizeShrink - P 数量缩减

### Phase 5 测试 (6 个)
- [x] TestRunqsteal - 基本工作窃取
- [x] TestRunqstealFromP - 从指定 P 窃取
- [x] TestRunqstealEmpty - 空队列不窃取
- [x] TestRunqstealOneG - 单个 G 的窃取
- [x] TestFindrunableWithSteal - 通过工作窃取查找
- [x] TestWorkStealingBalance - 负载均衡

**总计: 27/27 测试全部通过 ✅**

## 与 Go Runtime 的对比

### 相同点
1. **核心数据结构**: G、M、P、Sched 结构与 Go runtime 一致
2. **调度流程**: schedule → findrunnable → execute → goexit
3. **队列设计**: 本地队列（256）+ 全局队列
4. **优化机制**:
   - runnext 优先级
   - 工作窃取负载均衡
   - 队列满时自动分流
5. **函数命名**: 与 Go runtime 保持一致

### 简化点
1. **getg() 实现**: 使用全局变量而非 TLS（线程局部存储）
2. **无系统调用**: 不处理 entersyscall/exitsyscall
3. **无抢占**: 没有实现协作式或异步抢占
4. **无 GC 交互**: 不涉及垃圾回收相关逻辑
5. **无网络轮询器**: netpoller 未实现
6. **单 M**: 简化为单线程调度（M 实际上不启动独立的 OS 线程）

## 代码文件

```
src/gmp/
├── types.go              # 数据结构定义
├── proc_rem.go           # 调度器实现
├── types_test.go         # Phase 1 测试
├── queue_test.go         # Phase 2 测试
├── scheduler_test.go     # Phase 3 测试
├── worksteal_test.go     # Phase 5 测试
└── README.md            # 本文档
```

## 使用示例

```go
package main

import "go-rem/gmp"

func main() {
    // 1. 初始化调度器
    gmp.schedinit()

    // 2. 创建新的 G
    gmp.newproc(func() {
        println("Hello from G 1")
    })

    gmp.newproc(func() {
        println("Hello from G 2")
    })

    // 3. 开始调度
    gmp.schedule()
}
```

## 运行测试

```bash
# 运行所有测试
go test -v

# 运行特定阶段的测试
go test -v -run="TestRunq"      # Phase 2 队列测试
go test -v -run="TestSchedule"  # Phase 3 调度测试
go test -v -run="TestSteal"     # Phase 5 工作窃取测试

# 查看测试覆盖率
go test -cover
```

## 学习路径

1. **阅读文档** (src/docs/)
   - 00-getting-started/architecture.md - 整体架构
   - 02-deep-dive/why-schedinit-needs-getg.md - 为什么需要 getg
   - 03-implementation-guide/implementing-getg.md - getg 实现指南

2. **阅读代码**
   - types.go - 理解数据结构
   - proc_rem.go - 理解调度逻辑

3. **运行测试**
   - 从 Phase 1 到 Phase 5 逐步理解
   - 修改测试观察行为变化

4. **对比 Go Runtime**
   - 阅读 $GOROOT/src/runtime/proc.go
   - 对比函数实现的异同

## 关键收获

1. **G-M-P 模型的本质**: P 是调度上下文，M 是执行单元，G 是任务
2. **多级队列的好处**: 本地队列减少锁竞争，全局队列保证公平性
3. **工作窃取的重要性**: 自动负载均衡，充分利用资源
4. **runnext 的作用**: 提升缓存局部性，优化性能
5. **状态机设计**: G 的状态转换清晰，便于理解生命周期

## 参考资料

- Go Runtime 源码: $GOROOT/src/runtime/
  - runtime2.go - 数据结构定义
  - proc.go - 调度器实现
  - asm_amd64.s - getg() 汇编实现

- 官方文档:
  - [The Go scheduler](https://golang.org/s/go11sched)
  - [Scalable Go Scheduler Design Doc](https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw)

## 贡献

这是一个学习项目，欢迎提出改进建议！

## License

MIT License - 用于学习和教育目的

---

**实现完成时间**: 2025-12-30
**总代码行数**: ~900 行
**测试覆盖**: 27 个测试用例全部通过 ✅
