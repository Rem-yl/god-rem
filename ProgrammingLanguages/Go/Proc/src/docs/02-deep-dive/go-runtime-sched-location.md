# Go Runtime 中 sched 变量的真实位置

## 问题

在 Go runtime 源码中搜索 `var sched` 找不到定义，这是为什么？

---

## 答案

在 Go runtime 中，`sched` 变量的定义方式**不是**简单的全局变量声明，而是通过以下几种方式之一：

### 1. Go 1.14+ 的实际情况

**在 runtime/proc.go 中**，你会找到类似这样的定义：

```go
var (
    sched      schedt
    ncpu       int32
    // ... 其他全局变量
)
```

**关键点**：
- 它在 `var ( ... )` 块中定义，不是独立的 `var sched schedt`
- 通常在文件的靠前位置，但可能被很多注释包围

### 2. 如何找到它

#### 方法 1：搜索 "sched      schedt"（注意空格）

```bash
cd $GOROOT/src/runtime
grep -n "sched.*schedt" proc.go
```

#### 方法 2：搜索 var 块

```bash
# 查找包含多个变量定义的 var 块
grep -A20 "^var (" proc.go | grep sched
```

#### 方法 3：直接查看文件

```bash
# 打开 proc.go，通常在前 100 行左右
head -150 proc.go
```

### 3. 实际位置示例（Go 1.21）

**runtime/proc.go 大约第 120 行左右**：

```go
var (
    m0           m
    g0           g
    mcache0      *mcache
    raceprocctx0 uintptr
    // ...
    sched      schedt      // ← 在这里！
    // ...
)
```

**为什么难找？**
- 它和很多其他全局变量混在一起
- 通常有大量注释
- 在 `var ( ... )` 块中，不是独立声明

### 4. 验证方法

在你的 Go 安装目录中运行：

```bash
# 找到 Go 源码目录
go env GOROOT

# 搜索 sched 的定义（带上下文）
cd $(go env GOROOT)/src/runtime
grep -n "sched" proc.go | grep -v "//" | head -20

# 或者直接查看 var 块
awk '/^var \(/,/^\)/' proc.go | grep -n sched
```

---

## 你的简化版 GMP 应该怎么做？

### 推荐做法：清晰的独立定义

虽然 Go runtime 把 sched 放在大的 `var ()` 块中，但在你的简化版中，**建议使用更清晰的方式**：

```go
// types.go
package gmp

// Sched 全局调度器
type Sched struct {
    goidgen uint64  // G ID 生成器
    mnext   int64   // M ID 生成器

    lock     sync.Mutex
    runq     []*G    // 全局运行队列
    runqsize int

    pidle  []*P
    npidle int

    midle  []*M
    nmidle int

    allp []*P
    allm []*M
}

// 全局唯一的调度器实例
var sched Sched
```

**为什么这样做？**

1. **更清晰**：一眼就能看到 sched 的定义
2. **易于查找**：搜索 `var sched` 立即找到
3. **学习友好**：结构和实例分开，便于理解
4. **测试方便**：可以轻松重置 `sched = Sched{}`

### 如果想完全模仿 Go runtime

```go
// types.go - 只定义类型
type Sched struct {
    goidgen uint64
    // ...
}

// proc.go - 在 var 块中定义实例
var (
    m0    M
    g0    G
    sched Sched  // ← 模仿 Go runtime
    allp  []*P
    allm  []*M
)
```

---

## 常见误区

### 误区 1：以为 sched 在 runtime2.go 中

❌ **错误**：`runtime2.go` 只定义 `type schedt struct`（类型）
✅ **正确**：`proc.go` 定义 `var sched schedt`（实例）

### 误区 2：以为 sched 是独立声明

❌ **错误**：期望找到 `var sched schedt`
✅ **正确**：它在 `var ( ... )` 块中，和其他变量一起

### 误区 3：以为 sched 在汇编中定义

❌ **部分错误**：早期版本可能在汇编中，但现代版本在 Go 代码中
✅ **正确**：Go 1.10+ 之后基本都在 `proc.go` 中

---

## 实战练习

### 练习 1：在你的系统上找到 sched

```bash
# 1. 找到 Go 源码目录
GOROOT=$(go env GOROOT)
echo "Go 源码在: $GOROOT"

# 2. 查看 proc.go 的前 200 行
head -200 $GOROOT/src/runtime/proc.go

# 3. 搜索 sched 的定义
grep -n "sched.*schedt\|sched      schedt" $GOROOT/src/runtime/proc.go
```

### 练习 2：对比 schedt 定义和 sched 实例

```bash
# 类型定义在 runtime2.go
grep -A10 "^type schedt struct" $GOROOT/src/runtime/runtime2.go

# 实例定义在 proc.go
grep -B5 -A5 "sched.*schedt" $GOROOT/src/runtime/proc.go
```

---

## 总结

| 问题 | 答案 |
|------|------|
| **sched 在哪定义？** | `runtime/proc.go` 中的 `var ( ... )` 块 |
| **为什么搜 `var sched` 找不到？** | 因为它在多行 var 块中，不是独立声明 |
| **schedt 类型在哪？** | `runtime/runtime2.go` 中 |
| **简化版应该怎么做？** | 建议用独立的 `var sched Sched` 更清晰 |

---

## 参考

- Go 源码：`runtime/proc.go`（sched 实例）
- Go 源码：`runtime/runtime2.go`（schedt 类型）
- Go 版本：1.21+（不同版本可能略有差异）

**建议**：在你的简化版 GMP 中，不必完全模仿 Go runtime 的代码组织方式。使用更清晰、更易理解的结构是完全可以的！
