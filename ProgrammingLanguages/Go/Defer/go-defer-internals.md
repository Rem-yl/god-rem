# Go defer 实现原理完全解析

> 从使用到源码，从面试到实战

---

## 📖 目录

```
第一部分：基础使用
├── 1. defer 基本概念
├── 2. defer 的三大特性
└── 3. defer 常见用法

第二部分：实现原理
├── 4. defer 数据结构
├── 5. defer 注册流程
├── 6. defer 执行流程
└── 7. defer 性能优化演进

第三部分：深入剖析
├── 8. 编译器优化
├── 9. 栈上分配 vs 堆分配
└── 10. defer 与 panic/recover

第四部分：实战应用
├── 11. 手撕代码 5 题
├── 12. 面试高频考点
└── 13. 性能优化技巧
```

---

## 第一部分：基础使用

### 1.1 defer 是什么

defer 用于延迟函数调用，确保函数在当前函数返回前执行，通常用于资源清理。

```go
func example() {
    defer fmt.Println("world")
    fmt.Println("hello")
}

// 输出:
// hello
// world
```

**核心特点：**
1. 延迟执行：在函数返回前执行
2. LIFO 顺序：后注册的先执行
3. 参数立即求值：defer 时确定参数值

---

### 1.2 defer 的三大特性

#### 特性 1: 延迟执行

```go
func readFile() error {
    f, err := os.Open("file.txt")
    if err != nil {
        return err
    }
    defer f.Close() // 函数返回前执行

    // 读取文件...
    data := make([]byte, 100)
    f.Read(data)

    return nil
} // 这里执行 f.Close()
```

#### 特性 2: LIFO 顺序（后进先出）

```go
func order() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}

// 输出:
// 3
// 2
// 1
```

**为什么是 LIFO？**
```go
func example() {
    mutex.Lock()
    defer mutex.Unlock() // 最后释放锁

    file, _ := os.Open("file.txt")
    defer file.Close() // 先关闭文件

    conn, _ := net.Dial("tcp", "example.com:80")
    defer conn.Close() // 最先关闭连接

    // 执行顺序: conn.Close() -> file.Close() -> mutex.Unlock()
}
```

#### 特性 3: 参数立即求值

```go
func trap1() {
    i := 0
    defer fmt.Println(i) // defer 时 i=0
    i++
    // 输出: 0 (不是 1)
}

func trap2() {
    i := 0
    defer func() {
        fmt.Println(i) // 闭包捕获 i 的引用
    }()
    i++
    // 输出: 1
}

func trap3() {
    i := 0
    defer func(n int) {
        fmt.Println(n) // 参数立即求值
    }(i) // 这里 i=0
    i++
    // 输出: 0
}
```

**对比表格：**

| 场景 | 代码 | 输出 | 原因 |
|------|------|------|------|
| 直接传参 | `defer fmt.Println(i)` | 0 | 立即求值 |
| 闭包引用 | `defer func() { fmt.Println(i) }()` | 1 | 捕获引用 |
| 闭包传参 | `defer func(n int) { fmt.Println(n) }(i)` | 0 | 立即求值 |

---

### 1.3 defer 常见用法

#### 用法 1: 资源释放

```go
// 文件操作
func processFile(filename string) error {
    f, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    // 使用文件...
    return nil
}

// 锁操作
func update() {
    mutex.Lock()
    defer mutex.Unlock()

    // 临界区代码...
}

// 数据库连接
func query() error {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return err
    }
    defer db.Close()

    // 执行查询...
    return nil
}
```

#### 用法 2: panic 恢复

```go
func safeCall(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
        }
    }()

    fn()
    return nil
}
```

#### 用法 3: 修改返回值

```go
// 注意: 必须使用命名返回值
func example() (result int, err error) {
    defer func() {
        if err != nil {
            err = fmt.Errorf("example failed: %w", err)
        }
    }()

    result = 10
    err = doSomething()
    return // 等价于 return result, err
}

// 实际应用
func divide(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("panic: %v", r)
            result = 0
        }
    }()

    return a / b, nil
}
```

#### 用法 4: 记录函数执行时间

```go
func trace(name string) func() {
    start := time.Now()
    fmt.Printf("enter %s\n", name)

    return func() {
        fmt.Printf("exit %s (took %v)\n", name, time.Since(start))
    }
}

func business() {
    defer trace("business")()

    time.Sleep(1 * time.Second)
    // 业务逻辑...
}

// 输出:
// enter business
// exit business (took 1.00s)
```

---

## 第二部分：实现原理

### 2.1 defer 数据结构

#### Go 1.12 及之前（堆分配）

```go
// runtime/runtime2.go
type _defer struct {
    siz     int32        // 参数和返回值的总大小
    started bool         // defer 是否已执行
    sp      uintptr      // 调用者栈指针
    pc      uintptr      // 返回地址
    fn      *funcval     // 延迟函数
    _panic  *_panic      // 触发 defer 的 panic
    link    *_defer      // 链表指针（指向下一个 defer）
}

// 每个 goroutine 维护一个 defer 链表
type g struct {
    // ...
    _defer *_defer  // defer 链表头（最新的 defer）
    // ...
}
```

**defer 链表结构：**
```
goroutine
    |
    v
  _defer (最新) -> _defer -> _defer -> nil
```

#### Go 1.13 优化（栈分配）

```go
// 在栈上分配小的 defer
type _defer struct {
    siz       int32
    started   bool
    heap      bool     // 是否在堆上分配
    sp        uintptr
    pc        uintptr
    fn        *funcval
    _panic    *_panic
    link      *_defer
    // 栈分配时的额外字段
    fd        unsafe.Pointer // funcdata
    varp      uintptr        // 变量指针
    framepc   uintptr        // 函数 PC
}
```

#### Go 1.14+ 优化（开放编码）

对于简单的 defer，编译器直接在函数末尾插入调用代码，无需 runtime 支持。

---

### 2.2 defer 注册流程

#### 步骤 1: 编译阶段

```go
// 源代码
func example() {
    defer println("hello")
    println("world")
}

// 编译器转换（简化）
func example() {
    // 注册 defer
    deferproc(siz, fn)  // 创建 _defer 结构体并加入链表

    println("world")

    // 函数返回前
    deferreturn()  // 执行所有 defer
}
```

#### 步骤 2: runtime 注册

```go
// runtime/panic.go
func deferproc(siz int32, fn *funcval) {
    // 获取调用者的 SP 和 PC
    sp := getcallersp()
    pc := getcallerpc()

    // 创建 _defer 结构体
    d := newdefer(siz)
    d.fn = fn
    d.sp = sp
    d.pc = pc

    // 加入 goroutine 的 defer 链表头部
    d.link = gp._defer
    gp._defer = d
}
```

**内存分配策略：**

```go
// Go 1.13+ 的优化
func newdefer(siz int32) *_defer {
    var d *_defer

    // 1. 小对象尝试从 P 的本地缓存获取
    if siz <= 32 && gp.p.deferpool != nil {
        d = gp.p.deferpool
        gp.p.deferpool = d.link
        d.heap = false // 标记为栈分配
    }

    // 2. 从全局缓存池获取
    if d == nil {
        d = sched.deferpool
        if d != nil {
            sched.deferpool = d.link
        }
    }

    // 3. 创建新的 defer
    if d == nil {
        d = new(_defer)
        d.heap = true // 标记为堆分配
    }

    d.siz = siz
    return d
}
```

---

### 2.3 defer 执行流程

#### 执行时机

```go
func example() (result int) {
    result = 1

    defer func() {
        result++  // ③ 修改返回值
    }()

    return 100  // ① 设置返回值 result=100
                // ② 执行 defer
                // ④ 真正返回 result=101
}
```

**完整流程：**
```
1. 执行 return 语句（设置返回值）
2. 执行 defer 链表（LIFO 顺序）
3. 函数真正返回
```

#### deferreturn 实现

```go
// runtime/panic.go
func deferreturn() {
    gp := getg()
    d := gp._defer

    // 遍历 defer 链表
    for {
        if d == nil {
            return
        }

        // 检查 SP，确保是当前函数的 defer
        sp := getcallersp()
        if d.sp != sp {
            return
        }

        // 执行 defer 函数
        fn := d.fn
        fn()

        // 移动到下一个 defer
        gp._defer = d.link

        // 释放 _defer 结构体
        freedefer(d)

        d = gp._defer
    }
}
```

---

### 2.4 defer 性能优化演进

#### Go 1.12 及之前：堆分配

```go
// 每个 defer 都在堆上分配
func example() {
    defer foo()  // 堆分配 _defer
    defer bar()  // 堆分配 _defer
}

// 性能：约 50ns/op
```

**问题：**
- 每个 defer 都需要堆分配
- GC 压力大
- 性能开销高

#### Go 1.13：栈分配优化

```go
// 大部分 defer 在栈上分配
func example() {
    defer foo()  // 栈分配 _defer（快）
    defer bar()  // 栈分配 _defer（快）
}

// 性能：约 10ns/op（提升 5 倍）
```

**优化原理：**
```go
// 在栈上预分配空间
type stackDefer struct {
    d _defer
    args [32]byte  // 参数空间
}

func example() {
    var d stackDefer  // 栈上分配

    // 初始化
    d.d.siz = 32
    d.d.heap = false  // 标记为栈分配
    d.d.fn = foo

    // 注册到链表
    d.d.link = gp._defer
    gp._defer = &d.d
}
```

#### Go 1.14+：开放编码（Open-coded defer）

```go
// 简单的 defer 直接内联
func example() {
    defer foo()
    bar()
}

// 编译器优化为：
func example() {
    bar()
    foo()  // 直接调用，无 runtime 开销
}

// 性能：约 1ns/op（提升 50 倍）
```

**开放编码条件：**
1. 函数内 defer 数量 ≤ 8
2. defer 数量 * 返回语句数量 ≤ 15
3. 没有循环中的 defer

**示例：**
```go
// ✅ 可以开放编码
func simple() {
    defer f1()
    defer f2()
    return
}

// ❌ 不能开放编码（defer 太多）
func tooMany() {
    defer f1()
    defer f2()
    // ... 10 个 defer
    return
}

// ❌ 不能开放编码（在循环中）
func inLoop() {
    for i := 0; i < 10; i++ {
        defer f(i)
    }
}
```

---

### 2.5 三种 defer 模式对比

| 模式 | Go 版本 | 性能 | 分配位置 | 适用场景 |
|------|---------|------|----------|----------|
| 堆分配 | ≤1.12 | 50ns | 堆 | 所有 defer |
| 栈分配 | 1.13 | 10ns | 栈 | 大部分 defer |
| 开放编码 | ≥1.14 | 1ns | - | 简单 defer |

**性能测试：**
```go
// benchmark_test.go
func BenchmarkDefer(b *testing.B) {
    for i := 0; i < b.N; i++ {
        deferFunc()
    }
}

func deferFunc() {
    defer func() {}()
}

// Go 1.12: 50 ns/op
// Go 1.13: 10 ns/op
// Go 1.14: 1 ns/op
```

---

## 第三部分：深入剖析

### 3.1 编译器视角

#### 查看编译后的汇编

```bash
# 生成汇编代码
go tool compile -S main.go > main.s
```

**示例代码：**
```go
package main

func example() {
    defer println("hello")
    println("world")
}
```

**汇编输出（简化）：**
```asm
"".example STEXT size=100 args=0x0 locals=0x18
    ; 注册 defer
    CALL runtime.deferproc(SB)

    ; 函数体
    CALL runtime.printstring(SB)

    ; 执行 defer
    CALL runtime.deferreturn(SB)
    RET
```

#### 开放编码优化（Go 1.14+）

```go
// 源代码
func example() {
    defer func() { println("1") }()
    defer func() { println("2") }()
    println("3")
}

// 编译器优化为（伪代码）
func example() {
    deferBits := 0  // 使用位图标记哪些 defer 需要执行

    deferBits |= 1<<0  // 标记 defer 0
    deferBits |= 1<<1  // 标记 defer 1

    println("3")

    // 函数返回前
    if deferBits & (1<<1) != 0 {
        println("2")
    }
    if deferBits & (1<<0) != 0 {
        println("1")
    }
}
```

---

### 3.2 defer 与 panic/recover

#### panic 触发时的 defer 执行

```go
func example() {
    defer fmt.Println("defer 1")
    defer fmt.Println("defer 2")

    panic("error")

    defer fmt.Println("defer 3") // 不会注册
}

// 输出:
// defer 2
// defer 1
// panic: error
```

**执行流程：**
```
1. 注册 defer 1
2. 注册 defer 2
3. panic 发生
4. 执行 defer 2（LIFO）
5. 执行 defer 1
6. 程序崩溃或被 recover
```

#### recover 只在 defer 中有效

```go
// ✅ 正确：recover 在 defer 中
func correct() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()

    panic("oops")
}

// ❌ 错误：recover 不在 defer 中
func wrong1() {
    if r := recover(); r != nil {
        fmt.Println("recovered:", r)
    }
    panic("oops") // 无法恢复
}

// ❌ 错误：recover 在错误的 goroutine
func wrong2() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()

    go func() {
        panic("oops") // 不同的 goroutine
    }()
}
```

#### defer + panic + recover 的完整流程

```go
func example() {
    defer func() {
        fmt.Println("defer 1")
    }()

    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()

    defer func() {
        fmt.Println("defer 3")
    }()

    panic("error")
}

// 输出:
// defer 3
// recovered: error
// defer 1
```

**数据结构关系：**
```go
type g struct {
    _defer *_defer  // defer 链表
    _panic *_panic  // panic 链表
}

type _panic struct {
    arg       interface{} // panic 参数
    recovered bool        // 是否被 recover
    link      *_panic     // 下一个 panic
}

// panic 时的处理
func gopanic(arg interface{}) {
    // 创建 _panic
    p := &_panic{arg: arg}
    p.link = gp._panic
    gp._panic = p

    // 执行 defer 链表
    for d := gp._defer; d != nil; d = d.link {
        d.fn()  // 执行 defer 函数

        // 检查是否 recover
        if gp._panic.recovered {
            // 恢复执行
            return
        }
    }

    // 没有 recover，程序崩溃
    fatalpanic()
}
```

---

### 3.3 defer 的常见陷阱

#### 陷阱 1: 循环中的 defer

```go
// ❌ 错误：defer 累积导致内存泄漏
func wrong() {
    for i := 0; i < 10000; i++ {
        f, _ := os.Open(fmt.Sprintf("file%d.txt", i))
        defer f.Close() // defer 会累积，函数结束才执行
    }
}

// ✅ 正确：使用匿名函数
func correct() {
    for i := 0; i < 10000; i++ {
        func() {
            f, _ := os.Open(fmt.Sprintf("file%d.txt", i))
            defer f.Close() // 每次迭代都会执行
        }()
    }
}

// ✅ 正确：手动 Close
func correct2() {
    for i := 0; i < 10000; i++ {
        f, _ := os.Open(fmt.Sprintf("file%d.txt", i))
        // 处理文件...
        f.Close()
    }
}
```

#### 陷阱 2: defer 与闭包

```go
func trap() {
    for i := 0; i < 3; i++ {
        defer func() {
            fmt.Println(i) // 闭包捕获 i 的引用
        }()
    }
}

// 输出: 3 3 3（不是 2 1 0）

// 修复方案 1: 传参
func fix1() {
    for i := 0; i < 3; i++ {
        defer func(n int) {
            fmt.Println(n)
        }(i) // 立即求值
    }
}

// 修复方案 2: 局部变量
func fix2() {
    for i := 0; i < 3; i++ {
        i := i // 创建新变量
        defer func() {
            fmt.Println(i)
        }()
    }
}
```

#### 陷阱 3: defer 与返回值

```go
// 陷阱：修改返回值
func trap1() int {
    result := 0
    defer func() {
        result++ // 无效！
    }()
    return 1 // 返回 1，不是 2
}

// 正确：使用命名返回值
func correct1() (result int) {
    defer func() {
        result++ // 有效！
    }()
    return 1 // 返回 2
}

// 陷阱：指针返回值
func trap2() *int {
    n := 1
    defer func() {
        n++ // 修改的是栈上的变量
    }()
    return &n // 返回的是 n 的地址
}

// 输出: 2（defer 确实执行了）
```

#### 陷阱 4: defer 与 nil

```go
// 陷阱：defer 一个 nil 函数
func trap() {
    var f func()
    defer f() // 运行时 panic: nil pointer dereference
}

// 正确：检查 nil
func correct() {
    var f func()
    if f != nil {
        defer f()
    }
}
```

---

## 第四部分：实战应用

### 🔥 手撕代码题 1: 实现 defer 执行顺序

**题目：** 预测输出结果

```go
func main() {
    defer func() {
        fmt.Println("defer 1")
    }()

    defer func() {
        fmt.Println("defer 2")
    }()

    fmt.Println("main")

    defer func() {
        fmt.Println("defer 3")
    }()
}
```

<details>
<summary>💡 答案</summary>

```
main
defer 3
defer 2
defer 1
```

**解释：**
1. 注册 defer 1
2. 注册 defer 2
3. 打印 "main"
4. 注册 defer 3
5. 函数返回，执行 defer（LIFO）：defer 3 -> defer 2 -> defer 1
</details>

---

### 🔥 手撕代码题 2: defer 修改返回值

**题目：** 预测输出结果

```go
func f1() int {
    x := 5
    defer func() {
        x++
    }()
    return x
}

func f2() (x int) {
    defer func() {
        x++
    }()
    return 5
}

func f3() (x int) {
    defer func(n int) {
        n++
    }(x)
    return 5
}

func main() {
    fmt.Println(f1()) // ?
    fmt.Println(f2()) // ?
    fmt.Println(f3()) // ?
}
```

<details>
<summary>💡 答案</summary>

```go
5  // f1: 返回匿名变量，defer 无法修改
6  // f2: 返回命名变量，defer 可以修改
5  // f3: 传参立即求值，修改的是副本
```

**详细解释：**

```go
// f1 等价于
func f1() int {
    x := 5
    _result := x  // 设置匿名返回值
    func() {
        x++  // 修改的是局部变量 x，不是返回值
    }()
    return _result  // 返回 5
}

// f2 等价于
func f2() (x int) {
    x = 5  // 设置命名返回值
    func() {
        x++  // 修改返回值 x
    }()
    return x  // 返回 6
}

// f3 等价于
func f3() (x int) {
    x = 5  // 设置命名返回值
    func(n int) {
        n++  // 修改的是参数 n（x 的副本），不是返回值 x
    }(x)  // 传参时 x=5
    return x  // 返回 5
}
```
</details>

---

### 🔥 手撕代码题 3: 实现资源池（使用 defer）

**题目：** 实现一个支持自动回收的资源池

```go
type Pool struct {
    resources chan interface{}
}

func NewPool(size int, factory func() interface{}) *Pool {
    // TODO: 实现
}

func (p *Pool) Acquire() (resource interface{}, release func()) {
    // TODO: 实现
    // 返回资源和释放函数，使用 defer release() 自动归还
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "sync"
    "time"
)

type Pool struct {
    resources chan interface{}
    factory   func() interface{}
    mu        sync.Mutex
    closed    bool
}

func NewPool(size int, factory func() interface{}) *Pool {
    p := &Pool{
        resources: make(chan interface{}, size),
        factory:   factory,
    }

    // 预创建资源
    for i := 0; i < size; i++ {
        p.resources <- factory()
    }

    return p
}

func (p *Pool) Acquire() (interface{}, func(), error) {
    p.mu.Lock()
    if p.closed {
        p.mu.Unlock()
        return nil, nil, errors.New("pool is closed")
    }
    p.mu.Unlock()

    select {
    case resource := <-p.resources:
        // 返回释放函数
        release := func() {
            p.mu.Lock()
            if !p.closed {
                p.resources <- resource
            }
            p.mu.Unlock()
        }
        return resource, release, nil

    case <-time.After(1 * time.Second):
        return nil, nil, errors.New("acquire timeout")
    }
}

func (p *Pool) Close() {
    p.mu.Lock()
    defer p.mu.Unlock()

    if p.closed {
        return
    }

    p.closed = true
    close(p.resources)

    // 清理资源
    for resource := range p.resources {
        if closer, ok := resource.(io.Closer); ok {
            closer.Close()
        }
    }
}

// 使用示例
func example() {
    pool := NewPool(5, func() interface{} {
        return &http.Client{Timeout: 5 * time.Second}
    })
    defer pool.Close()

    client, release, err := pool.Acquire()
    if err != nil {
        log.Fatal(err)
    }
    defer release() // 自动归还资源

    // 使用 client...
    httpClient := client.(*http.Client)
    resp, err := httpClient.Get("https://example.com")
    if err != nil {
        return
    }
    defer resp.Body.Close()
}
```
</details>

---

### 🔥 手撕代码题 4: 实现事务管理器

**题目：** 实现一个支持回滚的事务管理器

```go
type Transaction struct {
    // TODO: 定义字段
}

func NewTransaction() *Transaction {
    // TODO: 实现
}

func (t *Transaction) Execute(fn func() error) {
    // TODO: 实现
    // 如果失败，自动回滚之前的操作
}

func (t *Transaction) Commit() error {
    // TODO: 实现
}

func (t *Transaction) Rollback() error {
    // TODO: 实现
}
```

<details>
<summary>💡 参考答案</summary>

```go
import (
    "errors"
    "fmt"
)

type Transaction struct {
    operations []func() error  // 操作列表
    rollbacks  []func() error  // 回滚操作列表
    committed  bool
    rolledback bool
}

func NewTransaction() *Transaction {
    return &Transaction{
        operations: make([]func() error, 0),
        rollbacks:  make([]func() error, 0),
    }
}

// AddOperation 添加操作和对应的回滚函数
func (t *Transaction) AddOperation(op func() error, rollback func() error) {
    t.operations = append(t.operations, op)
    t.rollbacks = append(t.rollbacks, rollback)
}

// Execute 执行操作，失败时自动回滚
func (t *Transaction) Execute(op func() error, rollback func() error) error {
    if t.committed || t.rolledback {
        return errors.New("transaction already finished")
    }

    // 执行操作
    if err := op(); err != nil {
        // 执行失败，回滚所有已执行的操作
        t.Rollback()
        return fmt.Errorf("operation failed: %w", err)
    }

    // 记录回滚操作（LIFO）
    t.rollbacks = append([]func() error{rollback}, t.rollbacks...)

    return nil
}

// Commit 提交事务
func (t *Transaction) Commit() error {
    if t.committed {
        return errors.New("already committed")
    }
    if t.rolledback {
        return errors.New("already rolled back")
    }

    t.committed = true
    t.rollbacks = nil // 清空回滚操作
    return nil
}

// Rollback 回滚事务
func (t *Transaction) Rollback() error {
    if t.committed {
        return errors.New("cannot rollback committed transaction")
    }
    if t.rolledback {
        return nil // 已回滚
    }

    t.rolledback = true

    // 执行所有回滚操作（LIFO）
    var errs []error
    for _, rollback := range t.rollbacks {
        if err := rollback(); err != nil {
            errs = append(errs, err)
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("rollback errors: %v", errs)
    }

    return nil
}

// 使用示例
func transferMoney(fromAccount, toAccount *Account, amount int) error {
    tx := NewTransaction()

    // 操作1: 扣款
    err := tx.Execute(
        func() error {
            return fromAccount.Deduct(amount)
        },
        func() error {
            return fromAccount.Add(amount) // 回滚：加回去
        },
    )
    if err != nil {
        return err
    }

    // 操作2: 加款
    err = tx.Execute(
        func() error {
            return toAccount.Add(amount)
        },
        func() error {
            return toAccount.Deduct(amount) // 回滚：扣回去
        },
    )
    if err != nil {
        return err
    }

    // 操作3: 记录日志
    err = tx.Execute(
        func() error {
            return writeLog(fmt.Sprintf("transfer %d from %s to %s",
                amount, fromAccount.ID, toAccount.ID))
        },
        func() error {
            return deleteLog() // 回滚：删除日志
        },
    )
    if err != nil {
        return err
    }

    // 提交事务
    return tx.Commit()
}
```
</details>

---

### 🔥 手撕代码题 5: defer 性能优化

**题目：** 优化以下代码的性能

```go
// 原始代码（慢）
func processItems(items []Item) error {
    for _, item := range items {
        defer cleanup(item) // 每次循环都 defer

        if err := process(item); err != nil {
            return err
        }
    }
    return nil
}

// TODO: 优化这段代码
```

<details>
<summary>💡 优化方案</summary>

```go
// 方案1: 使用匿名函数（推荐）
func processItems1(items []Item) error {
    for _, item := range items {
        if err := func() error {
            defer cleanup(item) // 每次迭代结束就执行

            return process(item)
        }(); err != nil {
            return err
        }
    }
    return nil
}

// 方案2: 手动清理
func processItems2(items []Item) error {
    for _, item := range items {
        err := process(item)
        cleanup(item) // 手动调用

        if err != nil {
            return err
        }
    }
    return nil
}

// 方案3: 收集需要清理的资源，统一处理
func processItems3(items []Item) error {
    var toCleanup []Item

    defer func() {
        // 统一清理
        for _, item := range toCleanup {
            cleanup(item)
        }
    }()

    for _, item := range items {
        if err := process(item); err != nil {
            return err
        }
        toCleanup = append(toCleanup, item)
    }

    return nil
}

// 性能对比
// 原始代码: 100000 次循环，约 500ms
// 方案1:    100000 次循环，约 100ms（提升 5 倍）
// 方案2:    100000 次循环，约 50ms（提升 10 倍）
// 方案3:    100000 次循环，约 100ms（适合必须延迟清理的场景）
```
</details>

---

## 面试高频考点

### 考点 1: defer 的执行顺序

**问题：** defer 为什么是 LIFO（后进先出）？

**答案：**
```go
// 1. 数据结构：单链表
type g struct {
    _defer *_defer  // 链表头（最新的 defer）
}

type _defer struct {
    link *_defer  // 指向下一个（更早的）defer
}

// 2. 注册过程：头插法
func deferproc(fn *funcval) {
    d := newdefer()
    d.fn = fn
    d.link = gp._defer  // 新 defer 指向旧链表
    gp._defer = d       // 更新链表头
}

// 3. 执行过程：从头开始
func deferreturn() {
    for d := gp._defer; d != nil; d = d.link {
        d.fn()  // 执行当前 defer
    }
}

// 结果：最后注册的最先执行（LIFO）
```

**为什么设计成 LIFO？**
1. 资源清理的自然顺序（先获取的后释放）
2. 链表头插法效率高（O(1)）
3. 符合栈的语义（函数调用栈）

---

### 考点 2: defer 参数求值时机

**问题：** defer 的参数什么时候求值？

**答案：**
```go
// defer 时立即求值
func example1() {
    i := 0
    defer fmt.Println(i) // defer 时 i=0
    i++
}
// 输出: 0

// 闭包延迟求值
func example2() {
    i := 0
    defer func() {
        fmt.Println(i) // 函数返回时求值
    }()
    i++
}
// 输出: 1

// 原因
func deferproc(siz int32, fn *funcval, args ...interface{}) {
    d := newdefer(siz)
    d.fn = fn

    // 拷贝参数（立即求值）
    memmove(d.args, args, siz)

    // ...
}
```

**面试题：**
```go
func test() {
    for i := 0; i < 3; i++ {
        defer fmt.Println(i)
    }
}
// 输出: 2 1 0（不是 3 3 3）
```

---

### 考点 3: defer 与返回值

**问题：** defer 如何修改返回值？

**答案：**
```go
// 完整的 return 流程
func example() (result int) {
    result = 1     // ① 设置返回值
    // defer 执行  // ② 执行 defer
    return         // ③ 返回
}

// 等价于
func example() (result int) {
    result = 1
    // defer 在这里执行，可以修改 result
    return result
}

// 只有命名返回值才能修改
func f1() (x int) {
    defer func() { x++ }()
    return 5  // 返回 6
}

func f2() int {
    x := 5
    defer func() { x++ }()
    return x  // 返回 5（修改的是局部变量）
}
```

---

### 考点 4: defer 性能优化

**问题：** Go 1.14 对 defer 做了什么优化？

**答案：**

| 版本 | 优化 | 性能 | 原理 |
|------|------|------|------|
| ≤1.12 | 无 | 50ns | 堆分配 |
| 1.13 | 栈分配 | 10ns | 在栈上分配 _defer |
| ≥1.14 | 开放编码 | 1ns | 编译器直接内联 |

**开放编码示例：**
```go
// 源代码
func f() {
    defer g()
    h()
}

// 编译器优化为
func f() {
    deferBits := 0
    deferBits |= 1 << 0

    h()

    if deferBits & (1<<0) != 0 {
        g()
    }
}

// 好处：
// 1. 无 runtime 调用
// 2. 无堆/栈分配
// 3. 几乎零开销
```

**限制条件：**
```go
// ✅ 可以开放编码
func simple() {
    defer f1()
    defer f2()
    return
}

// ❌ 不能开放编码
func complex() {
    for i := 0; i < 10; i++ {
        defer f(i) // 在循环中
    }
}
```

---

### 考点 5: defer 与 panic

**问题：** panic 时 defer 如何执行？

**答案：**
```go
// panic 流程
func gopanic(arg interface{}) {
    // 1. 创建 _panic 结构体
    p := &_panic{arg: arg}
    gp._panic = p

    // 2. 执行 defer 链表
    for d := gp._defer; d != nil; d = d.link {
        d.fn()

        // 3. 检查是否 recover
        if gp._panic.recovered {
            // 恢复正常执行
            goexit()
        }
    }

    // 4. 没有 recover，程序崩溃
    fatalpanic()
}

// recover 流程
func gorecover() interface{} {
    p := gp._panic
    if p != nil && !p.recovered {
        p.recovered = true
        return p.arg
    }
    return nil
}
```

**示例：**
```go
func test() {
    defer fmt.Println("1")
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("recovered:", r)
        }
    }()
    defer fmt.Println("3")

    panic("error")
}

// 输出:
// 3
// recovered: error
// 1
```

---

### 考点 6: defer 的内存泄漏

**问题：** defer 可能导致什么问题？

**答案：**
```go
// 问题1: 循环中的 defer
func leak1() {
    for i := 0; i < 100000; i++ {
        f, _ := os.Open(fmt.Sprintf("file%d.txt", i))
        defer f.Close() // 累积 100000 个 defer
    }
} // 函数返回时才执行，可能耗尽文件描述符

// 问题2: 长时间运行的函数
func leak2() {
    defer cleanup() // 只在函数结束时执行

    for {
        data := allocateLargeMemory()
        // 使用 data...
        // data 无法及时释放
    }
}

// 解决方案
func fix() {
    for i := 0; i < 100000; i++ {
        func() {
            f, _ := os.Open(fmt.Sprintf("file%d.txt", i))
            defer f.Close() // 每次迭代都执行
            // 处理文件...
        }()
    }
}
```

---

### 考点 7: defer 底层实现

**问题：** 手写 defer 的简化实现

**答案：**
```go
// 简化的 defer 实现

// defer 链表节点
type _defer struct {
    fn   func()
    link *_defer
}

// goroutine 结构
type g struct {
    _defer *_defer
}

var currentG *g = &g{}

// 注册 defer
func deferproc(fn func()) {
    d := &_defer{
        fn:   fn,
        link: currentG._defer,
    }
    currentG._defer = d
}

// 执行 defer
func deferreturn() {
    for d := currentG._defer; d != nil; d = d.link {
        d.fn()
    }
    currentG._defer = nil
}

// 使用示例
func example() {
    deferproc(func() { fmt.Println("defer 1") })
    deferproc(func() { fmt.Println("defer 2") })

    fmt.Println("main")

    deferreturn()
}
// 输出: main, defer 2, defer 1
```

---

## 性能优化技巧

### 技巧 1: 避免循环中的 defer

```go
// ❌ 慢（100000 次循环约 500ms）
for i := 0; i < 100000; i++ {
    defer func() {}()
}

// ✅ 快（100000 次循环约 100ms）
for i := 0; i < 100000; i++ {
    func() {
        defer func() {}()
    }()
}
```

### 技巧 2: 使用开放编码条件

```go
// ✅ 快（开放编码，约 1ns）
func f() {
    defer g()
    defer h()
    return
}

// ❌ 慢（无法开放编码，约 10ns）
func f() {
    for i := 0; i < 10; i++ {
        defer g()
    }
}
```

### 技巧 3: 减少 defer 数量

```go
// ❌ 慢（多个 defer）
func f() {
    defer cleanup1()
    defer cleanup2()
    defer cleanup3()
}

// ✅ 快（合并 defer）
func f() {
    defer func() {
        cleanup1()
        cleanup2()
        cleanup3()
    }()
}
```

### 技巧 4: 关键路径避免 defer

```go
// 高频调用的函数避免 defer
func hotPath() {
    mu.Lock()
    // 手动 unlock
    result := compute()
    mu.Unlock()
    return result
}

// 低频调用的函数使用 defer
func coldPath() {
    mu.Lock()
    defer mu.Unlock()
    // ...
}
```

---

## 总结

### 核心要点

1. **defer 三大特性**
   - 延迟执行（函数返回前）
   - LIFO 顺序（后进先出）
   - 参数立即求值

2. **实现原理**
   - 链表结构（头插法）
   - 三种模式（堆/栈/开放编码）
   - Go 1.14+ 性能提升 50 倍

3. **使用场景**
   - 资源释放（文件、锁、连接）
   - panic 恢复
   - 修改返回值
   - 记录日志

4. **常见陷阱**
   - 循环中的 defer
   - 闭包捕获变量
   - 修改返回值（需要命名返回值）
   - nil 函数

5. **性能优化**
   - 避免循环中的 defer
   - 使用开放编码条件
   - 关键路径避免 defer
   - 合并多个 defer

### 学习路线

```
1️⃣ 基础使用（1天）
   └── 掌握三大特性，避免常见陷阱

2️⃣ 实现原理（2天）
   └── 理解数据结构、执行流程、性能优化

3️⃣ 实战应用（2天）
   └── 完成手撕代码，学习最佳实践

4️⃣ 面试准备（1天）
   └── 背诵核心考点，模拟面试
```

### 面试必背

1. defer 执行顺序（LIFO）
2. defer 参数求值时机（立即）
3. defer 修改返回值（命名返回值）
4. defer 性能优化（开放编码）
5. defer 与 panic/recover
6. defer 的内存泄漏
7. defer 底层实现（链表）

---

**掌握 defer，你就掌握了 Go 资源管理的精髓！** 🚀
