# Go世界：从开始到终结

我们将从以下这个简单代码入手，去探究go语言从编译到运行到程序结束发生的一系列故事，让我们探索go世界的奥妙！

Go Version: 1.23.0

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello, world!")

	go func() {
		fmt.Println("Hello, goroutine world!")
		time.Sleep(5 * time.Second)
	}()

	time.Sleep(6 * time.Second)
}
```

## 1. 编译世界

### 编译阶段

本小节我们探究，当我们运行以下的编译指令编译go代码时发生了什么。

```bash
go build -o main_bin main.go
```

我们可以使用如下的命令来可视化编译时的操作，注意如果你在之前使用该命令编译过，那么你运行相同的命令编译时go会使用缓存跳过编译阶段，此时需要使用 `go clean -cache`来清除缓存来重新编译

```bash
go build -x -work -o main_bin main.go 2>&1 | tee build_log.log
```

**参数说明**

- `-x`: 打印编译过程中执行的命令
- `-work`: 打印编译时的临时工作目录，并且不删除该目录

我们来分析编译产生的 `build_log.log`文件，看看go是如何将我们的代码编译成计算机可以执行的文件的。

**Tips: 善用 `grep -n "partten" file`命令从文件中查找**

接下来是对 `build_log.log` 编译日志分析

#### 编译顺序

日志范围：第1-589行，到 cat >/tmp/go-build3886251060/b002/importcfg这段结束

这段展示了go编译时的依赖树构建：底层基础包 → 中间层包 → 高层包 → main 包

可以使用下面命令从日志中查找对应的编译指令

```bash
grep -n "compile.*internal/goarch" build_log.log
```

```bash
# 第6行：最底层的包之一
/usr/local/go/pkg/tool/linux_amd64/compile -p internal/coverage/rtcov ...

# 第21行：架构相关的底层包
/usr/local/go/pkg/tool/linux_amd64/compile -p internal/goarch ...

# 第256行：runtime 包（核心中的核心）
/usr/local/go/pkg/tool/linux_amd64/compile -p runtime ...

# 第590行：fmt 包（依赖 runtime）
/usr/local/go/pkg/tool/linux_amd64/compile -p fmt ...

# 第600行：main.go（最后编译）
/usr/local/go/pkg/tool/linux_amd64/compile -p main ... ./main.go
```

学习要点：

- Go 编译器自动解析依赖关系
- 并行编译（多个 mkdir 同时执行）
- 编译顺序严格遵循依赖树

#### 汇编代码的编译流程

**重点**： 两阶段编译(第45, 235行等)

阶段A：生成符号表(-gensymabis)

```bash
# 第235行：runtime 汇编文件生成符号表
/usr/local/go/pkg/tool/linux_amd64/asm \
-p runtime \
-gensymabis \
-o $WORK/b009/symabis \
./asm.s ./asm_amd64.s ./rt0_linux_amd64.s ...
```

**symabis(符号表)的作用：让汇编代码和go代码可以互相调用**

```bash
# 第235行生成的 symabis 文件包含：
./asm.s              # 包含 runtime·gogo, runtime·mcall 等
./rt0_linux_amd64.s  # 包含 _rt0_amd64_linux, runtime·rt0_go
./sys_linux_amd64.s  # 包含系统调用相关函数
```

symabis 会告诉 Go 编译器：

- 这些汇编函数的名称
- 函数的参数数量和类型
- 函数的返回值
- 调用约定（栈帧大小等）

关键文件：

- rt0_linux_amd64.s - 程序启动入口！
- asm_amd64.s - 核心汇编函数（gogo, mcall等）
- sys_linux_amd64.s - 系统调用

阶段B：编译Go代码（使用符号表）

```bash
# 第256行：编译 runtime 包时使用 symabis
/usr/local/go/pkg/tool/linux_amd64/compile \
-p runtime \
-symabis $WORK/b009/symabis \  # ← 使用汇编符号表
-asmhdr $WORK/b009/go_asm.h \  # ← 生成头文件供汇编使用
/usr/local/go/src/runtime/proc.go ...
```

作用：

- Go编译器读取 symabis
- 知道哪些函数在汇编中实现（不报错“未定义”）
- 生成 `go_asm.h`包含Go常量和结构体偏移量

阶段C：汇编到目标文件

```bash
# 第268-269行：将汇编代码编译为 .o 文件
# 汇编代码使用 go_asm.h
/usr/local/go/pkg/tool/linux_amd64/asm \
-I $WORK/b009/ \              # ← 包含 go_asm.h 的目录
-o $WORK/b009/asm.o \
./asm.s

/usr/local/go/pkg/tool/linux_amd64/asm \
-o $WORK/b009/asm_amd64.o \
./asm_amd64.s
```

作用：

- 汇编代码可以使用 `#include "go_asm.h"`
- 访问Go结构体的字段偏移量
- 使用Go定义的常量

**💡 为什么需要两阶段编译？**
问题：鸡和蛋的循环依赖

Go 代码需要：
→ 知道汇编函数的签名（才能调用）

汇编代码需要：
→ 知道 Go 结构体的偏移量（才能访问字段）

解决方案：两阶段编译

阶段A：symabis
汇编 → 提取签名 → symabis 文件

阶段B：Go 编译
symabis + Go 代码 → 编译 → _pkg_.a + go_asm.h

阶段C：汇编编译
go_asm.h + 汇编代码 → 编译 → .o 文件

最终链接：
_pkg_.a + .o 文件 → 完整的包

#### runtime包编译

这是整个编译的核心部分，runtime包包含了150+个源文件，使用 `-symabis`链接汇编代码

```bash
grep -n "compile.*-p runtime " build_log.log
```

```bash
# 第256行：runtime 包包含 150+ 个源文件
/usr/local/go/pkg/tool/linux_amd64/compile -p runtime ...
	/usr/local/go/src/runtime/proc.go        # ← GMP 调度核心
	/usr/local/go/src/runtime/runtime2.go    # ← g, m, p 结构定义
	/usr/local/go/src/runtime/chan.go        # channel 实现
	/usr/local/go/src/runtime/mgc.go         # GC
	/usr/local/go/src/runtime/malloc.go      # 内存分配
	/usr/local/go/src/runtime/panic.go       # panic/recover
	/usr/local/go/src/runtime/netpoll.go     # 网络轮询
	/usr/local/go/src/runtime/signal_unix.go # 信号处理
	... 还有 140+ 个文件
```

#### main.go编译

1. 创建 `importcfg`

   ```bash
   # 第594-599行：创建 importcfg（main.go 的依赖）
   cat > importcfg << 'EOF'
   	packagefile fmt=/tmp/go-build3886251060/b002/_pkg_.a      			# ← 临时目录中的编译结果
   	packagefile time=/tmp/go-build3886251060/b045/_pkg_.a
   	packagefile runtime=/tmp/go-build3886251060/b009/_pkg_.a
   	EOF
   ```

   作用：告诉编译器 `main.go`直接导入的包在哪里

   使用位置：

   ```bash
   # 第600行
   /usr/local/go/pkg/tool/linux_amd64/compile \
   	-importcfg $WORK/b001/importcfg \  # ← 读取这个配置
   	./main.go
   ```
2. 编译 `main.go`

   ```bash
   # 第600行：编译 main.go
   /usr/local/go/pkg/tool/linux_amd64/compile \
   	-o $WORK/b001/_pkg_.a \
   	-p main \                    # 包名
   	-lang=go1.23 \              # Go 版本
   	-complete \                 # 完整包（非增量）
   	-buildid ... \
   	-c=4 \                      # 4个并发 goroutine
   	-importcfg $WORK/b001/importcfg \  # ← 依赖配置
   	-pack \
   	./main.go
   ```

   - `-c=4`表示使用4个CPU核心并发编译
   - `importcfg`指向临时编译目录，并不是缓存
3. 缓存编译结果

   ```bash
   # 第601行：设置 buildid
   /usr/local/go/pkg/tool/linux_amd64/buildid -w $WORK/b001/_pkg_.a

   # 第602行：缓存编译结果
   cp $WORK/b001/_pkg_.a /root/.cache/go-build/0b/...
   ```

   buildid（Build ID）是 Go 编译器为每个编译产物生成的唯一标识符，用于：

   1. 增量编译的基础

      - 通过 actionID 判断是否需要重新编译
      - 避免重复编译未修改的代码
   2. 缓存管理的索引

      - contentID 用作缓存文件名的一部分
      - 快速定位缓存文件
   3. 依赖追踪

      - 依赖包的 buildid 变化会传播
      - 确保所有受影响的包重新编译
   4. 完整性验证

      - contentID 验证缓存没有被篡改
      - 保证编译结果的正确性

### 链接阶段

1. 创建链接配置

   ```bash
   # 第603-658行：创建链接配置（55个依赖包）
   cat > importcfg.link << 'EOF'
   	packagefile command-line-arguments=/tmp/.../b001/_pkg_.a  # ← main 包
   	packagefile fmt=/tmp/.../b002/_pkg_.a
   	packagefile time=/tmp/.../b045/_pkg_.a
   	packagefile runtime=/tmp/.../b009/_pkg_.a  # ← runtime 包
   	... 还有 51 个包
   	packagefile path=/tmp/.../b044/_pkg_.a

   	# 第657行：modinfo（构建信息）
   	modinfo "...GOOS=linux\nGOARCH=amd64..."
   	EOF
   ```
2. 链接

   链接器将55个包合并成一个可执行文件

   ```bash
   # 第661行：链接（最关键的一步）
   GOROOT='/usr/local/go' /usr/local/go/pkg/tool/linux_amd64/link 
   	-o $WORK/b001/exe/a.out \              # 输出文件
   	-importcfg $WORK/b001/importcfg.link \ # 所有依赖
   	-buildmode=exe \                       # 可执行文件
   	-buildid=... \
   	-extld=gcc \                           # 外部链接器（CGO）
   	$WORK/b001/_pkg_.a                     # main 包
   ```
3. 设置最终的buildid

   ```bash
   # 第662行：设置最终 buildid
   /usr/local/go/pkg/tool/linux_amd64/buildid s-w $WORK/b001/exe/a.out

   # 第663行：移动到目标位置
   mv $WORK/b001/exe/a.out main_bin
   ```

### 可执行文件结构(ELF)

经过上一小节的探索，我们已经使用 `go build`命令将我们的 `main.go`代码编译成了机器可执行的文件 `main_bin`

关于这个可执行文件结构的分析比较复杂，不在本文档的讨论范围中，可以参考[Go 可执行文件结构深度分析](./executable-structure-analysis.md)来进行深入研究

不过我们可以简单通过几个命令来了解文件中包含了什么

- 使用 `nm`命令查看符号表

  ```bash
  nm main_bin | grep -E "(rt0|main\.|runtime\.main|schedinit)"
  ```

  启动相关的符号：

  ```text
  0x46ce40  T  _rt0_amd64_linux      ← 程序入口 (Entry Point)
  0x469740  T  _rt0_amd64            ← 平台无关入口
  0x469760  T  runtime.rt0_go        ← Runtime 启动
  0x4361e0  T  runtime.schedinit     ← 调度器初始化
  0x434e60  T  runtime.main          ← Runtime main
  0x48f080  T  main.main             ← 用户 main
  ```

  ```bash
  nm main_bin | grep -E "runtime\.(g0|m0|allp)"
  ```

  全局变量相关符号：

  ```text
  0x552c80  B  runtime.g0      ← 主线程的 g0
  0x5538e0  B  runtime.m0      ← 主线程 M
  0x5726e0  B  runtime.allp    ← 所有 P 的数组
  ```
- 执行流程

  ```text
  文件结构                        运行时
  ─────────────────────────────────────────
  .text (0x401000)
  ├─ 0x46ce40: _rt0_amd64_linux  → Entry Point
  │                                  ↓
  ├─ 0x469740: _rt0_amd64        → 设置 argc, argv
  │                                  ↓
  ├─ 0x469760: runtime.rt0_go    → g0, m0 初始化
  │                                  ↓
  ├─ 0x4361e0: runtime.schedinit → 调度器初始化
  │                                  ↓
  └─ 0x434e60: runtime.main      → runtime.main
  									↓
  								main.main (0x48f080)
  ```

## 2. 世界开始前的工作

这一节我们讨论在用户代码执行前，Go语言的“世界初始化”操作，其中 `runtime.main`的启动是用户 `main`函数启动之前的关键。本节包含大量的汇编代码

**世界开始前启动链路**

```text
Linux Loader
    ↓
0x46ce40: _rt0_amd64_linux
    ↓ (jmp)
0x469740: _rt0_amd64
    ├─ 设置 argc, argv
    ↓ (jmp)
0x469760: runtime.rt0_go
    ├─ 初始化 g0 栈
    ├─ CPUID 检测
    ├─ 设置 TLS
    ├─ 连接 g0 ↔ m0
    ├─ call runtime.check
    ├─ call runtime.args
    ├─ call runtime.osinit
    ├─ call runtime.schedinit     ← 调度器初始化
    ├─ call runtime.newproc       ← 创建 main goroutine
    └─ call runtime.mstart        ← 启动调度（永不返回）
```

接下来可以使用两种方法来查看启动前的汇编代码，**反汇编可执行文件**和**查看go语言的源码**，我们来分别查看下

### 反汇编 `main_bin`

从上一章节我们知道了程序的入口地址是：0x46ce40  T  _rt0_amd64_linux，我们可以使用反汇编指令 `objdump`来一步步查看代码的执行情况

#### _rt0_amd64_linux: 程序入口

```bash
objdump -d main_bin --start-address=0x46ce40 --stop-address=0x46ce50 -M intel
```

输出结果如下：

```asm
main_bin:     file format elf64-x86-64


Disassembly of section .text:

000000000046ce40 <_rt0_amd64_linux>:
  46ce40:       e9 fb c8 ff ff          jmp    469740 <_rt0_amd64>
  46ce45:       cc                      int3   
  46ce46:       cc                      int3   
  46ce47:       cc                      int3   
  46ce48:       cc                      int3   
  46ce49:       cc                      int3   
  46ce4a:       cc                      int3   
  46ce4b:       cc                      int3   
  46ce4c:       cc                      int3   
  46ce4d:       cc                      int3   
  46ce4e:       cc                      int3   
  46ce4f:       cc                      int3   
```

可以看到指令只有简单的一条jump，于是根据jump的地址我们可以继续使用 `objdump`来追踪代码

#### _rt0_amd64: 设置参数

```bash
objdump -d main_bin --start-address=0x469740 --stop-address=0x469750 -M intel
```

输出如下：

```asm
main_bin:     file format elf64-x86-64


Disassembly of section .text:

0000000000469740 <_rt0_amd64>:
  469740:       48 8b 3c 24             mov    rdi,QWORD PTR [rsp]
  469744:       48 8d 74 24 08          lea    rsi,[rsp+0x8]
  469749:       e9 12 00 00 00          jmp    469760 <runtime.rt0_go.abi0>
  46974e:       cc                      int3   
  46974f:       cc                      int3   
```

这段汇编指令做了三件事：

1. 从栈上读取 `argc`到 `rdi`寄存器
2. 获取 `argv`地址到 `rsi`寄存器
3. 跳转到 `runtime.rt0_go`指令

我们继续执行 `objdump`

#### runtime.rt0_go: 核心初始化

```bash
objdump -d main_bin --start-address=0x469760 --stop-address=0x469890 -M intel
```

关键部分输出：

```asm
0000000000469760 <runtime.rt0_go.abi0>:
  ; 保存 argc 和 argv
  469760:	48 89 f8             	mov    rax,rdi                    # 保存 argc
  469763:	48 89 f3             	mov    rbx,rsi                    # 保存 argv
  469766:	48 83 ec 28          	sub    rsp,0x28                   # 分配栈空间
  46976a:	48 83 e4 f0          	and    rsp,0xfffffffffffffff0     # 对齐栈到 16 字节

  ; 初始化 g0 栈
  469778:	48 8d 3d 01 95 0e 00 	lea    rdi,[rip+0xe9501]          # rdi = &runtime.g0 (0x552c80)
  46977f:	48 8d 9c 24 00 00 ff ff lea    rbx,[rsp-0x10000]          # rbx = 栈底
  469787:	48 89 5f 10          	mov    QWORD PTR [rdi+0x10],rbx   # g0.stackguard0 = 栈底
  46978b:	48 89 5f 18          	mov    QWORD PTR [rdi+0x18],rbx   # g0.stackguard1 = 栈底
  46978f:	48 89 1f             	mov    QWORD PTR [rdi],rbx        # g0.stack.lo = 栈底
  469792:	48 89 67 08          	mov    QWORD PTR [rdi+0x8],rsp    # g0.stack.hi = 栈顶

  ; CPUID 检测（省略）
  469796:	b8 00 00 00 00       	mov    eax,0x0
  46979b:	0f a2                	cpuid

  ; 设置 TLS（Thread Local Storage）
  469812:	e8 a9 3e 00 00       	call   46d6c0 <runtime.settls>

  ; 连接 g0 和 m0
  469838:	48 8d 0d 41 94 0e 00 	lea    rcx,[rip+0xe9441]          # rcx = &runtime.g0
  46983f:	64 48 89 0c 25 f8 ff ff ff mov QWORD PTR fs:0xfffffffffffffff8,rcx  # TLS 设置为 g0
  469848:	48 8d 05 91 a0 0e 00 	lea    rax,[rip+0xea091]          # rax = &runtime.m0
  46984f:	48 89 08             	mov    QWORD PTR [rax],rcx        # m0.g0 = &g0
  469852:	48 89 41 30          	mov    QWORD PTR [rcx+0x30],rax   # g0.m = &m0

  ; Runtime 初始化函数调用链
  469856:	fc                   	cld
  469857:	e8 64 46 00 00       	call   46dec0 <runtime.check>          # 运行时检查
  46986d:	e8 0e 46 00 00       	call   46de80 <runtime.args>           # 处理命令行参数
  469872:	e8 29 44 00 00       	call   46dca0 <runtime.osinit>         # OS 初始化
  469877:	e8 64 45 00 00       	call   46dde0 <runtime.schedinit>      # 调度器初始化

  ; 创建 main goroutine
  46987c:	48 8d 05 1d a8 06 00 	lea    rax,[rip+0x6a81d]          # rax = runtime.mainPC
  469883:	50                   	push   rax
  469884:	e8 b7 45 00 00       	call   46de40 <runtime.newproc>        # 创建 main goroutine

  ; 启动调度器（永不返回）
  46988a:	e8 71 00 00 00       	call   469900 <runtime.mstart>         # 启动 M
  46988f:	e8 2c 1e 00 00       	call   46b6c0 <runtime.abort>          # 不应该到达这里
```

### 查看go源码

我们的源码目录: /root/rem/go-master/src
本小节的所有命令操作都是默认在此目录下进行的

**Tips: 善用 `grep -rn "pattern" path`和 `sed -n 'line1,line2p' file`来查找源码**

首先我们查看程序启动的汇编文件 `rt0_linux_amd64`

```bash
find ./ -name "rt0_linux_amd64*" 

# 输出 ./runtime/rt0_linux_amd64.s
# 使用cat 命令查看汇编代码
cat ./runtime/rt0_linux_amd64.s
```

可以看到代码里有两个启动函数：

```asm
#include "textflag.h"

TEXT _rt0_amd64_linux(SB),NOSPLIT,$-8
	JMP     _rt0_amd64(SB)

TEXT _rt0_amd64_linux_lib(SB),NOSPLIT,$0
	JMP     _rt0_amd64_lib(SB)
```

根据我们之前反汇编得到信息，我们知道代码的启动函数是 `_rt0_amd64_linux`，于是我们接着查找 `_rt0_amd64(SB)`函数的定义

```bash
grep -rn "TEXT _rt0_amd64(SB)" ./

# ./runtime/asm_amd64.s:15:TEXT _rt0_amd64(SB),NOSPLIT,$-8

# 查看指定行
sed -n '15,30p' ./runtime/asm_amd64.s
```

得到汇编代码

```asm
TEXT _rt0_amd64(SB),NOSPLIT,$-8
	MOVQ    0(SP), DI       // argc
	LEAQ    8(SP), SI       // argv
	JMP     runtime·rt0_go(SB)
```

同理，接着查看 `runtime·rt0_go(SB)`代码定义，关于这个汇编函数的具体阅读这里就不放出来，下面简单介绍下这个函数的功能：

```text
  runtime·rt0_go
      │
      ├─ 1️⃣ 保存 argc, argv
      │   └─ 对齐栈到 16 字节
      │
      ├─ 2️⃣ 初始化 g0 栈⭐
      │   ├─ g0.stack.lo = SP - 64KB
      │   ├─ g0.stack.hi = SP       
      │   └─ g0.stackguard0/1 = SP - 64KB
      │
      ├─ 3️⃣ CPU 特性检测
      │   ├─ CPUID 检测 Intel/AMD
      │   └─ 保存 CPU 版本信息
      │
      ├─ 4️⃣ CGO 初始化
      │   └─ 调用 _cgo_init
      │
      ├─ 5️⃣ 设置 TLS
      │   ├─ 调用 runtime·settls
      │   └─ 测试 TLS 是否工作
      │
      ├─ 6️⃣ 连接 g0 ↔ m0
      │   ├─ m0.g0 = &g0
      │   ├─ g0.m = &m0
      │   └─ TLS.g = g0
      │
      ├─ 7️⃣ CPU 微架构检查
      │   └─ 检查 GOAMD64 要求的特性
      │
      ├─ 8️⃣ Runtime 初始化
      │   ├─ runtime·check()      // 一致性检查
      │   ├─ runtime·args()       // 命令行参数
      │   ├─ runtime·osinit()     // OS 初始化
      │   └─ runtime·schedinit()  // 调度器初始化 ⭐⭐⭐
      │
      ├─ 9️⃣ 创建 main goroutine
      │   └─ runtime·newproc(runtime.main)
      │
      └─ 🔟 启动调度器
          └─ runtime·mstart() → 永不返回！
              └─ schedule() 循环
                  ├─ findrunnable()  // 找可运行的 g
                  ├─ execute(g)      // 执行 g
                  └─ gogo()          // 切换到 g 的栈
```

关于 `runtime`的几个初始化函数都可以在源码包中找到

```bash
grep -rn "func check(" ./runtime
```

**runtime.check()**

- 检查内部数据结构的大小和对齐
- 确保编译器和运行期的一致性

**runtime.args()**

- 解析程序启动时的参数，提取出辅助向量auxv
- 从操作系统获取运行时所需要的基础系统信息

**runtime.osinit()**

```go
func osinit() {
	numCPUStartup = getCPUCount() // 计算可用的CPU核心数
	physHugePageSize = getHugePageSize()   // 获取系统的大页大小
	vgetrandomInit()  // 初始化 vgetrandom, 用于 Go 运行时需要随机数的场景
}
```

**runtime.schedinit()**
调度器初始化的核心函数，后续详细分析

**runtime.newproc()**
函数功能：
newproc 是 Go 编译器在遇到 go 语句时生成的调用目标函数。它负责创建一个新的 goroutine 并将其加入调度队列。

主要步骤：

- 获取当前 goroutine (gp)
- 获取调用者的 PC (用于调试/追踪)
- 切换到系统栈执行核心创建逻辑
- 将新 goroutine 放入运行队列
- 必要时唤醒一个空闲的 P

```go
// Create a new g running fn.
// Put it on the queue of g's waiting to run.
// The compiler turns a go statement into a call to this.
func newproc(fn *funcval) {
	gp := getg()
	if goexperiment.RuntimeSecret && gp.secret > 0 {
		panic("goroutine spawned while running in secret mode")
	}

	pc := sys.GetCallerPC()
	systemstack(func() {
		newg := newproc1(fn, gp, pc, false, waitReasonZero)

		pp := getg().m.p.ptr()
		runqput(pp, newg, true)

		if mainStarted {
			wakep()
		}
	})
}
```

[ ] `newproc1`是真正创建goroutine的函数，后续我们要对这个函数进行详尽分析

**runtime.mstart()**
本质上是一段汇编代码，调用了 `runtime.mstart0()`，我们可以查看后者的go源码

`mstart0()` 是新创建的 M (机器线程) 的 Go 语言入口点，在汇编函数 mstart 之后被调用。它负责初始化 M 的运行环境，最终进入调度循环。

主要步骤

- 获取当前g0
- 初始化g0栈边界
- 设置g0栈保护边界
- 调用 `mstart1`进行核心初始化并进入调度循环

[ ] `mstart1`是M核心初始化并进入调度循环的函数，后续我们要对这个函数进行详尽分析

至此我们分析完了一段简单的go代码是如何经过**点火装配（编译）** 以及 **点火（汇编启动）** 的过程，下面让我们用一个简单的流程图来展示一下这个过程。

![img](./img/all_line.svg)