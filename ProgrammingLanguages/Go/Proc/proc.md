# Go世界：从开始到终结

我们将从以下这个简单代码入手，去探究go语言从编译到运行到程序结束发生的一系列故事，让我们探索go世界的奥妙！

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

## 1. 程序启动

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

#### 链接阶段

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
