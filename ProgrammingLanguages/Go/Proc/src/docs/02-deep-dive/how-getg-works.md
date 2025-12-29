# getg() 函数的实现原理

## 问题

在 Go runtime 源码中经常看到：

```go
_g_ := getg()
```

这个 `getg()` 函数是如何实现的？它如何知道当前运行的是哪个 g？

---

## 核心答案

**getg() 通过 TLS (Thread Local Storage) 获取当前线程正在运行的 g 的指针**

### TLS 是什么？

**TLS (Thread Local Storage)** = 线程本地存储
- 每个 OS 线程都有自己的 TLS 区域
- 存储该线程的私有数据
- 访问速度极快（CPU 寄存器级别）

---

## Go Runtime 中的实现

### 1. 函数声明

**runtime/stubs.go**:

```go
// getg returns the pointer to the current g.
// The compiler rewrites calls to this function into instructions
// that fetch the g directly (from TLS or from the dedicated register).
func getg() *g
```

**关键点**：
- 这只是一个声明，没有 Go 代码实现
- 实际实现在汇编代码中
- 编译器会特殊处理这个函数

### 2. 汇编实现

**runtime/asm_amd64.s** (x86-64 架构):

```assembly
// func getg() *g
TEXT runtime·getg(SB),NOSPLIT,$0-8
    get_tls(CX)          // 1. 获取 TLS 基址到 CX 寄存器
    MOVQ    g(CX), AX    // 2. 从 TLS 中读取 g 指针到 AX
    MOVQ    AX, ret+0(FP)// 3. 将 AX 的值作为返回值
    RET
```

**详细解释**：

```
get_tls(CX)
    ↓
获取当前线程的 TLS 基址
TLS 是一块内存区域，每个线程独有

g(CX)
    ↓
TLS + 偏移量 (g 在 TLS 中的位置)

MOVQ g(CX), AX
    ↓
从 [TLS基址 + g偏移] 读取 g 指针
```

### 3. get_tls 宏定义

**runtime/go_tls.h** (x86-64):

```assembly
#ifdef GOOS_darwin
#define    get_tls(r)    MOVQ TLS, r
#else
#define    get_tls(r)    MOVQ 0(FS), r
#endif
```

**平台差异**：
- **Linux/BSD**: 使用 FS 段寄存器
- **macOS**: 使用 TLS 伪寄存器
- **Windows**: 使用 GS 段寄存器

---

## TLS 中存储的是什么？

### TLS 布局

在 Go runtime 中，每个 M（OS 线程）的 TLS 主要存储：

```
TLS 内存布局:
┌──────────────────┐
│ g 指针           │ ← 当前运行的 g
├──────────────────┤
│ m 指针 (可选)    │ ← 当前的 m (某些架构)
└──────────────────┘
```

**关键代码** (runtime/proc.go):

```go
// save 保存当前 g 到 TLS
func save(pc, sp uintptr) {
    _g_ := getg()
    // 当切换 g 时，会更新 TLS 中的 g 指针
}
```

---

## 完整的 g 切换流程

### 场景：从 g1 切换到 g2

```
1. 当前在 g1 上运行
   TLS.g = g1

2. 调度器决定切换到 g2
   gogo(&g2.sched)  // 汇编函数

3. gogo 做的事情：
   a. 保存 g1 的寄存器状态到 g1.sched
   b. 从 g2.sched 恢复寄存器
   c. 更新 TLS: 写入 g2 指针
   d. 跳转到 g2 的 PC

4. 现在 getg() 返回 g2
   TLS.g = g2
```

### gogo 的汇编实现

**runtime/asm_amd64.s**:

```assembly
// func gogo(buf *gobuf)
// restore state from Gobuf; longjmp
TEXT runtime·gogo(SB), NOSPLIT, $0-8
    MOVQ    buf+0(FP), BX        // 1. 获取 gobuf 指针
    MOVQ    gobuf_g(BX), DX      // 2. 取出要切换到的 g

    get_tls(CX)                   // 3. 获取 TLS
    MOVQ    DX, g(CX)            // 4. 更新 TLS.g = 新的 g ★

    MOVQ    gobuf_sp(BX), SP     // 5. 恢复栈指针
    MOVQ    gobuf_pc(BX), BX     // 6. 恢复程序计数器
    JMP     BX                    // 7. 跳转到新 g 的代码
```

**关键步骤 4**：`MOVQ DX, g(CX)` —— 更新 TLS 中的 g 指针

---

## 为什么需要 TLS？

### 原因 1: 性能

**访问速度对比**：

```
全局变量:          ~10 ns   (需要内存访问)
TLS:               ~1 ns    (CPU 寄存器/缓存)
函数参数传递:       ~5 ns    (需要栈操作)
```

**getg() 被大量调用**：
- 每次调度都要调用
- 每次系统调用都要调用
- 每次栈检查都要调用

### 原因 2: 线程安全

每个线程有自己的 TLS，天然线程安全：

```go
// 错误的方式（全局变量，不安全）
var currentG *G

func getg() *G {
    return currentG  // 多线程会冲突！
}

// 正确的方式（TLS，线程安全）
func getg() *G {
    return TLS.g  // 每个线程独立
}
```

### 原因 3: 简化代码

不需要在函数间传递 g：

```go
// 没有 getg() 时（繁琐）
func scheduleWithG(_g_ *g) {
    findrunnable(_g_)
    execute(_g_, nextG)
}

// 有 getg() 时（简洁）
func schedule() {
    _g_ := getg()
    findrunnable()
    execute(nextG)
}
```

---

## 不同架构的实现

### x86-64 (amd64)

```assembly
// runtime/asm_amd64.s
TEXT runtime·getg(SB),NOSPLIT,$0-8
    MOVQ    (TLS), AX       // 从 FS 段读取
    MOVQ    AX, ret+0(FP)
    RET
```

### ARM64

```assembly
// runtime/asm_arm64.s
TEXT runtime·getg(SB),NOSPLIT,$0-8
    MOVD    g, R0           // ARM64 有专用的 g 寄存器！
    MOVD    R0, ret+0(FP)
    RET
```

**ARM64 的优化**：
- 直接用寄存器存储 g
- 比 TLS 更快

### RISC-V

```assembly
// runtime/asm_riscv64.s
TEXT runtime·getg(SB),NOSPLIT,$0-8
    MOV    g, A0            // RISC-V 也用专用寄存器
    MOV    A0, ret+0(FP)
    RET
```

---

## 编译器的特殊处理

### 内联优化

编译器会把 `getg()` 调用直接替换为读取 TLS 的指令：

**Go 代码**:
```go
func foo() {
    _g_ := getg()
    _g_.m.locks++
}
```

**编译后的汇编** (简化):
```assembly
foo:
    MOVQ (TLS), AX      // 直接内联，不是函数调用
    ADDQ $1, m_locks(AX)
```

### 栈检查

每个函数开头都会检查栈空间：

```assembly
functionStart:
    MOVQ (TLS), R14          // getg() 内联
    CMPQ SP, g_stackguard0(R14)
    JBE  stackOverflow       // 栈不够，需要增长
    // ... 函数体
```

---

## TLS 的初始化

### M 创建时设置 TLS

**runtime/proc.go**:

```go
func newosproc(mp *m) {
    // 创建新的 OS 线程
    var oset sigset
    sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
    ret := clone(
        cloneFlags,
        stk,
        unsafe.Pointer(mp),
        unsafe.Pointer(&mp.procid),
        unsafe.Pointer(&mp.gsignal.stack.hi))

    // 设置 TLS
    settls()  // ← 关键步骤
}
```

**runtime/sys_linux_amd64.s**:

```assembly
// func settls()
TEXT runtime·settls(SB),NOSPLIT,$32
    ADDQ $16, DI            // 计算 TLS 地址
    MOVQ DI, SI
    MOVQ $0x1002, DI        // arch_prctl 系统调用号
    MOVQ $158, AX           // SYS_arch_prctl
    SYSCALL                 // 设置 FS 寄存器指向 TLS
    RET
```

**系统调用**：
- Linux: `arch_prctl(ARCH_SET_FS, tls_addr)`
- macOS: `thread_fast_set_cthread_self64(tls_addr)`
- Windows: 设置 GS 寄存器

---

## 在你的简化版 GMP 中如何处理？

### 选项 1: 不实现 getg()（推荐）

**原因**：
- 需要汇编代码（复杂）
- 需要 TLS 设置（平台相关）
- 不影响学习调度逻辑

**替代方案**：

```go
// 全局变量（简化版可接受）
var g0 = &G{goid: 0}
var m0 = &M{id: 0, curg: g0}

func schedinit() {
    // 直接使用全局变量，不需要 getg()
    g0.m = m0
    m0.g0 = g0

    // 创建 P
    for i := 0; i < numP; i++ {
        p := newP(int32(i))
        sched.allp = append(sched.allp, p)
    }

    // m0 绑定 p0
    m0.p = sched.allp[0]
}
```

### 选项 2: 模拟 TLS（学习用）

如果你想理解 TLS 机制，可以用 Go 的 `sync.Map` 模拟：

```go
// 模拟的 TLS
var tlsMap sync.Map  // map[threadID]*G

// 模拟 getg()
func getg() *G {
    tid := gettid()  // 获取当前线程 ID
    if g, ok := tlsMap.Load(tid); ok {
        return g.(*G)
    }
    return nil
}

// 设置当前 g
func setg(g *G) {
    tid := gettid()
    tlsMap.Store(tid, g)
}

// 获取线程 ID (简化版)
func gettid() int64 {
    // 可以用 goroutine ID（虽然不完美）
    // 或者用 runtime.Callers 的技巧
    return int64(getCurrentGoroutineID())
}
```

**优点**：
- ✅ 能理解 TLS 的概念
- ✅ 纯 Go 实现，可移植

**缺点**：
- ❌ 性能比真实 TLS 差很多
- ❌ 不是真正的线程级别隔离

### 选项 3: 使用 CGO（高级）

如果你想体验真实的 TLS：

```go
// tls_linux_amd64.go

/*
#include <pthread.h>

static __thread void* g_tls = NULL;

void set_g(void* g) {
    g_tls = g;
}

void* get_g() {
    return g_tls;
}
*/
import "C"
import "unsafe"

func setg(g *G) {
    C.set_g(unsafe.Pointer(g))
}

func getg() *G {
    return (*G)(C.get_g())
}
```

**优点**：
- ✅ 真实的 TLS 实现
- ✅ 性能接近 Go runtime

**缺点**：
- ❌ 需要 CGO（复杂化项目）
- ❌ 平台相关

---

## 核心概念总结

### getg() 的实现层次

```
高层 (Go 代码):
    _g_ := getg()

中层 (编译器):
    内联优化，直接生成读取 TLS 的指令

低层 (汇编):
    get_tls(CX)
    MOVQ g(CX), AX

底层 (CPU):
    读取 FS/GS 段寄存器指向的内存
```

### 为什么 getg() 如此重要？

1. **频繁调用** - 几乎每个 runtime 函数都会调用
2. **性能关键** - 必须极快（纳秒级）
3. **线程安全** - 每个线程独立，无竞争
4. **简化代码** - 不需要到处传递 g 指针

### 关键数据流

```
M 创建
  ↓
settls() - 设置 FS 寄存器指向 TLS 区域
  ↓
TLS.g = g0 - 初始化 TLS 中的 g 指针
  ↓
getg() - 从 TLS 读取 g 指针
  ↓
gogo() 切换 g 时更新 TLS.g
  ↓
getg() 返回新的 g
```

---

## 对照 Go Runtime 学习要点

### 相关源码位置

| 文件 | 内容 | 行号参考 |
|------|------|---------|
| `runtime/stubs.go` | getg() 声明 | ~30 |
| `runtime/asm_amd64.s` | getg() 汇编实现 | ~22 |
| `runtime/asm_amd64.s` | gogo() 实现 | ~250 |
| `runtime/sys_linux_amd64.s` | settls() 实现 | ~650 |
| `runtime/proc.go` | TLS 使用示例 | 全文 |

### 学习建议

1. **先理解概念** - TLS 是什么，为什么需要它
2. **看汇编代码** - 理解 getg() 的实际实现
3. **追踪数据流** - 从 settls 到 getg 到 gogo
4. **简化实现** - 在你的版本中用全局变量即可

---

## 实战建议

### 对于你的简化版 GMP

**不要纠结于实现 getg()**：

```go
// 你的简化版（清晰易懂）
var (
    g0 = &G{goid: 0}
    m0 = &M{id: 0, curg: g0}
)

func schedinit() {
    // 直接用全局变量
    g0.m = m0
    m0.g0 = g0
    // ...
}

func schedule() {
    // 直接访问 m0
    _g_ := m0.curg
    // ...
}
```

**专注于理解**：
- ✅ 为什么 Go runtime 需要 getg()
- ✅ TLS 的作用和原理
- ✅ g 切换时 TLS 如何更新

**而不是**：
- ❌ 如何写汇编代码
- ❌ 如何设置 FS 寄存器
- ❌ 不同平台的 TLS 实现差异

---

## 参考资料

- Go 源码：`runtime/stubs.go`（getg 声明）
- Go 源码：`runtime/asm_*.s`（各架构实现）
- 文章：[Thread Local Storage in Go](https://dr-knz.net/go-tls.html)
- 文章：[Understanding Go's scheduler](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part1.html)

---

**总结**：`getg()` 通过 TLS 极快地获取当前 g 的指针，是 Go runtime 的基础设施。在你的简化版中，用全局变量替代即可，重点是理解调度逻辑！
