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

#### build_log.log 编译日志分析

##### 依赖树构建

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
