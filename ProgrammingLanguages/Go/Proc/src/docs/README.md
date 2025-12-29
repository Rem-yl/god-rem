# 简化版 GMP 实现 - 文档索引

本目录包含简化版 GMP 调度模型的设计文档和学习指南。

---

## 📚 文档列表

### 1. [FAQ.md](./FAQ.md) - 常见问题解答

记录了实现过程中遇到的设计问题和决策思考：

- **Q1: 是否需要保证 goid 的唯一性？**
  - 答案：需要（使用 atomic.AddInt64）
  - 理由：调试追踪、测试验证、实现成本低

- **Q2: goid 应该用独立全局变量还是 sched.goidgen？**
  - 答案：使用 sched.goidgen
  - 理由：集中管理、模块化、符合 Go runtime 设计

- **Q3: 为什么 P.id 是 int32，而 G.goid 是 int64？**
  - 答案：根据实际数量级选择类型
  - P 数量少（< 1000） → int32 足够且高效
  - G 可能数十亿 → int64 防止溢出
  - M 为了一致性和扩展 → int64

- **Q4: 为什么用 signed (int) 而不是 unsigned (uint)？**
  - 答案：signed 类型更安全，可以用负数表示特殊状态
  - goidgen 使用 uint64（只递增，需要更大范围）
  - goid 存储用 int64（可以用负数表示错误状态）

---

### 2. [architecture.md](./architecture.md) - 架构设计文档

完整的 GMP 模型架构设计，包括：

- **核心数据结构**: G, M, P, Sched 的详细定义
- **核心流程**: 初始化、创建 G、调度循环、工作窃取
- **渐进式实现路径**: Phase 1-5 的详细规划
- **关键算法细节**: 循环数组、runnext、窃取数量计算
- **对照 Go Runtime**: 源码位置和学习要点
- **三周学习计划**: 每天的具体任务和测试

**适合**：
- 在开始编码前了解整体架构
- 作为实现过程中的参考手册

---

### 3. tdd-guide.md - TDD 开发指南

手把手的 TDD 教程，包含完整代码示例：

- **TDD 三步循环**: Red → Green → Blue
- **实战示例**: 从零实现 G 的创建（3 轮完整循环）
- **练习任务**: 实现 P 的本地队列（含测试代码）
- **TDD 节奏**: 步子要小，频繁运行测试
- **常见问题**: 什么该测，什么不该测

**适合**：
- 没有 TDD 经验的开发者
- 需要具体代码示例的学习者

---

## 🚀 快速开始

### 1. 阅读顺序建议

如果你是**第一次接触 TDD**：
```
1. tdd-guide.md     - 学习 TDD 基础
2. architecture.md  - 了解整体架构
3. FAQ.md           - 理解设计决策
```

如果你**有 TDD 经验**：
```
1. architecture.md  - 了解整体架构
2. FAQ.md           - 理解设计决策
3. tdd-guide.md     - 快速浏览代码示例
```

### 2. 实现步骤

按照 architecture.md 中的 **Phase 1-5** 渐进式实现：

**Phase 1: 最小可运行版本**
- 实现 G 的创建和执行
- 实现 P 的本地队列
- 实现单 M 调度循环

**Phase 2: 添加全局队列**
- 实现 Sched 全局队列
- 处理本地队列溢出
- 实现全局队列检查逻辑

**Phase 3: 多 P 多 M**
- 创建多个 P
- 创建多个 M 并绑定 P
- 验证负载均衡

**Phase 4: 工作窃取**
- 实现 findrunnable
- 实现 stealWork
- 验证窃取逻辑

**Phase 5: 可观测性**
- 添加日志输出
- 添加统计信息
- 类似 GODEBUG=schedtrace 的输出

---

## 📖 学习资源

### Go Runtime 源码位置

- **数据结构**: `runtime/runtime2.go`
  - type g struct (line 407)
  - type p struct (line 629)
  - type m struct (line 536)
  - type schedt struct (line 779)

- **调度逻辑**: `runtime/proc.go`
  - schedinit() (line 720)
  - newproc() (line 4239)
  - schedule() (line 3365)
  - findrunnable() (line 2839)
  - execute() (line 2741)

### 关键常量

- GOMAXPROCS: 默认 = CPU 核心数
- P 本地队列大小: 256
- 全局队列检查频率: 每 61 次调度
- M 最大数量: 10000

---

## 🎯 学习目标

通过实现简化版 GMP，你将理解：

1. **GMP 模型的核心概念**
   - G: Goroutine，轻量级协程
   - M: Machine，OS 线程
   - P: Processor，调度上下文

2. **调度器的工作原理**
   - 本地队列优先（提高局部性）
   - 全局队列防饥饿（每 61 次检查）
   - 工作窃取实现负载均衡

3. **Go Runtime 的设计哲学**
   - 精确的类型选择（int32 vs int64）
   - 性能与实用的平衡
   - 模块化和可扩展性

4. **TDD 开发方法**
   - 先写测试再写代码
   - 小步快跑，频繁验证
   - 测试即文档

---

## 💡 常见陷阱

### 1. ID 类型选择

❌ **错误**：所有 ID 都用 int
```go
type P struct { id int }
type G struct { goid int }
```

✅ **正确**：根据实际需求选择
```go
type P struct { id int32 }   // P 数量少
type G struct { goid int64 }  // G 可能很多
type Sched struct { goidgen uint64 }  // 只递增
```

### 2. goid 管理

❌ **错误**：用独立全局变量
```go
var goidgen int64
```

✅ **正确**：集中在 sched 中
```go
type Sched struct { goidgen uint64 }
var sched Sched
```

### 3. 循环数组索引

❌ **错误**：直接用 tail 作为索引
```go
runq[tail] = g  // 会越界
```

✅ **正确**：取模
```go
runq[tail % 256] = g
```

---

## 🔧 推荐工具

### 测试工具

```bash
# 运行测试
go test -v

# 查看覆盖率
go test -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# 持续测试（自动运行）
reflex -r '\.go$' go test -v
```

### 调试工具

```bash
# 查看 Go runtime 调度信息
GODEBUG=schedtrace=1000 ./your-program

# 查看 GC 信息
GODEBUG=gctrace=1 ./your-program
```

---

## 📝 贡献指南

在实现过程中如果发现：
- 文档错误或不清楚的地方
- 更好的实现方式
- 新的设计问题

可以：
1. 在 FAQ.md 中添加新的问题和解答
2. 在 architecture.md 中补充细节
3. 在 tdd-guide.md 中添加示例

---

## 📊 项目进度跟踪

建议使用 TODO 列表跟踪进度：

- [ ] Phase 1: 最小可运行版本
  - [ ] 实现 G 的创建
  - [ ] 实现 P 的本地队列
  - [ ] 实现单 M 调度循环
  - [ ] 测试：10 个 G 顺序执行

- [ ] Phase 2: 全局队列
  - [ ] 实现全局队列
  - [ ] 实现溢出逻辑
  - [ ] 实现全局队列检查
  - [ ] 测试：300 个 G

- [ ] Phase 3: 多 P 多 M
  - [ ] 创建多个 P
  - [ ] 创建多个 M
  - [ ] 测试：负载均衡

- [ ] Phase 4: 工作窃取
  - [ ] 实现 findrunnable
  - [ ] 实现 stealWork
  - [ ] 测试：不平衡负载

- [ ] Phase 5: 可观测性
  - [ ] 添加日志
  - [ ] 添加统计
  - [ ] 测试：输出 schedtrace

---

**祝学习顺利！通过实现这个简化版 GMP，你会对 Go 调度器有深刻的理解。** 🎉
