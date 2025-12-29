# 为什么 schedinit() 需要先获取当前的 g？

## 问题

在 Go runtime 源码中，`schedinit()` 函数开头通常会调用 `getg()` 获取当前的 g：

```go
func schedinit() {
    _g_ := getg()  // ← 为什么需要这一行？
    // ...
}
```

这是为什么？

---

## 答案：因为 schedinit() 本身也在一个 g 上运行

### 核心概念：g0（系统 goroutine）

在 Go runtime 中，**所有代码都必须在某个 g 上运行**，包括调度器初始化代码本身。

**关键事实**：
- `schedinit()` 是在 **g0** 上运行的
- g0 是每个 M 的**系统 goroutine**，专门用于执行调度代码

---

## 详细解释

### 1. Go 程序启动时的 g 层次结构

```
程序启动
  ↓
rt0_go (汇编入口)
  ↓ 创建 m0 和 g0
  ↓ 设置 m0.g0 = &g0
  ↓ 切换到 g0 栈
  ↓
schedinit()  ← 在 g0 上运行！
  ↓
mstart()
  ↓
schedule()
  ↓
执行 runtime.main (在新的 g 上)
```

**对应源码**：`runtime/asm_amd64.s` (rt0_go 函数)

```assembly
// 创建 m0 和 g0
LEAQ    runtime·m0+m_tls(SB), DI
CALL    runtime·settls(SB)

// 获取 TLS 中的 g
get_tls(BX)
LEAQ    runtime·g0(SB), CX
MOVQ    CX, g(BX)       // 设置当前 g = g0

LEAQ    runtime·m0(SB), AX
MOVQ    CX, m_g0(AX)    // m0.g0 = g0
MOVQ    AX, g_m(CX)     // g0.m = m0

// 调用 schedinit
CALL    runtime·schedinit(SB)  // 此时运行在 g0 上
```

### 2. 为什么 schedinit() 需要 getg()？

**runtime/proc.go** 中的实际代码：

```go
func schedinit() {
    _g_ := getg()  // 获取当前的 g (即 g0)

    // 理由 1: 验证我们在 g0 上运行
    if _g_ != g0 {
        throw("schedinit not on g0")
    }

    // 理由 2: 访问当前 g 的字段
    sched.maxmcount = 10000
    _g_.m.mcache = allocmcache()  // ← 需要通过 _g_ 访问 m

    // 理由 3: 设置当前 g 的状态
    _g_.stackguard0 = _g_.stack.lo + _StackGuard
    _g_.stackguard1 = ^uintptr(0)

    // 理由 4: 初始化 P 时需要当前 m 的信息
    procresize(procs)
    _g_.m.p.ptr().status = _Prunning  // ← 通过 _g_ 访问 p

    // ...
}
```

### 3. getg() 返回的是什么？

**runtime/stubs.go** (声明)：

```go
// getg 返回当前的 g
func getg() *g
```

**runtime/asm_amd64.s** (汇编实现)：

```assembly
// func getg() *g
TEXT runtime·getg(SB),NOSPLIT,$0-8
    get_tls(CX)          // 获取 TLS (Thread Local Storage)
    MOVQ    g(CX), AX    // 从 TLS 中读取 g
    MOVQ    AX, ret+0(FP)
    RET
```

**关键点**：
- 每个 M（OS 线程）都有自己的 TLS
- TLS 中存储了当前正在运行的 g 的指针
- `getg()` 从 TLS 中读取这个指针

---

## 为什么需要获取当前 g 的具体原因

### 理由 1: 验证执行环境

```go
func schedinit() {
    _g_ := getg()

    // 确保我们在 g0 上运行
    if _g_ != getg().m.g0 {
        throw("schedinit must run on g0")
    }
}
```

**为什么重要？**
- `schedinit()` 会修改全局调度器状态
- 如果在普通的 g 上运行，可能导致栈空间不足或其他问题
- g0 有特殊的大栈（通常 8KB+），适合执行系统代码

### 理由 2: 访问当前 M 和 P

```go
func schedinit() {
    _g_ := getg()

    // 通过 g 访问绑定的 m
    _g_.m.mcache = allocmcache()

    // 通过 m 访问绑定的 p
    _g_.m.p.ptr().status = _Prunning
}
```

**为什么不直接用全局的 m0？**
- 虽然初始化时确实是 m0，但代码更通用
- 相同的模式在其他地方也适用（如 procresize）
- 保持代码一致性

### 理由 3: 设置栈保护

```go
func schedinit() {
    _g_ := getg()

    // 设置栈溢出检测边界
    _g_.stackguard0 = _g_.stack.lo + _StackGuard
    _g_.stackguard1 = ^uintptr(0)  // 禁用抢占
}
```

**为什么需要？**
- 即使是 g0，也需要栈保护
- 防止栈溢出导致内存损坏

### 理由 4: 调试和诊断

```go
func schedinit() {
    _g_ := getg()

    if raceenabled {
        _g_.racectx = raceinit()  // 初始化 race detector
    }

    if trace.enabled {
        traceGStart(_g_)  // 记录 trace 事件
    }
}
```

---

## g0 vs 普通 g 的区别

| 特性 | g0（系统 g） | 普通 g（用户 g） |
|------|-------------|----------------|
| **用途** | 执行调度代码、GC、栈增长 | 执行用户代码 |
| **栈大小** | 固定大小（通常 8KB+） | 动态增长（初始 2KB） |
| **栈位置** | M 的系统栈 | 堆上分配 |
| **生命周期** | 随 M 创建/销毁 | 动态创建/销毁 |
| **ID** | 通常为 0 | 递增的 goid |
| **可以执行的代码** | 调度器、runtime 内部函数 | 用户代码、标准库 |

---

## 完整的启动流程

```go
// runtime/asm_amd64.s
TEXT runtime·rt0_go(SB)
    // 1. 创建 m0
    LEAQ runtime·m0(SB), AX

    // 2. 创建 g0
    LEAQ runtime·g0(SB), CX

    // 3. 关联 m0 和 g0
    MOVQ CX, m_g0(AX)   // m0.g0 = &g0
    MOVQ AX, g_m(CX)    // g0.m = &m0

    // 4. 设置 TLS，让 getg() 返回 g0
    get_tls(BX)
    MOVQ CX, g(BX)      // TLS.g = g0

    // 5. 调用 schedinit (在 g0 上运行)
    CALL runtime·schedinit(SB)

    // 6. 创建 main goroutine
    PUSHQ $runtime·mainPC(SB)
    CALL runtime·newproc(SB)

    // 7. 启动调度器
    CALL runtime·mstart(SB)
```

---

## 你的简化版 GMP 应该怎么做？

### 选项 1: 完全模仿 Go runtime

```go
// 定义 g0
var g0 G

// 定义 m0
var m0 M

func schedinit() {
    _g_ := getg()  // 需要实现 getg()

    if _g_.goid != 0 {
        panic("schedinit must run on g0")
    }

    // 初始化逻辑...
}

// 需要实现 TLS 和 getg()
func getg() *G {
    // 从 TLS 读取当前 g
    // 这在 Go 中需要用 unsafe 或 cgo
}
```

**缺点**：
- ❌ 需要实现 TLS（复杂）
- ❌ 需要汇编代码
- ❌ 平台相关

### 选项 2: 简化实现（推荐）

```go
// 全局的 g0 和 m0
var (
    g0 = &G{goid: 0}
    m0 = &M{id: 0, curg: g0}
)

func schedinit() {
    // 简化版：假设总是在 m0/g0 上初始化

    // 设置关联
    m0.g0 = g0
    g0.m = m0

    // 创建 P
    procs := runtime.NumCPU()
    for i := 0; i < procs; i++ {
        p := newP(int32(i))
        sched.allp = append(sched.allp, p)
    }

    // m0 绑定 p0
    m0.p = sched.allp[0]
    sched.allp[0].m = m0
    sched.allp[0].status = Prunning
}
```

**优点**：
- ✅ 简单易懂
- ✅ 专注于调度逻辑
- ✅ 不需要平台相关代码

---

## 核心启示

### 为什么 schedinit() 需要 getg()？

**三个层次的理解**：

1. **表面原因**：需要访问当前 g 的字段（m, stackguard 等）

2. **深层原因**：Go runtime 的设计哲学 —— **所有代码都在 g 上运行**
   - 即使是调度器代码本身
   - 这保证了运行时的一致性和可预测性

3. **本质原因**：**g 是 Go 并发模型的基本执行单元**
   - 提供了栈空间
   - 提供了上下文信息（m, p 的关联）
   - 提供了调试和追踪能力

---

## 总结

| 问题 | 答案 |
|------|------|
| **schedinit 在哪个 g 上运行？** | g0（系统 goroutine） |
| **为什么需要 getg()？** | 访问当前 g 的字段，验证执行环境 |
| **getg() 怎么实现的？** | 从 TLS 读取当前 g 指针 |
| **简化版需要实现吗？** | 不需要，用全局 g0/m0 即可 |

**关键概念**：
- 所有代码都在 g 上运行（包括调度器代码）
- g0 是特殊的系统 g，用于执行 runtime 内部逻辑
- getg() 通过 TLS 获取当前 g 指针

---

**参考资料**：
- Go 源码：`runtime/proc.go` (schedinit 实现)
- Go 源码：`runtime/asm_amd64.s` (rt0_go, getg 实现)
- Go 源码：`runtime/runtime2.go` (g, m 结构定义)
