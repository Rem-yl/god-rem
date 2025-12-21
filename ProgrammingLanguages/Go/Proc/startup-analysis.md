# Go 程序启动流程分析

## 一、编译阶段

### 1.1 编译命令

```bash
go build -x -work -o main_bin main.go
```

### 1.2 编译过程

Go 编译分为两个主要阶段：

#### 阶段1：编译（Compile）

```bash
/usr/local/go/pkg/tool/linux_amd64/compile \
  -o $WORK/b001/_pkg_.a \
  -p main \
  -lang=go1.23 \
  -complete \
  -buildid oQIQ34onvGq6FLrJ9jBJ/oQIQ34onvGq6FLrJ9jBJ \
  -importcfg $WORK/b001/importcfg \
  -pack ./main.go
```

**关键点**：

- 输出：`.a` 归档文件（包含目标代码）
- 导入配置：`importcfg` 指定依赖包的位置
- 主要依赖：fmt, time, runtime

#### 阶段2：链接（Link）

```bash
/usr/local/go/pkg/tool/linux_amd64/link \
  -o $WORK/b001/exe/a.out \
  -importcfg $WORK/b001/importcfg.link \
  -buildmode=exe \
  -buildid=JWCKsZkDc3FB174LP1D0/... \
  -extld=gcc \
  $WORK/b001/_pkg_.a
```

**关键点**：

- 静态链接所有依赖（包括 runtime）
- 生成独立的可执行文件
- 文件大小：2.1MB（包含完整的 Go runtime）

---

## 二、可执行文件结构（ELF）

### 2.1 基本信息

```
文件格式：ELF 64-bit LSB executable
架构：x86-64
链接方式：静态链接（statically linked）
入口点：0x46ce40 (_rt0_amd64_linux)
```

### 2.2 关键 Section

| Section    | 地址     | 大小    | 说明          |
| ---------- | -------- | ------- | ------------- |
| .text      | 0x401000 | 0x8e168 | 代码段        |
| .rodata    | 0x490000 | 0x464de | 只读数据      |
| .gopclntab | 0x4d6c20 | 0x70bb0 | PC-Line Table |
| .gosymtab  | 0x4d6c08 | 0       | Go 符号表     |
| .data      | 0x54d680 | 0x4e30  | 已初始化数据  |
| .bss       | 0x5524c0 | 0x1ff00 | 未初始化数据  |

**重要**：

- `.gopclntab`：程序计数器到源代码行号的映射表，用于 stack trace
- `.gosymtab`：Go 特有的符号信息

### 2.3 Program Headers

```
LOAD 0x000000 0x400000 0x400000 R E  (代码段，可读可执行)
LOAD 0x090000 0x490000 0x490000 R    (只读数据段)
LOAD 0x148000 0x548000 0x548000 RW   (读写数据段)
```

---

## 三、程序启动流程

### 3.1 启动链路

```
OS 加载器
  ↓
_rt0_amd64_linux (入口点 0x46ce40)
  ↓
_rt0_amd64 (0x469740)
  ↓
runtime.rt0_go (0x469760)  ← 核心启动函数
  ↓
runtime.schedinit  ← 调度器初始化
  ↓
runtime.main
  ↓
main.main
```

### 3.2 _rt0_amd64_linux (入口点)

**地址**: 0x46ce40
**汇编代码**:

```asm
_rt0_amd64_linux:
  jmp    _rt0_amd64
```

**功能**：平台特定的入口点，直接跳转到通用启动代码。

---

### 3.3 _rt0_amd64

**地址**: 0x469740
**汇编代码**:

```asm
_rt0_amd64:
  mov    rdi, QWORD PTR [rsp]       # argc
  lea    rsi, [rsp+0x8]             # argv
  jmp    runtime.rt0_go
```

**功能**：

1. 从栈上读取 argc（参数个数）
2. 计算 argv（参数数组）的地址
3. 跳转到 runtime.rt0_go

**重要**：Linux 内核启动进程时，栈的布局：

```
[rsp+0]   = argc
[rsp+8]   = argv[0]
[rsp+16]  = argv[1]
...
```

---

### 3.4 runtime.rt0_go（核心启动）

**地址**: 0x469760
**源码位置**: `runtime/asm_amd64.s`

这是 Go runtime 启动的核心函数，执行以下步骤：

#### 步骤1：栈对齐与参数保存

```asm
mov    rax, rdi              # 保存 argc
mov    rbx, rsi              # 保存 argv
sub    rsp, 0x28
and    rsp, 0xfffffffffffffff0   # 16字节对齐
mov    [rsp+0x18], rax
mov    [rsp+0x20], rbx
```

#### 步骤2：初始化 g0（goroutine 0）

```asm
lea    rdi, [runtime.g0]           # g0 的地址
lea    rbx, [rsp-0x10000]          # 栈底
mov    [rdi+0x10], rbx             # g0.stackguard0
mov    [rdi+0x18], rbx             # g0.stackguard1
mov    [rdi], rbx                  # g0.stack.lo
mov    [rdi+0x8], rsp              # g0.stack.hi
```

**g0 说明**：

- g0 是每个 M（OS 线程）的特殊 goroutine
- 用于执行调度、GC 等运行时任务
- 不同于用户代码的 goroutine

#### 步骤3：CPU 特性检测

```asm
mov    eax, 0x0
cpuid                              # 执行 CPUID 指令
cmp    eax, 0x0
je     skip_cpu_check

# 检测是否为 Intel CPU
cmp    ebx, 0x756e6547    # "Genu"
jne    not_intel
cmp    edx, 0x49656e69    # "ineI"
jne    not_intel
cmp    ecx, 0x6c65746e    # "ntel"
jne    not_intel
mov    BYTE PTR [runtime.isIntel], 0x1

not_intel:
mov    eax, 0x1
cpuid
mov    [runtime.processorVersionInfo], eax
```

**功能**：检测 CPU 厂商和特性，用于后续优化。

#### 步骤4：CGO 初始化（如果需要）

```asm
mov    rax, [_cgo_init]
test   rax, rax
je     no_cgo
# ... CGO 初始化代码
```

#### 步骤5：设置 TLS（Thread Local Storage）

```asm
lea    rdi, [runtime.m0+0x88]     # m0.tls[0] 的地址
call   runtime.settls              # 设置线程本地存储

# 验证 TLS 是否工作
mov    QWORD PTR fs:0xfffffffffffffff8, 0x123
mov    rax, [runtime.m0+0x88]
cmp    rax, 0x123
je     tls_ok
call   runtime.abort
```

**TLS 说明**：

- `fs:0xfffffffffffffff8` 指向当前 goroutine (g)
- 这是 Go runtime 快速访问当前 g 的关键机制

#### 步骤6：关联 g0 和 m0

```asm
lea    rcx, [runtime.g0]
mov    QWORD PTR fs:0xfffffffffffffff8, rcx  # TLS = g0
lea    rax, [runtime.m0]
mov    [rax], rcx                # m0.g0 = g0
mov    [rcx+0x30], rax           # g0.m = m0
```

**数据结构**：

```
m0 (主线程)
  ├── g0 → g0 (调度 goroutine)
  └── curg → (当前运行的 goroutine)

g0
  └── m → m0
```

#### 步骤7：Runtime 初始化调用链

```asm
cld                               # 清除方向标志
call   runtime.check              # 内部一致性检查

mov    eax, [rsp+0x18]           # argc
mov    [rsp], eax
mov    rax, [rsp+0x20]           # argv
mov    [rsp+0x8], rax

call   runtime.args               # 解析命令行参数
call   runtime.osinit             # OS 特定初始化
call   runtime.schedinit          # 调度器初始化 ← 关键
```

#### 步骤8：创建 main goroutine

```asm
lea    rax, [runtime.mainPC]      # runtime.main 的地址
push   rax
call   runtime.newproc            # 创建新 goroutine
pop    rax
```

**重要**：这里创建的是运行 `runtime.main` 的 goroutine，不是直接的 `main.main`。

#### 步骤9：启动调度

```asm
call   runtime.mstart             # 启动 M，永不返回
call   runtime.abort              # 如果返回了，则异常
```

---

## 四、关键 Runtime 函数

### 4.1 runtime.check

- 检查内部数据结构的大小和对齐
- 确保编译期和运行期的一致性

### 4.2 runtime.args

- 解析 `argc` 和 `argv`
- 保存到 `runtime.args` 全局变量

### 4.3 runtime.osinit

- 获取 CPU 核心数（`GOMAXPROCS` 默认值）
- Linux 上读取 `/proc/self/auxv` 获取辅助信息
- 初始化页大小、物理页数等

### 4.4 runtime.schedinit

**这是调度器初始化的核心函数**，后续详细分析。

### 4.5 runtime.newproc

创建新的 goroutine：

- 分配 goroutine 结构体（g）
- 设置栈空间
- 将函数地址和参数保存到栈上
- 将 g 放入运行队列

### 4.6 runtime.mstart

启动 M（machine，OS 线程）：

- 永不返回的函数
- 进入调度循环
- 不断从队列中取出 goroutine 并执行

---

## 五、关键数据结构

### 5.1 runtime.g0

```go
// 全局变量，主线程的 g0
var g0 g
```

**地址**: 0x552c80
**作用**：

- 每个 M 都有一个 g0
- 用于执行运行时代码（调度、GC）
- 拥有固定大小的栈（~64KB）

### 5.2 runtime.m0

```go
// 全局变量，主线程
var m0 m
```

**地址**: 0x5538e0
**作用**：

- 程序启动时的主 OS 线程
- 关联 g0
- 执行第一个 goroutine

### 5.3 runtime.mainPC

```go
// runtime.main 函数的程序计数器
var mainPC uintptr
```

**地址**: 0x4d40a0
**值**: runtime.main 的地址

---

## 六、OS 到 Runtime 的交互

### 6.1 进程启动

1. **内核加载器**：读取 ELF 文件
2. **设置内存映射**：代码段、数据段、栈
3. **设置寄存器**：
   - `rip = 0x46ce40` (入口点)
   - `rsp = 栈顶`
4. **跳转到入口点**

### 6.2 栈布局

```
高地址
  ┌────────────┐
  │   envp[]   │  环境变量
  ├────────────┤
  │   argv[]   │  命令行参数
  ├────────────┤
  │    argc    │  ← rsp（初始）
  └────────────┘
低地址
```

### 6.3 TLS 机制

在 x86-64 Linux 上：

- 使用 `fs` 段寄存器
- `arch_prctl(ARCH_SET_FS, addr)` 系统调用设置
- Go 使用 `fs:-8` 存储当前 g 的指针

### 6.4 系统调用

Go runtime 使用原始系统调用（syscall），不依赖 libc：

```asm
# Linux syscall 示例
mov rax, syscall_number
mov rdi, arg1
mov rsi, arg2
syscall
```

---

## 七、下一步分析

1. **runtime.schedinit** 详细分析

   - GMP 模型初始化
   - P（Processor）的创建
   - 全局队列和本地队列的初始化
2. **runtime.main** 分析

   - 包初始化
   - main.main 的调用
   - 程序退出处理
3. **runtime.mstart** 和调度循环

   - schedule()
   - execute()
   - goroutine 切换
4. **goroutine 创建和调度**

   - newproc 详解
   - 栈管理
   - 抢占机制
5. **系统调用处理**

   - entersyscall / exitsyscall
   - sysmon 监控线程
