# getg() 函数实现指南

## 文档目的

本文档提供在简化版 GMP 中实现 `getg()` 函数的**实践指南**，让你的代码结构与 Go runtime 保持一致，同时保持简单易懂。

---

## 为什么需要 getg()？

在 Go runtime 中，几乎所有函数都需要获取"当前的 g"：

```go
func schedinit() {
    gp := getg()  // 获取当前 g
    // ...
}

func mcommoninit(mp *m, id int64) {
    gp := getg()  // 获取当前 g
    // ...
}
```

**理由**：
- 所有代码都在某个 g 上运行（包括调度器代码）
- 需要访问当前 g 的字段（m, stackguard 等）
- 需要验证执行环境（是否在 g0 上）

详细原理见：[why-schedinit-needs-getg.md](../02-deep-dive/why-schedinit-needs-getg.md)

---

## Go 源码中的 getg() 和 setg()

### getg() 在 Go runtime 中的实现

**声明位置**：`runtime/stubs.go:218`
```go
func getg() *g
```

**实现位置**：`runtime/asm_amd64.s`（汇编）
```asm
TEXT runtime·getg(SB),NOSPLIT,$0-8
    get_tls(CX)
    MOVQ    g(CX), AX
    MOVQ    AX, ret+0(FP)
    RET
```

### setg() 在 Go runtime 中的实现

**是的，Go 源码中也有 setg() 函数！**

**声明位置**：`runtime/stubs.go:219`
```go
func setg(gg *g)
```

**实现位置**：`runtime/asm_amd64.s`（汇编）
```asm
TEXT runtime·setg(SB), NOSPLIT, $0-8
    MOVQ    gg+0(FP), BX    // 从参数中获取新的 g
    get_tls(CX)              // 获取 TLS 地址
    MOVQ    BX, g(CX)        // 将新的 g 写入 TLS
    RET
```

### setg() 的使用场景

在 Go runtime 中，`setg()` 主要在以下场景使用：

**1. M 启动时设置 g0**（`proc.go:2320`）
```go
func mstart1() {
    // ...
    setg(mp.g0)  // 设置当前 M 的 g0
    // ...
}
```

**2. M 退出时清空当前 g**（`proc.go:2533`）
```go
func mexit(osStack bool) {
    // ...
    setg(nil)  // 清空当前 g
}
```

**3. 信号处理时切换 g**（`signal_unix.go:437,476,491`）
```go
func sighandler(...) {
    gp := getg()
    // ...
    setg(gp)              // 恢复用户 g
    // ...
    setg(gp.m.gsignal)    // 切换到信号处理 g
    // ...
}
```

### 为什么需要 setg()？

在真实的 Go runtime 中：

1. **M 绑定到 OS 线程**：每个 M 是一个真正的 OS 线程
2. **TLS 是线程局部的**：每个线程有独立的 TLS 存储
3. **需要更新 TLS**：当切换执行的 G 时，必须更新 TLS 中的 g 指针

**关键时刻**：
- M 启动：`setg(g0)` - 初始化线程的 TLS
- 执行用户 G：在 `gogo()` 中隐式更新（汇编中直接操作 TLS）
- 系统调用：`entersyscall()` / `exitsyscall()` 时可能切换 g
- 信号处理：在用户 g 和信号 g 之间切换

### 我们的简化实现

在简化版 GMP 中：
- **Phase 1**：用全局变量 `currentG` 模拟 TLS
- **setg()** 变成简单的赋值：`currentG = gp`
- **getg()** 变成简单的读取：`return currentG`

这保持了与 Go runtime 相同的**接口**和**调用方式**，但实现大幅简化，便于学习理解。

---

## 渐进式实现方案

### 🎯 推荐顺序

1. **Phase 1**: 简化版（全局变量）← **从这里开始**
2. **Phase 2**: 中级版（模拟 TLS）
3. **Phase 3**: 完整版（真实 TLS）

---

## Phase 1: 简化版（推荐）

### 适用场景

- ✅ 学习和理解 GMP 模型
- ✅ 单线程或简单的并发场景
- ✅ 快速原型开发

### 实现代码

在 `proc_rem.go` 中添加：

```go
package gmp

// 当前正在运行的 g（简化版）
var currentG *g = g0

// getg 返回当前的 g
// 在真实的 Go runtime 中，这是通过 TLS 实现的
// 这里用全局变量简化
func getg() *g {
    return currentG
}

// setg 设置当前 g
// 在 g 切换时调用
func setg(gp *g) {
    currentG = gp
}
```

### 使用示例

现在你可以像 Go runtime 一样写代码：

```go
func schedinit() {
    gp := getg()  // ← 和 Go runtime 一致

    // 验证在 g0 上运行
    if gp != g0 {
        panic("schedinit must run on g0")
    }

    sched.maxmcount = 10000
    mcommoninit(gp.m, -1)
}

func mcommoninit(mp *m, id int64) {
    gp := getg()  // 获取当前 g

    // 通过 g 访问 m
    if gp.m != mp {
        panic("mcommoninit: wrong m")
    }

    mp.id = id
}

func ExecuteG(newg *g) {
    oldg := getg()      // 保存当前 g
    setg(newg)          // 切换到新 g

    newg.status = _Grunning
    if newg.fn != nil {
        newg.fn()
    }
    newg.status = _Gdead

    setg(oldg)          // 恢复原来的 g
}
```

### 优点与缺点

**优点**：
- ✅ 代码结构与 Go runtime 完全一致
- ✅ 实现简单，无外部依赖
- ✅ 易于调试和理解
- ✅ 为后续升级留下接口

**缺点**：
- ❌ 只能有一个"当前 g"（单线程）
- ❌ 不支持真正的并发 M

**适合吗？**

对于学习 GMP 模型，**完全够用**！你可以实现完整的调度逻辑，理解 G、M、P 的协作，只是不能真正并发运行多个 M。

---

## Phase 2: 模拟 TLS（进阶）

### 适用场景

- ✅ 需要支持多个 M 并发调度
- ✅ 理解 TLS 的概念和作用
- ✅ 不想依赖 CGO

### 实现代码

```go
package gmp

import (
    "sync"
    "runtime"
    "bytes"
    "strconv"
)

// 模拟 TLS：goroutine ID -> g 的映射
var gls sync.Map

// getCurrentGoroutineID 获取当前 goroutine 的 ID
func getCurrentGoroutineID() int64 {
    var buf [64]byte
    n := runtime.Stack(buf[:], false)
    // 解析 "goroutine 123 [running]:"
    idField := bytes.Fields(buf[:n])[1]
    id, _ := strconv.ParseInt(string(idField), 10, 64)
    return id
}

// getg 返回当前 goroutine 对应的 g
func getg() *g {
    gid := getCurrentGoroutineID()
    if gp, ok := gls.Load(gid); ok {
        return gp.(*g)
    }
    // 默认返回 g0
    return g0
}

// setg 设置当前 goroutine 对应的 g
func setg(gp *g) {
    gid := getCurrentGoroutineID()
    gls.Store(gid, gp)
}
```

### 使用方式

与 Phase 1 完全相同，但支持多 goroutine：

```go
func mstart(mp *m) {
    // 每个 M 在独立的 goroutine 中运行
    go func() {
        setg(mp.g0)  // 设置当前 goroutine 的 g

        schedule(mp) // 调度循环
    }()
}

func schedule(mp *m) {
    for {
        gp := getg()  // 获取当前 goroutine 的 g

        // 找到下一个要执行的 G
        nextg := findrunnable(mp.p)

        if nextg != nil {
            execute(nextg)
        }
    }
}
```

### 优点与缺点

**优点**：
- ✅ 支持真正的多 M 并发
- ✅ 每个 goroutine 有独立的 g
- ✅ 纯 Go 实现，跨平台

**缺点**：
- ❌ 性能比真实 TLS 差（sync.Map 查找）
- ❌ 依赖 runtime.Stack 解析（有点 hack）
- ❌ 代码复杂度增加

**何时使用？**

当你完成了 Phase 1-4（多 P 多 M），需要真正的并发调度时。

---

## Phase 3: 真实 TLS（完整版）

### 适用场景

- ✅ 完全模拟 Go runtime
- ✅ 需要最佳性能
- ✅ 深入理解 TLS 机制

### 实现代码

```go
// getg_tls.go
// +build cgo

package gmp

/*
#include <pthread.h>

static pthread_key_t g_key;
static pthread_once_t g_key_once = PTHREAD_ONCE_INIT;

static void make_g_key() {
    pthread_key_create(&g_key, NULL);
}

void set_current_g(void* g) {
    pthread_once(&g_key_once, make_g_key);
    pthread_setspecific(g_key, g);
}

void* get_current_g() {
    pthread_once(&g_key_once, make_g_key);
    return pthread_getspecific(g_key);
}
*/
import "C"
import "unsafe"

func getg() *g {
    gp := (*g)(C.get_current_g())
    if gp == nil {
        return g0
    }
    return gp
}

func setg(gp *g) {
    C.set_current_g(unsafe.Pointer(gp))
}
```

### 优点与缺点

**优点**：
- ✅ 真实的 TLS 实现
- ✅ 性能接近 Go runtime（纳秒级）
- ✅ 完全模拟 Go 的行为

**缺点**：
- ❌ 需要 CGO（构建复杂）
- ❌ 平台相关（需要 pthread）
- ❌ 调试困难

**何时使用？**

当你完成了整个 GMP 实现，需要进行性能对比和深入研究时。

---

## 实践建议

### 🚀 推荐路径

```
开始学习 GMP
    ↓
实现 Phase 1（全局变量）
    ↓
完成 Phase 1-2 测试
    ↓
实现完整的单 M 调度器
    ↓
（可选）升级到 Phase 2
    ↓
实现多 M 多 P 并发调度
    ↓
（可选）升级到 Phase 3
    ↓
性能对比和优化
```

### 代码组织

建议的文件结构：

```
src/gmp/
├── types.go          # 数据结构定义
├── proc_rem.go       # 调度器实现
├── getg_simple.go    # Phase 1: 简化版 getg
├── getg_gls.go       # Phase 2: 模拟 TLS (可选)
├── getg_tls.go       # Phase 3: 真实 TLS (可选)
└── *_test.go         # 测试文件
```

使用 build tags 选择实现：

```go
// getg_simple.go
// +build !gls,!tls

package gmp
var currentG *g = g0
func getg() *g { return currentG }
```

```go
// getg_gls.go
// +build gls

package gmp
// ... Phase 2 实现
```

### 测试验证

无论使用哪个 Phase，测试代码都相同：

```go
func TestGetG(t *testing.T) {
    // 初始应该是 g0
    if getg() != g0 {
        t.Error("初始 g 应该是 g0")
    }

    // 切换 g
    newg := newG(func() {})
    setg(newg)

    if getg() != newg {
        t.Error("getg() 返回错误的 g")
    }

    // 恢复
    setg(g0)
}
```

---

## 常见问题

### Q: 为什么不直接用参数传递 g？

**A**: 虽然可以，但会让代码很繁琐：

```go
// ❌ 不好：到处传递 g
func schedule(gp *g) {
    findrunnable(gp)
    execute(gp, nextG)
}

// ✅ 好：用 getg() 获取
func schedule() {
    gp := getg()
    findrunnable()
    execute(nextG)
}
```

### Q: Phase 1 能完成所有学习目标吗？

**A**: 能！你可以：
- ✅ 理解 G、M、P 的关系
- ✅ 实现调度循环和队列操作
- ✅ 理解工作窃取算法
- ✅ 完成所有测试

只是不能真正"并发"运行多个 M，但这不影响理解调度逻辑。

### Q: 何时升级到 Phase 2 或 3？

**A**: 建议顺序：
1. 先用 Phase 1 完成所有功能
2. 所有测试通过
3. 深入理解了调度原理
4. 想要体验真正的并发调度
5. 再考虑升级

---

## 参考资料

### 相关文档

- [how-getg-works.md](../02-deep-dive/how-getg-works.md) - getg() 原理详解
- [why-schedinit-needs-getg.md](../02-deep-dive/why-schedinit-needs-getg.md) - 为什么需要 getg
- [architecture.md](../00-getting-started/architecture.md) - GMP 架构设计

### Go Runtime 源码

- `runtime/stubs.go:218` - getg() 声明
- `runtime/stubs.go:219` - setg() 声明
- `runtime/asm_amd64.s` - getg() 和 setg() 汇编实现
- `runtime/proc.go:2320` - setg() 在 M 启动时的使用
- `runtime/proc.go:2533` - setg() 在 M 退出时的使用
- `runtime/signal_unix.go` - setg() 在信号处理中的使用

---

## 快速开始

**立即开始使用 Phase 1**：

1. 在 `proc_rem.go` 中添加：
   ```go
   var currentG *g = g0
   func getg() *g { return currentG }
   func setg(gp *g) { currentG = gp }
   ```

2. 修改 `schedinit()`：
   ```go
   func schedinit() {
       gp := getg()
       // ...
   }
   ```

3. 运行测试：
   ```bash
   go test -v
   ```

就这么简单！祝实现顺利 🎉
