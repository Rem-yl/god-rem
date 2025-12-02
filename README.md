# 技术知识体系 - 完整路线图

> 基于你的学习目标重新组织的技术体系，与LEARNING_ROADMAP.md对应

---

## 📋 知识体系结构

### 1. 编程语言基础 (Programming Languages)

#### 1.1 Go 语言 ⭐⭐⭐⭐⭐
```
基础语法
├── 数据类型（基本类型、复合类型）
├── 控制流（if/for/switch）
├── 函数（defer、panic/recover）
└── 面向对象（struct、interface、方法）

进阶特性
├── 并发编程
│   ├── Goroutine
│   ├── Channel
│   ├── Select
│   └── Context
├── 错误处理
├── 包管理（go mod）
├── 反射（reflect）
└── 测试（testing、benchmark）

底层原理（Week 5-6学习）⭐⭐⭐⭐⭐
├── GPM调度模型
│   ├── G（Goroutine）
│   ├── P（Processor）
│   ├── M（Machine/Thread）
│   └── 调度流程
├── Channel实现原理
│   ├── hchan结构
│   ├── 发送接收流程
│   └── select实现
├── GC垃圾回收
│   ├── 三色标记法
│   ├── 写屏障（Write Barrier）
│   └── STW优化
├── 内存分配器（基于TCMalloc）
│   ├── span、mcentral、mheap
│   └── 小对象、大对象分配
├── Map/Slice/Interface底层
│   ├── map的hmap结构
│   ├── slice扩容机制
│   └── interface的eface/iface
└── 逃逸分析

标准库源码
├── net/http
├── sync包（Mutex、RWMutex、WaitGroup）
└── context包
```

#### 1.2 Python ⭐⭐⭐
```
基础语法
├── 数据类型
├── 函数与装饰器
├── 类与面向对象
└── 异常处理

进阶特性
├── 迭代器与生成器
├── 上下文管理器
├── 元类（Metaclass）
├── 多线程/多进程
└── 异步编程（asyncio）

底层原理
├── GIL全局解释器锁
├── 内存管理
├── 字节码与虚拟机
└── C扩展

常用框架
├── Web框架（Django、Flask、FastAPI）
├── 数据科学（NumPy、Pandas）
└── 机器学习（scikit-learn、PyTorch）
```

#### 1.3 C/C++ ⭐⭐⭐
```
C语言
├── 指针与内存管理
├── 结构体与联合体
├── 系统调用
└── 网络编程（socket）

C++
├── 面向对象（继承、多态、封装）
├── STL容器
│   ├── vector、list、deque
│   ├── map、set
│   └── unordered_map、unordered_set
├── 智能指针（shared_ptr、unique_ptr）
├── RAII资源管理
├── 模板与泛型编程
└── C++11/14/17新特性
```

---

### 2. 数据结构与算法 ⭐⭐⭐⭐⭐

#### 2.1 基础数据结构（Week 1-2）
```
线性结构
├── 数组与动态数组
│   ├── 连续内存存储
│   ├── 随机访问O(1)
│   ├── 动态扩容（Go slice）
│   └── LeetCode: 1(两数之和), 15(三数之和), 53(最大子数组和), 238(除自身以外数组的乘积)
├── 链表
│   ├── 单链表
│   │   └── LeetCode: 206(反转链表), 21(合并两个有序链表), 141(环形链表), 160(相交链表)
│   ├── 双向链表 ⭐⭐⭐⭐⭐（LRU Cache必备）
│   │   └── LeetCode: 146(LRU缓存), 707(设计链表)
│   └── 循环链表
│       └── LeetCode: 708(循环有序列表的插入)
├── 栈（Stack）
│   ├── 数组实现
│   ├── 链表实现
│   ├── 应用（函数调用栈、表达式求值）
│   └── LeetCode: 20(有效的括号), 155(最小栈), 739(每日温度), 84(柱状图中最大的矩形)
└── 队列（Queue）
    ├── 普通队列
    │   └── LeetCode: 232(用栈实现队列), 622(设计循环队列)
    ├── 循环队列
    │   └── LeetCode: 622(设计循环队列)
    ├── 优先队列（堆实现）
    │   └── LeetCode: 23(合并K个升序链表), 215(数组中的第K个最大元素), 295(数据流的中位数)
    └── 双端队列（Deque）
        └── LeetCode: 239(滑动窗口最大值), 641(设计循环双端队列)

哈希表 ⭐⭐⭐⭐⭐
├── 哈希函数设计
│   ├── 除法散列
│   ├── 乘法散列
│   └── 一致性哈希
├── 冲突解决
│   ├── 链地址法（Chaining）
│   └── 开放寻址（Linear Probing、Quadratic Probing）
├── 扩容与Rehash
├── 基础练习
│   └── LeetCode: 1(两数之和), 49(字母异位词分组), 128(最长连续序列), 383(赎金信)
└── 应用
    ├── LRU Cache ⭐⭐⭐⭐⭐（Week 1 Day 1）
    │   └── LeetCode: 146(LRU缓存机制)
    ├── LFU Cache ⭐⭐⭐⭐
    │   └── LeetCode: 460(LFU缓存)
    └── 一致性哈希（分布式缓存）

树结构
├── 二叉树基础
│   ├── 遍历（前中后序、层序）⭐⭐⭐⭐⭐
│   │   └── LeetCode: 94(中序遍历), 144(前序遍历), 145(后序遍历), 102(层序遍历)
│   ├── 递归与迭代实现
│   │   └── LeetCode: 104(最大深度), 111(最小深度), 226(翻转二叉树), 101(对称二叉树)
│   └── Morris遍历（O(1)空间）
│       └── LeetCode: 94(中序遍历-Morris), 99(恢复二叉搜索树)
├── 二叉搜索树（BST）⭐⭐⭐⭐
│   ├── 查找、插入、删除
│   │   └── LeetCode: 700(BST中的搜索), 701(BST中的插入), 450(删除BST中的节点)
│   ├── 中序遍历有序性
│   │   └── LeetCode: 98(验证BST), 230(BST中第K小的元素), 538(BST转换为累加树)
│   └── 退化问题
│       └── LeetCode: 108(有序数组转BST), 1382(平衡BST)
├── 平衡树 ⭐⭐⭐⭐⭐
│   ├── AVL树
│   │   ├── 平衡因子
│   │   ├── 四种旋转（LL、RR、LR、RL）
│   │   ├── 严格平衡
│   │   └── LeetCode: 110(平衡二叉树判断), 1382(平衡BST)
│   ├── 红黑树 ⭐⭐⭐⭐⭐（Week 2）
│   │   ├── 5条性质
│   │   ├── 插入修复
│   │   ├── 删除修复
│   │   ├── Linux内核应用（CFS、虚拟内存）
│   │   └── Java TreeMap/TreeSet
│   └── B树/B+树 ⭐⭐⭐⭐⭐（数据库索引）
│       ├── 多路搜索树
│       ├── 节点分裂与合并
│       ├── B+树叶子链表
│       ├── 磁盘I/O优化
│       └── MySQL索引实现
├── 跳表（Skip List）⭐⭐⭐⭐⭐（Week 1）
│   ├── 多层索引结构
│   ├── 概率化平衡（随机层数）
│   ├── 并发优势（无需全局锁）
│   ├── Redis Sorted Set实现
│   └── LeetCode: 1206(设计跳表)
├── Trie树（前缀树）⭐⭐⭐⭐（Week 1）
│   ├── 前缀匹配
│   ├── 字典树实现
│   ├── 压缩Trie（Radix Tree）
│   ├── 应用（敏感词过滤、自动补全、IP路由）
│   └── LeetCode: 208(实现Trie), 211(添加与搜索单词), 212(单词搜索II), 648(单词替换)
└── 堆（Heap）⭐⭐⭐⭐
    ├── 最小堆、最大堆
    ├── 堆化（heapify）
    ├── 堆排序
    ├── 应用（优先队列、Top K问题）
    └── LeetCode: 215(第K个最大元素), 347(前K个高频元素), 23(合并K个链表), 295(数据流中位数)

高级数据结构
├── 布隆过滤器（Bloom Filter）⭐⭐⭐⭐⭐
│   ├── 概率数据结构
│   ├── 多个哈希函数
│   ├── 误判率计算
│   ├── 应用（缓存穿透、URL去重）
│   └── Counting Bloom Filter（支持删除）
├── 并查集（Union-Find）⭐⭐⭐⭐
│   ├── 路径压缩
│   ├── 按秩合并
│   ├── 应用（连通性问题、最小生成树）
│   └── LeetCode: 200(岛屿数量), 547(省份数量), 684(冗余连接), 721(账户合并), 1971(寻找图中是否存在路径)
├── 线段树（Segment Tree）⭐⭐⭐
│   ├── 区间查询与更新
│   └── LeetCode: 307(区域和检索-数组可修改), 218(天际线问题), 715(Range模块)
├── 树状数组（Fenwick Tree）⭐⭐⭐
│   ├── 前缀和查询
│   └── LeetCode: 307(区域和检索), 315(计算右侧小于当前元素的个数)
└── LSM-Tree ⭐⭐⭐⭐⭐
    ├── Log-Structured Merge-Tree
    ├── MemTable + SSTable
    ├── Compaction策略
    └── 应用（RocksDB、LevelDB、Cassandra）
```

#### 2.2 算法专题（Week 3-4）
```
排序算法 ⭐⭐⭐⭐
├── 比较排序
│   ├── 快速排序 ⭐⭐⭐⭐⭐
│   │   ├── 分治思想
│   │   ├── 三路快排
│   │   ├── 时间复杂度：平均O(n log n)、最坏O(n²)
│   │   └── LeetCode: 912(排序数组), 75(颜色分类-三路快排)
│   ├── 归并排序 ⭐⭐⭐⭐
│   │   ├── 分治 + 合并
│   │   ├── 稳定排序
│   │   └── LeetCode: 912(排序数组), 148(排序链表), 315(计算右侧小于当前元素的个数)
│   ├── 堆排序 ⭐⭐⭐⭐
│   │   ├── 原地排序、不稳定
│   │   └── LeetCode: 912(排序数组), 215(数组中的第K个最大元素)
│   ├── 插入排序、选择排序
│   │   └── LeetCode: 147(对链表进行插入排序)
│   └── 希尔排序
└── 非比较排序
    ├── 计数排序
    │   └── LeetCode: 1122(数组的相对排序)
    ├── 桶排序
    │   └── LeetCode: 164(最大间距), 220(存在重复元素III)
    └── 基数排序

搜索算法 ⭐⭐⭐⭐⭐
├── 二分查找 ⭐⭐⭐⭐⭐
│   ├── 基础二分
│   │   └── LeetCode: 704(二分查找), 35(搜索插入位置), 69(x的平方根)
│   ├── 左边界、右边界
│   │   └── LeetCode: 34(在排序数组中查找元素的第一个和最后一个位置), 278(第一个错误的版本)
│   └── 旋转数组二分
│       └── LeetCode: 33(搜索旋转排序数组), 81(搜索旋转排序数组II), 153(寻找旋转排序数组中的最小值)
├── DFS深度优先搜索 ⭐⭐⭐⭐⭐
│   ├── 递归实现
│   │   └── LeetCode: 104(二叉树的最大深度), 543(二叉树的直径), 124(二叉树中的最大路径和)
│   ├── 栈实现
│   │   └── LeetCode: 94(二叉树的中序遍历-迭代)
│   └── 回溯剪枝
│       └── LeetCode: 200(岛屿数量), 695(岛屿的最大面积), 130(被围绕的区域)
└── BFS广度优先搜索 ⭐⭐⭐⭐⭐
    ├── 队列实现
    │   └── LeetCode: 102(二叉树的层序遍历), 199(二叉树的右视图), 103(二叉树的锯齿形层序遍历)
    ├── 最短路径
    │   └── LeetCode: 542(01矩阵), 994(腐烂的橘子), 1091(二进制矩阵中的最短路径)
    └── 拓扑排序
        └── LeetCode: 207(课程表), 210(课程表II)

动态规划（Dynamic Programming）⭐⭐⭐⭐⭐
├── 基础理论
│   ├── 最优子结构
│   ├── 重叠子问题
│   ├── 状态转移方程
│   └── LeetCode: 70(爬楼梯), 509(斐波那契数), 1137(第N个泰波那契数)
├── 背包问题 ⭐⭐⭐⭐⭐
│   ├── 0-1背包
│   │   └── LeetCode: 416(分割等和子集), 494(目标和), 1049(最后一块石头的重量II)
│   ├── 完全背包
│   │   └── LeetCode: 322(零钱兑换), 518(零钱兑换II), 279(完全平方数)
│   ├── 多重背包
│   └── 二维背包
│       └── LeetCode: 474(一和零), 879(盈利计划)
├── 序列DP ⭐⭐⭐⭐⭐
│   ├── 最长公共子序列（LCS）
│   │   └── LeetCode: 1143(最长公共子序列), 583(两个字符串的删除操作), 712(两个字符串的最小ASCII删除和)
│   ├── 最长递增子序列（LIS）
│   │   └── LeetCode: 300(最长递增子序列), 673(最长递增子序列的个数), 354(俄罗斯套娃信封问题)
│   ├── 编辑距离
│   │   └── LeetCode: 72(编辑距离), 115(不同的子序列), 44(通配符匹配)
│   └── 回文子串/子序列
│       └── LeetCode: 5(最长回文子串), 516(最长回文子序列), 647(回文子串), 131(分割回文串)
├── 区间DP ⭐⭐⭐⭐
│   ├── 石子合并
│   │   └── LeetCode: 1000(合并石头的最低成本)
│   ├── 矩阵链乘法
│   └── 戳气球
│       └── LeetCode: 312(戳气球), 1039(多边形三角剖分的最低得分)
├── 状态压缩DP ⭐⭐⭐⭐
│   ├── 位运算表示状态
│   ├── 旅行商问题（TSP）
│   │   └── LeetCode: 847(访问所有节点的最短路径)
│   └── 子集枚举
│       └── LeetCode: 78(子集), 90(子集II), 698(划分为k个相等的子集)
├── 树形DP ⭐⭐⭐⭐
│   ├── 树的遍历
│   │   └── LeetCode: 337(打家劫舍III), 124(二叉树中的最大路径和)
│   ├── 换根DP
│   │   └── LeetCode: 834(树中距离之和)
│   └── 树上背包
└── 数位DP、概率DP
    └── LeetCode: 233(数字1的个数), 902(最大为N的数字组合)

图算法 ⭐⭐⭐⭐⭐
├── 图的表示
│   ├── 邻接矩阵
│   ├── 邻接表
│   └── 边集数组
├── 遍历
│   ├── DFS（连通性、环检测）
│   │   └── LeetCode: 547(省份数量), 200(岛屿数量), 695(岛屿的最大面积)
│   └── BFS（最短路径）
│       └── LeetCode: 1091(二进制矩阵中的最短路径), 127(单词接龙)
├── 最短路径 ⭐⭐⭐⭐⭐
│   ├── Dijkstra算法（单源最短路）
│   │   └── LeetCode: 743(网络延迟时间), 787(K站中转内最便宜的航班), 1514(概率最大的路径)
│   ├── Bellman-Ford算法（负权边）
│   │   └── LeetCode: 787(K站中转内最便宜的航班)
│   ├── Floyd-Warshall算法（全源最短路）
│   │   └── LeetCode: 1334(阈值距离内邻居最少的城市)
│   └── SPFA算法
├── 最小生成树 ⭐⭐⭐⭐
│   ├── Prim算法
│   │   └── LeetCode: 1584(连接所有点的最小费用)
│   └── Kruskal算法（并查集）
│       └── LeetCode: 1584(连接所有点的最小费用), 1631(最小体力消耗路径)
├── 拓扑排序 ⭐⭐⭐⭐
│   ├── Kahn算法（BFS）
│   │   └── LeetCode: 207(课程表), 210(课程表II), 310(最小高度树)
│   └── DFS实现
│       └── LeetCode: 207(课程表), 802(找到最终的安全状态)
└── 强连通分量 ⭐⭐⭐
    ├── Tarjan算法
    │   └── LeetCode: 1192(查找集群内的关键连接)
    └── Kosaraju算法

字符串算法 ⭐⭐⭐⭐
├── KMP算法 ⭐⭐⭐⭐⭐
│   ├── next数组
│   ├── 模式匹配
│   └── LeetCode: 28(找出字符串中第一个匹配项的下标), 459(重复的子字符串), 686(重复叠加字符串匹配)
├── Boyer-Moore算法
├── Rabin-Karp算法
│   └── LeetCode: 187(重复的DNA序列)
├── AC自动机 ⭐⭐⭐⭐
│   └── 多模式匹配
├── 后缀数组
└── 字符串哈希
    └── LeetCode: 1044(最长重复子串), 214(最短回文串)

贪心算法 ⭐⭐⭐⭐
├── 区间调度
│   └── LeetCode: 435(无重叠区间), 452(用最少数量的箭引爆气球), 56(合并区间)
├── 最优装载
│   └── LeetCode: 455(分发饼干), 135(分发糖果), 860(柠檬水找零)
└── 哈夫曼编码
    └── LeetCode: 1167(连接棒材的最低费用)

回溯算法 ⭐⭐⭐⭐
├── 全排列
│   └── LeetCode: 46(全排列), 47(全排列II), 31(下一个排列)
├── N皇后
│   └── LeetCode: 51(N皇后), 52(N皇后II)
├── 组合总和
│   └── LeetCode: 39(组合总和), 40(组合总和II), 216(组合总和III), 377(组合总和IV)
└── 剪枝优化
    └── LeetCode: 37(解数独), 22(括号生成), 93(复原IP地址), 131(分割回文串)

分治算法 ⭐⭐⭐⭐
├── 归并排序
│   └── LeetCode: 148(排序链表), 315(计算右侧小于当前元素的个数)
├── 快速排序
│   └── LeetCode: 912(排序数组), 215(数组中的第K个最大元素)
└── 大数乘法（Karatsuba）
    └── LeetCode: 241(为运算表达式设计优先级)

双指针技巧 ⭐⭐⭐⭐⭐
├── 快慢指针（链表环检测）
│   └── LeetCode: 141(环形链表), 142(环形链表II), 876(链表的中间结点), 234(回文链表)
├── 左右指针（两数之和、三数之和）
│   └── LeetCode: 167(两数之和II-输入有序数组), 15(三数之和), 16(最接近的三数之和), 18(四数之和)
└── 滑动窗口 ⭐⭐⭐⭐⭐
    ├── 最长无重复子串
    │   └── LeetCode: 3(无重复字符的最长子串), 159(至多包含两个不同字符的最长子串)
    ├── 最小覆盖子串
    │   └── LeetCode: 76(最小覆盖子串), 438(找到字符串中所有字母异位词), 567(字符串的排列)
    └── 固定/可变窗口
        └── LeetCode: 209(长度最小的子数组), 424(替换后的最长重复字符), 1004(最大连续1的个数III)
```

---

### 3. 计算机基础 ⭐⭐⭐⭐⭐

#### 3.1 操作系统（Week 7）⭐⭐⭐⭐⭐

> **学习建议**: 从基础概念开始，逐步深入到内核机制。每个主题都包含理论学习和实践项目。
>
> **推荐资源**:
> - 📚 入门: 《操作系统导论》(OSTEP) - 清晰易懂，适合零基础
> - 📚 进阶: 《深入理解计算机系统》(CSAPP) - 系统级编程
> - 📚 深入: 《Linux内核设计与实现》- 内核源码分析
> - 🎓 课程: MIT 6.S081 (Operating System Engineering)
> - 🎓 课程: 南京大学 操作系统：设计与实现 (蒋炎岩)

```
===== 第一阶段：基础概念 ⭐⭐ (2-3天) =====

1. 操作系统概述
├── 什么是操作系统
│   ├── 资源管理者
│   ├── 抽象层（硬件与应用之间）
│   └── 虚拟化（CPU、内存、I/O）
├── 操作系统启动流程
│   ├── BIOS/UEFI
│   ├── Bootloader
│   └── 内核初始化
└── 实践项目 ⭐
    ├── 使用虚拟机安装Linux系统
    ├── 查看系统启动日志：dmesg, journalctl
    └── 理解/proc和/sys文件系统

2. 进程基础概念 ⭐⭐
├── 什么是进程
│   ├── 程序 vs 进程
│   ├── 进程 = 代码 + 数据 + 状态
│   └── 进程的生命周期
├── 进程地址空间 ⭐⭐⭐⭐
│   ├── 代码段（Text）
│   ├── 数据段（Data/BSS）
│   ├── 堆（Heap）- 动态分配
│   ├── 栈（Stack）- 函数调用
│   └── 内存布局图
├── 进程控制块（PCB）
│   ├── PID（进程ID）
│   ├── 进程状态
│   ├── 寄存器上下文
│   └── 内存信息
└── 实践项目 ⭐⭐
    ├── C语言程序：打印进程地址空间布局
    │   ```c
    │   // 查看栈、堆、全局变量、代码段的地址
    │   #include <stdio.h>
    │   int global_var;
    │   int main() {
    │       int stack_var;
    │       int *heap_var = malloc(sizeof(int));
    │       printf("Stack: %p\n", &stack_var);
    │       printf("Heap:  %p\n", heap_var);
    │       printf("Data:  %p\n", &global_var);
    │       printf("Code:  %p\n", main);
    │   }
    │   ```
    ├── 使用 ps, top 查看进程信息
    ├── 查看进程内存映射：cat /proc/[pid]/maps
    └── 练习：编写多进程程序（fork）

3. 线程基础概念 ⭐⭐
├── 什么是线程
│   ├── 轻量级进程
│   ├── 进程内的执行单元
│   └── 共享进程地址空间
├── 进程 vs 线程
│   ├── 资源开销对比
│   ├── 切换成本对比
│   └── 通信方式对比
├── 线程实现模型
│   ├── 用户级线程（N:1）
│   ├── 内核级线程（1:1）
│   └── 混合模型（M:N）- Goroutine采用此模型
└── 实践项目 ⭐⭐
    ├── pthread多线程编程（C语言）
    ├── Go语言并发编程对比
    ├── 测量进程vs线程创建时间
    └── 使用 top -H 查看线程

===== 第二阶段：进程与线程管理 ⭐⭐⭐ (3-4天) =====

4. 进程状态与切换 ⭐⭐⭐
├── 进程状态模型
│   ├── 就绪（Ready）
│   ├── 运行（Running）
│   ├── 阻塞（Blocked/Waiting）
│   └── 僵尸（Zombie）、孤儿进程
├── 上下文切换 ⭐⭐⭐⭐
│   ├── 保存寄存器状态
│   ├── 切换页表（CR3寄存器）
│   ├── TLB刷新
│   └── 切换开销分析
├── 进程创建与终止
│   ├── fork() - 复制父进程
│   ├── exec() - 替换进程映像
│   ├── exit() - 进程终止
│   └── wait() - 等待子进程
└── 实践项目 ⭐⭐⭐
    ├── 编写多进程程序：fork + exec
    ├── 处理僵尸进程和孤儿进程
    ├── 使用 strace 追踪系统调用
    └── 测量上下文切换开销：vmstat, pidstat

5. 进程调度算法 ⭐⭐⭐⭐
├── 调度目标
│   ├── CPU利用率
│   ├── 吞吐量
│   ├── 响应时间
│   └── 公平性
├── 经典调度算法
│   ├── FCFS（先来先服务）⭐⭐
│   ├── SJF（最短作业优先）⭐⭐⭐
│   ├── 时间片轮转（Round Robin）⭐⭐⭐⭐
│   ├── 优先级调度 ⭐⭐⭐
│   └── 多级反馈队列 ⭐⭐⭐⭐
├── Linux CFS调度器 ⭐⭐⭐⭐⭐
│   ├── 完全公平调度（Completely Fair Scheduler）
│   ├── 虚拟运行时间（vruntime）
│   ├── 红黑树组织进程
│   ├── nice值与权重
│   └── 源码位置：kernel/sched/fair.c
└── 实践项目 ⭐⭐⭐⭐
    ├── 模拟实现调度算法（Python/Go）
    ├── 使用 nice/renice 调整进程优先级
    ├── 查看调度统计：cat /proc/sched_debug
    └── 编写CPU密集型vs IO密集型程序对比

6. 进程同步与通信 ⭐⭐⭐⭐
├── 临界区问题
│   ├── 竞态条件（Race Condition）
│   ├── 互斥访问
│   └── 原子操作
├── 同步原语 ⭐⭐⭐⭐⭐
│   ├── 互斥锁（Mutex）
│   │   ├── pthread_mutex
│   │   ├── 自旋锁 vs 睡眠锁
│   │   └── 锁的粒度
│   ├── 信号量（Semaphore）
│   │   ├── 二元信号量
│   │   ├── 计数信号量
│   │   └── 生产者-消费者问题
│   ├── 条件变量（Condition Variable）
│   │   └── wait/signal机制
│   └── 读写锁（RWLock）
│       └── 多读单写
├── 经典同步问题
│   ├── 生产者-消费者问题 ⭐⭐⭐⭐
│   ├── 读者-写者问题 ⭐⭐⭐⭐
│   ├── 哲学家就餐问题 ⭐⭐⭐⭐
│   └── 理发师问题
├── 进程间通信（IPC）⭐⭐⭐⭐
│   ├── 管道（Pipe）
│   │   ├── 匿名管道
│   │   └── 命名管道（FIFO）
│   ├── 消息队列（Message Queue）
│   ├── 共享内存（Shared Memory）⭐⭐⭐⭐
│   │   ├── 最快的IPC方式
│   │   ├── shmget/shmat/shmdt
│   │   └── 需要同步机制
│   ├── 信号（Signal）
│   └── Socket（网络通信）
└── 实践项目 ⭐⭐⭐⭐
    ├── 实现生产者-消费者（多种方式）
    ├── 编写多进程共享内存程序
    ├── 使用管道实现进程通信
    └── 解决哲学家就餐问题（避免死锁）

7. 死锁 ⭐⭐⭐⭐
├── 死锁四个必要条件
│   ├── 互斥（Mutual Exclusion）
│   ├── 持有并等待（Hold and Wait）
│   ├── 不可剥夺（No Preemption）
│   └── 循环等待（Circular Wait）
├── 死锁处理策略
│   ├── 死锁预防 - 破坏必要条件
│   ├── 死锁避免 - 银行家算法 ⭐⭐⭐⭐
│   ├── 死锁检测与恢复
│   └── 鸵鸟策略（忽略）
└── 实践项目 ⭐⭐⭐
    ├── 编写产生死锁的程序
    ├── 使用pstack/gdb检测死锁
    └── 实现银行家算法

===== 第三阶段：内存管理 ⭐⭐⭐⭐ (3-4天) =====

8. 内存管理基础 ⭐⭐⭐
├── 地址空间
│   ├── 物理地址 vs 虚拟地址
│   ├── 为什么需要虚拟内存
│   └── 地址翻译过程
├── 内存分配方式
│   ├── 连续分配
│   │   ├── 固定分区
│   │   └── 动态分区
│   ├── 碎片问题
│   │   ├── 内部碎片
│   │   └── 外部碎片
│   └── 伙伴系统（Buddy System）
└── 实践项目 ⭐⭐⭐
    ├── 查看系统内存：free -h, vmstat
    ├── 查看进程内存使用：pmap, smem
    └── 编写内存分配测试程序

9. 虚拟内存与分页 ⭐⭐⭐⭐⭐
├── 分页机制 ⭐⭐⭐⭐⭐
│   ├── 页（Page）与页框（Frame）
│   ├── 页表（Page Table）
│   ├── 地址翻译过程
│   │   ├── 虚拟地址 = 页号 + 页内偏移
│   │   └── 物理地址 = 帧号 + 页内偏移
│   └── 页表项（PTE）
│       ├── 有效位（Valid bit）
│       ├── 保护位（R/W/X）
│       └── 脏位、访问位
├── 多级页表 ⭐⭐⭐⭐⭐
│   ├── 为什么需要多级页表（节省空间）
│   ├── x86-64 四级页表
│   │   ├── PGD -> PUD -> PMD -> PTE
│   │   └── 48位虚拟地址
│   └── 页表遍历开销
├── TLB（快表）⭐⭐⭐⭐⭐
│   ├── Translation Lookaside Buffer
│   ├── 硬件缓存
│   ├── TLB命中 vs TLB未命中
│   └── TLB刷新时机
├── 缺页中断 ⭐⭐⭐⭐⭐
│   ├── 什么是缺页
│   ├── 缺页处理流程
│   ├── 按需分页（Demand Paging）
│   └── 预取（Prefetching）
└── 实践项目 ⭐⭐⭐⭐
    ├── 查看页表：cat /proc/[pid]/pagemap
    ├── 测量TLB大小：getconf PAGE_SIZE
    ├── 编写程序触发缺页中断
    └── 使用perf stat测量TLB miss率

10. 页面置换算法 ⭐⭐⭐⭐
├── 为什么需要页面置换
│   └── 物理内存有限
├── 页面置换算法
│   ├── 最优算法（OPT）⭐⭐ - 理论最优
│   ├── FIFO（先进先出）⭐⭐⭐
│   │   └── Belady异常
│   ├── LRU（最近最少使用）⭐⭐⭐⭐⭐
│   │   ├── 实现方式（链表、栈）
│   │   └── 近似LRU算法
│   ├── Clock算法 ⭐⭐⭐⭐
│   │   ├── 循环扫描
│   │   └── 访问位
│   ├── LFU（最不经常使用）⭐⭐⭐⭐
│   └── 工作集模型
├── Linux页面回收
│   ├── kswapd守护进程
│   ├── LRU链表
│   └── 页面类型（匿名页、文件页）
└── 实践项目 ⭐⭐⭐⭐
    ├── 模拟实现页面置换算法
    ├── 比较不同算法的缺页率
    └── 查看Linux内存回收：cat /proc/vmstat

11. 高级内存技术 ⭐⭐⭐⭐
├── mmap内存映射 ⭐⭐⭐⭐⭐
│   ├── 什么是mmap
│   ├── 文件映射 vs 匿名映射
│   ├── 私有映射 vs 共享映射
│   ├── 零拷贝技术
│   └── 应用场景
│       ├── 大文件读写
│       ├── 共享内存IPC
│       └── 动态链接库加载
├── 写时复制（COW）⭐⭐⭐⭐⭐
│   ├── Copy-On-Write
│   ├── fork()中的应用
│   ├── 延迟分配
│   └── 内存节省
├── 内存分配器
│   ├── malloc/free实现 ⭐⭐⭐⭐
│   │   ├── ptmalloc（glibc）
│   │   ├── tcmalloc（Google）
│   │   └── jemalloc
│   └── Go内存分配器 ⭐⭐⭐⭐⭐
│       ├── 基于TCMalloc
│       ├── span、mcentral、mheap
│       ├── 小对象、大对象分配
│       └── 逃逸分析
└── 实践项目 ⭐⭐⭐⭐⭐
    ├── 使用mmap读写大文件
    ├── 测量COW效果（fork前后内存）
    ├── 实现简单的内存池
    └── 分析Go程序内存分配（pprof）

===== 第四阶段：I/O与文件系统 ⭐⭐⭐⭐ (2-3天) =====

12. I/O系统基础 ⭐⭐⭐
├── I/O设备分类
│   ├── 块设备（磁盘）
│   └── 字符设备（键盘、网卡）
├── I/O控制方式
│   ├── 轮询（Polling）
│   ├── 中断（Interrupt）
│   └── DMA（Direct Memory Access）
└── 实践项目 ⭐⭐
    ├── 查看设备：ls /dev, lsblk
    └── 查看中断：cat /proc/interrupts

13. I/O模型 ⭐⭐⭐⭐⭐
├── 五种I/O模型对比
│   ├── 阻塞I/O ⭐⭐⭐
│   │   ├── 默认模式
│   │   ├── 进程阻塞等待
│   │   └── 简单但效率低
│   ├── 非阻塞I/O ⭐⭐⭐
│   │   ├── 立即返回
│   │   ├── 需要轮询
│   │   └── 浪费CPU
│   ├── I/O多路复用 ⭐⭐⭐⭐⭐
│   │   ├── select ⭐⭐⭐
│   │   │   ├── fd_set位图
│   │   │   ├── 1024限制
│   │   │   └── O(n)轮询
│   │   ├── poll ⭐⭐⭐
│   │   │   ├── pollfd数组
│   │   │   ├── 无数量限制
│   │   │   └── O(n)轮询
│   │   └── epoll ⭐⭐⭐⭐⭐
│   │       ├── Linux专有
│   │       ├── 红黑树管理fd
│   │       ├── 就绪链表
│   │       ├── O(1)复杂度
│   │       ├── epoll_create/ctl/wait
│   │       ├── LT vs ET模式
│   │       └── 应用：Nginx、Redis
│   ├── 信号驱动I/O ⭐⭐
│   └── 异步I/O ⭐⭐⭐⭐
│       ├── AIO
│       └── io_uring（新一代）
├── Reactor vs Proactor模式
│   ├── Reactor（同步非阻塞）
│   └── Proactor（异步）
└── 实践项目 ⭐⭐⭐⭐⭐
    ├── 实现echo服务器（5种模型）
    │   ├── 阻塞I/O版本
    │   ├── select版本
    │   ├── poll版本
    │   ├── epoll版本（LT和ET）
    │   └── Go netpoller版本
    ├── 性能对比测试（ab/wrk压测）
    └── 使用strace追踪系统调用

14. 文件系统 ⭐⭐⭐⭐
├── 文件系统基础
│   ├── inode（索引节点）⭐⭐⭐⭐
│   │   ├── 文件元数据
│   │   ├── inode编号
│   │   └── 数据块指针
│   ├── dentry（目录项）
│   │   └── 路径解析
│   ├── 文件描述符（FD）⭐⭐⭐⭐
│   │   ├── 进程私有
│   │   ├── 指向打开文件表
│   │   └── FD 0/1/2（stdin/stdout/stderr）
│   └── VFS（虚拟文件系统）
│       └── 统一接口
├── 磁盘文件系统
│   ├── ext4 ⭐⭐⭐⭐
│   │   ├── extent机制
│   │   └── 日志（Journal）
│   ├── XFS
│   └── Btrfs
├── 文件I/O优化
│   ├── 页缓存（Page Cache）⭐⭐⭐⭐⭐
│   │   ├── 读缓存
│   │   ├── 写缓存
│   │   └── dirty页回写
│   ├── 预读（Readahead）
│   ├── 零拷贝 ⭐⭐⭐⭐⭐
│   │   ├── sendfile
│   │   ├── splice
│   │   └── mmap + write
│   └── Direct I/O
│       └── 绕过Page Cache
└── 实践项目 ⭐⭐⭐⭐
    ├── 查看inode：ls -i, stat
    ├── 查看打开的文件：lsof, ls /proc/[pid]/fd
    ├── 测试零拷贝性能提升
    ├── 监控Page Cache：cat /proc/meminfo
    └── 实现简单的ls命令

===== 第五阶段：性能分析与优化 ⭐⭐⭐⭐⭐ (2-3天) =====

15. Linux性能分析工具 ⭐⭐⭐⭐⭐
├── CPU性能分析
│   ├── top/htop ⭐⭐⭐⭐
│   │   └── 实时监控
│   ├── mpstat ⭐⭐⭐
│   │   └── 多核CPU统计
│   ├── pidstat ⭐⭐⭐
│   │   └── 进程级统计
│   ├── perf ⭐⭐⭐⭐⭐
│   │   ├── perf top
│   │   ├── perf record/report
│   │   └── 性能事件采样
│   └── 火焰图（Flame Graph）⭐⭐⭐⭐⭐
│       └── 可视化性能瓶颈
├── 内存性能分析
│   ├── free ⭐⭐⭐
│   │   └── 内存使用统计
│   ├── vmstat ⭐⭐⭐⭐
│   │   └── 虚拟内存统计
│   ├── smem ⭐⭐⭐
│   │   └── 共享内存分析
│   ├── valgrind ⭐⭐⭐⭐
│   │   ├── 内存泄漏检测
│   │   └── massif（堆分析）
│   └── Go pprof ⭐⭐⭐⭐⭐
│       ├── heap profile
│       ├── goroutine profile
│       └── 可视化分析
├── I/O性能分析
│   ├── iostat ⭐⭐⭐⭐
│   │   └── 磁盘I/O统计
│   ├── iotop ⭐⭐⭐
│   │   └── 实时I/O监控
│   └── blktrace ⭐⭐⭐⭐
│       └── 块设备追踪
├── 网络性能分析
│   ├── iftop ⭐⭐⭐
│   ├── nethogs ⭐⭐⭐
│   ├── ss/netstat ⭐⭐⭐⭐
│   └── tcpdump ⭐⭐⭐⭐
├── 系统调用追踪
│   ├── strace ⭐⭐⭐⭐⭐
│   │   └── 追踪系统调用
│   ├── ltrace ⭐⭐⭐
│   │   └── 追踪库函数
│   └── bpftrace ⭐⭐⭐⭐⭐
│       └── eBPF动态追踪
└── 实践项目 ⭐⭐⭐⭐⭐
    ├── 性能分析实战
    │   ├── 找出CPU密集型进程
    │   ├── 定位内存泄漏
    │   ├── 分析磁盘I/O瓶颈
    │   └── 网络性能调优
    ├── 生成并分析火焰图
    ├── 使用perf分析程序热点
    └── Go程序性能优化实战

16. 系统优化技术 ⭐⭐⭐⭐
├── CPU优化
│   ├── 进程绑核（taskset）
│   ├── NUMA优化
│   └── CPU亲和性
├── 内存优化
│   ├── 大页（Huge Pages）
│   ├── NUMA策略
│   └── 内存限制（cgroup）
├── I/O优化
│   ├── I/O调度器
│   │   ├── noop
│   │   ├── deadline
│   │   └── cfq
│   ├── 文件系统调优
│   └── RAID配置
└── 实践项目 ⭐⭐⭐⭐
    ├── 调整系统参数（sysctl）
    ├── 配置大页内存
    └── I/O调度器性能对比

===== 综合实践项目 ⭐⭐⭐⭐⭐ =====

17. Mini OS项目（可选，但强烈推荐）
├── MIT 6.S081: xv6操作系统
│   ├── 实现系统调用
│   ├── 实现调度器
│   ├── 实现虚拟内存
│   └── 实现文件系统
├── 或：清华uCore OS实验
└── 或：自己实现简化版OS
    ├── Bootloader
    ├── 简单的内核
    ├── 进程管理
    └── 内存管理

18. 真实项目性能优化
├── Web服务器性能优化
│   ├── 多进程 vs 多线程 vs epoll
│   ├── 零拷贝应用
│   └── 连接池管理
└── 数据库性能优化
    ├── 缓冲池调优
    ├── I/O优化
    └── 并发控制
```

**学习路径总结**:
```
Week 7 建议学习顺序：
Day 1-2: 阶段一（基础概念）+ 简单实践
Day 3-4: 阶段二（进程线程管理）+ 编程实践
Day 5-6: 阶段三（内存管理）+ 虚拟内存实验
Day 7:   阶段四（I/O模型）+ epoll编程

后续深入（Week 8+）:
- 阶段五（性能分析）贯穿整个学习过程
- Mini OS项目（2-4周，课余时间）
```

#### 3.2 计算机网络（Week 8）⭐⭐⭐⭐⭐
```
网络分层模型
├── OSI七层模型
└── TCP/IP四层模型

链路层
├── MAC地址
├── ARP协议
└── 交换机工作原理

网络层
├── IP协议 ⭐⭐⭐⭐⭐
│   ├── IP地址分类
│   ├── 子网掩码
│   ├── CIDR
│   └── NAT
├── ICMP协议
│   └── ping、traceroute
└── 路由算法
    ├── 距离向量（RIP）
    └── 链路状态（OSPF）

传输层
├── TCP协议 ⭐⭐⭐⭐⭐
│   ├── 连接管理
│   │   ├── 三次握手 ⭐⭐⭐⭐⭐
│   │   │   ├── SYN、SYN-ACK、ACK
│   │   │   ├── 为什么是三次？
│   │   │   └── SYN洪水攻击
│   │   ├── 四次挥手 ⭐⭐⭐⭐⭐
│   │   │   ├── FIN、ACK流程
│   │   │   ├── TIME_WAIT状态（2MSL）
│   │   │   └── CLOSE_WAIT问题
│   │   └── TCP状态机
│   ├── 可靠传输 ⭐⭐⭐⭐⭐
│   │   ├── 序列号与确认号
│   │   ├── 超时重传
│   │   ├── 快速重传
│   │   └── SACK（选择性确认）
│   ├── 流量控制 ⭐⭐⭐⭐⭐
│   │   ├── 滑动窗口
│   │   ├── 接收窗口
│   │   └── 零窗口探测
│   ├── 拥塞控制 ⭐⭐⭐⭐⭐
│   │   ├── 慢启动（Slow Start）
│   │   ├── 拥塞避免
│   │   ├── 快速恢复
│   │   └── BBR算法
│   ├── TCP优化 ⭐⭐⭐⭐
│   │   ├── Nagle算法
│   │   ├── 延迟确认
│   │   ├── TCP_NODELAY
│   │   └── sysctl参数调优
│   └── TCP粘包/拆包
│       └── 应用层协议设计
└── UDP协议 ⭐⭐⭐⭐
    ├── 无连接、不可靠
    ├── 应用场景（DNS、视频流）
    └── QUIC协议

应用层协议
├── HTTP ⭐⭐⭐⭐⭐
│   ├── HTTP/1.0
│   │   └── 短连接
│   ├── HTTP/1.1 ⭐⭐⭐⭐⭐
│   │   ├── 持久连接（Keep-Alive）
│   │   ├── 管道化（Pipelining）
│   │   ├── 请求方法（GET、POST、PUT、DELETE等）
│   │   ├── 状态码（2xx、3xx、4xx、5xx）
│   │   ├── 缓存机制（ETag、Cache-Control）
│   │   └── 分块传输（Chunked）
│   ├── HTTP/2 ⭐⭐⭐⭐
│   │   ├── 多路复用（Multiplexing）
│   │   ├── 头部压缩（HPACK）
│   │   ├── Server Push
│   │   └── 二进制分帧
│   └── HTTP/3
│       └── 基于QUIC
├── HTTPS ⭐⭐⭐⭐⭐
│   ├── TLS/SSL握手 ⭐⭐⭐⭐⭐
│   │   ├── 对称加密
│   │   ├── 非对称加密
│   │   ├── 证书验证
│   │   └── 握手流程
│   ├── TLS 1.2 vs TLS 1.3
│   └── HTTPS优化（Session Resume、OCSP Stapling）
├── DNS ⭐⭐⭐⭐
│   ├── 域名解析流程
│   ├── 递归查询 vs 迭代查询
│   ├── DNS缓存
│   └── DNS劫持与防护
├── WebSocket ⭐⭐⭐⭐
│   ├── 全双工通信
│   ├── 握手过程
│   └── 应用场景（实时通信）
└── SSH ⭐⭐⭐⭐
    ├── 公钥认证
    └── 端口转发

网络工具 ⭐⭐⭐⭐
├── tcpdump
│   └── 抓包分析
├── wireshark ⭐⭐⭐⭐
│   └── 协议分析
├── netstat/ss
│   └── 连接状态查看
├── curl/wget
│   └── HTTP请求测试
└── nmap
    └── 端口扫描

高性能网络编程 ⭐⭐⭐⭐
├── Reactor模式
├── Proactor模式
├── 零拷贝（Zero Copy）
├── DPDK（数据平面开发套件）
├── XDP/eBPF
└── QUIC协议（基于UDP）
```

#### 3.3 Linux系统 ⭐⭐⭐⭐
```
Shell编程 ⭐⭐⭐⭐
├── bash脚本基础
│   ├── 变量
│   ├── 条件判断
│   ├── 循环
│   └── 函数
├── 文本处理 ⭐⭐⭐⭐⭐
│   ├── grep/egrep（正则匹配）
│   ├── sed（流编辑器）
│   ├── awk（文本分析）
│   ├── cut、sort、uniq
│   └── 管道与重定向
└── 实用脚本
    ├── 日志分析
    ├── 自动化部署
    └── 定时任务（cron）

进程管理 ⭐⭐⭐⭐
├── ps（进程查看）
├── top/htop（实时监控）
├── kill/killall（信号发送）
├── nohup/screen/tmux（后台运行）
└── systemd/systemctl（服务管理）

性能监控工具 ⭐⭐⭐⭐⭐
├── CPU监控
│   ├── top/htop
│   ├── mpstat
│   └── pidstat
├── 内存监控
│   ├── free
│   ├── vmstat
│   └── smem
├── IO监控
│   ├── iostat
│   ├── iotop
│   └── blktrace
├── 网络监控
│   ├── iftop
│   ├── nethogs
│   └── ss/netstat
└── 综合监控
    ├── sar（历史数据）
    ├── dstat
    └── glances

系统调试 ⭐⭐⭐⭐
├── strace（系统调用跟踪）
├── ltrace（库函数跟踪）
├── gdb（调试器）
├── perf（性能分析）
└── valgrind（内存检测）

正则表达式 ⭐⭐⭐⭐⭐
├── 基础语法
│   ├── 字符类 []
│   ├── 量词（*、+、?、{n,m}）
│   └── 锚点（^、$）
├── 高级特性
│   ├── 分组与捕获 ()
│   ├── 非捕获组 (?:)
│   ├── 前瞻/后顾 (?=)(?<=)
│   └── 贪婪 vs 非贪婪
└── 实际应用
    ├── 日志解析
    ├── 数据提取
    └── 输入验证
```

---

### 4. 数据库系统 ⭐⭐⭐⭐⭐

#### 4.1 数据库基础理论 ⭐⭐⭐⭐⭐

```
关系模型
├── 关系代数
├── 关系演算
├── 函数依赖
└── 范式理论（1NF、2NF、3NF、BCNF）

事务处理 ⭐⭐⭐⭐⭐
├── ACID特性
│   ├── 原子性（Atomicity）
│   ├── 一致性（Consistency）
│   ├── 隔离性（Isolation）
│   └── 持久性（Durability）
├── 事务隔离级别 ⭐⭐⭐⭐⭐
│   ├── 读未提交（Read Uncommitted）
│   ├── 读已提交（Read Committed）
│   ├── 可重复读（Repeatable Read）
│   └── 串行化（Serializable）
├── 并发问题
│   ├── 脏读（Dirty Read）
│   ├── 不可重复读（Non-Repeatable Read）
│   ├── 幻读（Phantom Read）
│   └── 丢失更新
└── MVCC（多版本并发控制）⭐⭐⭐⭐⭐
    ├── Read View
    ├── 版本链
    ├── 快照读 vs 当前读
    └── undo log

锁机制 ⭐⭐⭐⭐⭐
├── 锁的粒度
│   ├── 表锁
│   ├── 页锁
│   └── 行锁 ⭐⭐⭐⭐⭐
├── 锁的类型
│   ├── 共享锁（S锁）
│   ├── 排他锁（X锁）
│   ├── 意向锁（IS、IX）
│   └── 间隙锁（Gap Lock）⭐⭐⭐⭐
├── 死锁
│   ├── 死锁检测
│   ├── 死锁预防
│   └── 超时回滚
└── 乐观锁 vs 悲观锁

索引原理 ⭐⭐⭐⭐⭐
├── 索引类型
│   ├── B+树索引 ⭐⭐⭐⭐⭐
│   │   ├── 多路平衡树
│   │   ├── 叶子节点存储数据
│   │   ├── 叶子节点链表
│   │   └── 为什么用B+树而非红黑树？
│   ├── 哈希索引
│   │   ├── 等值查询快
│   │   └── 不支持范围查询
│   ├── 全文索引
│   └── 空间索引（R-Tree）
├── 索引结构
│   ├── 聚簇索引 vs 非聚簇索引
│   ├── 主键索引 vs 二级索引
│   └── 覆盖索引
├── 索引优化
│   ├── 最左前缀原则
│   ├── 索引下推（ICP）
│   ├── 索引合并
│   └── 索引失效场景
└── 索引设计原则

查询处理 ⭐⭐⭐⭐
├── 查询解析
│   ├── 词法分析
│   ├── 语法分析
│   └── 语义分析
├── 查询优化 ⭐⭐⭐⭐⭐
│   ├── 基于规则的优化（RBO）
│   ├── 基于代价的优化（CBO）
│   ├── 连接算法
│   │   ├── 嵌套循环连接
│   │   ├── 哈希连接
│   │   └── 排序合并连接
│   └── 执行计划分析
│       ├── EXPLAIN详解
│       ├── type类型（system、const、ref、range、index、all）
│       └── Extra信息
└── 慢查询优化
    ├── 慢查询日志
    ├── 查询分析工具
    └── 索引优化策略

日志系统 ⭐⭐⭐⭐⭐
├── WAL（Write-Ahead Logging）
│   └── 预写日志原理
├── Redo Log（重做日志）
│   ├── 崩溃恢复
│   └── 循环写入
├── Undo Log（回滚日志）
│   ├── 事务回滚
│   └── MVCC实现
└── Binlog（二进制日志）
    ├── Statement模式
    ├── Row模式
    └── Mixed模式

存储引擎架构
├── 存储引擎接口
├── 缓冲池（Buffer Pool）
│   ├── LRU管理
│   └── 脏页刷新
├── 数据页结构
└── 表空间管理

分布式数据库理论
├── 分片（Sharding）
│   ├── 垂直分片
│   ├── 水平分片
│   └── 一致性哈希
├── 复制（Replication）
│   ├── 主从复制
│   ├── 多主复制
│   └── 无主复制
└── 分布式事务
    ├── 两阶段提交（2PC）
    ├── 三阶段提交（3PC）
    ├── TCC
    └── Saga模式
```

#### 4.2 关系型数据库实现

##### 4.2.1 MySQL ⭐⭐⭐⭐⭐
```
架构与原理
├── 架构层次
│   ├── 连接层
│   ├── 服务层（SQL解析、优化器）
│   ├── 引擎层
│   └── 存储层
├── 存储引擎 ⭐⭐⭐⭐⭐
│   ├── InnoDB ⭐⭐⭐⭐⭐
│   │   ├── 聚簇索引
│   │   ├── 事务支持（ACID）
│   │   ├── MVCC实现
│   │   ├── 行锁
│   │   ├── 崩溃恢复
│   │   └── 自适应哈希索引
│   ├── MyISAM
│   │   ├── 非聚簇索引
│   │   ├── 表锁
│   │   └── 不支持事务
│   └── Memory、Archive等
└── 查询缓存（MySQL 8.0已移除）

索引实现
├── InnoDB B+树索引
│   ├── 主键索引（聚簇）
│   ├── 二级索引（非聚簇）
│   └── 回表查询
├── 联合索引
│   └── 最左前缀匹配
└── 索引统计信息

高可用方案 ⭐⭐⭐⭐⭐
├── 主从复制 ⭐⭐⭐⭐⭐
│   ├── binlog复制
│   ├── 异步复制 vs 半同步复制
│   ├── GTID复制
│   └── 主从延迟问题
├── 读写分离
│   ├── 应用层实现
│   ├── 中间件实现（MyCat、ProxySQL）
│   └── 一致性问题
├── 高可用架构
│   ├── MHA（Master High Availability）
│   ├── MGR（MySQL Group Replication）
│   └── Orchestrator
└── 分库分表 ⭐⭐⭐⭐
    ├── 垂直分库
    ├── 水平分表
    ├── 分片键选择
    ├── 跨分片查询
    └── 中间件（ShardingSphere、MyCat）

性能优化 ⭐⭐⭐⭐⭐
├── SQL优化
│   ├── EXPLAIN分析
│   ├── 索引优化
│   ├── 查询重写
│   └── 避免全表扫描
├── 参数调优
│   ├── innodb_buffer_pool_size
│   ├── max_connections
│   ├── query_cache_size
│   └── innodb_flush_log_at_trx_commit
├── 硬件优化
│   ├── SSD vs HDD
│   └── 内存配置
└── 架构优化
    ├── 缓存层（Redis）
    ├── 读写分离
    └── 分库分表

监控与运维
├── 慢查询日志
├── 性能模式（Performance Schema）
├── 监控工具（Prometheus + mysqld_exporter）
└── 备份恢复
    ├── 逻辑备份（mysqldump）
    └── 物理备份（XtraBackup）
```

##### 4.2.2 PostgreSQL ⭐⭐⭐⭐
```
架构特点
├── 完全ACID
├── 丰富的数据类型
│   ├── JSON/JSONB
│   ├── 数组
│   ├── hstore
│   └── 地理空间类型
└── 扩展性（Extension）

MVCC实现
├── 元组版本
├── 事务ID（XID）
├── 快照隔离
└── VACUUM机制

索引类型
├── B-Tree（默认）
├── Hash
├── GiST（通用搜索树）
├── GIN（倒排索引）
└── BRIN（块范围索引）

高级特性
├── 窗口函数
├── CTE（Common Table Expression）
├── 外部数据包装器（FDW）
└── 逻辑复制
```

##### 4.2.3 ClickHouse ⭐⭐⭐
```
OLAP数据库
├── 列式存储
│   ├── 压缩率高
│   └── 列扫描快
├── 向量化执行
├── 数据分片
└── 实时数据分析

应用场景
├── 日志分析
├── 用户行为分析
└── 实时报表
```

#### 4.3 NoSQL数据库实现

##### 4.3.1 Redis ⭐⭐⭐⭐⭐（Week 10源码阅读）
```
数据结构实现 ⭐⭐⭐⭐⭐
├── SDS（Simple Dynamic String）
│   ├── 预分配空间
│   ├── 二进制安全
│   └── O(1)长度获取
├── 链表（List）
│   ├── 双向链表
│   └── 应用（列表、队列）
├── 字典（Dict）
│   ├── 哈希表
│   ├── 渐进式rehash
│   └── 应用（Hash、数据库、过期键）
├── 跳表（Skip List）⭐⭐⭐⭐⭐
│   ├── 多层索引
│   ├── 随机层数
│   └── 应用（Sorted Set）
├── 整数集合（IntSet）
│   └── 节省内存
└── 压缩列表（ZipList）/ QuickList
    ├── 连续内存存储
    └── 内存优化

持久化机制 ⭐⭐⭐⭐⭐
├── RDB（Redis Database）⭐⭐⭐⭐
│   ├── 快照持久化
│   ├── fork子进程
│   ├── COW（Copy-On-Write）
│   └── 二进制格式
└── AOF（Append Only File）⭐⭐⭐⭐⭐
    ├── 命令追加
    ├── AOF重写
    ├── 三种同步策略
    │   ├── always（每次写）
    │   ├── everysec（每秒）
    │   └── no（交给OS）
    └── RDB vs AOF选择

缓存策略 ⭐⭐⭐⭐⭐
├── 缓存淘汰策略
│   ├── noeviction
│   ├── allkeys-lru ⭐⭐⭐⭐⭐
│   ├── allkeys-lfu
│   ├── volatile-lru
│   ├── volatile-lfu
│   └── volatile-ttl
├── 缓存问题 ⭐⭐⭐⭐⭐
│   ├── 缓存穿透
│   │   ├── 布隆过滤器
│   │   └── 缓存空值
│   ├── 缓存击穿
│   │   ├── 互斥锁
│   │   └── 热点数据永不过期
│   └── 缓存雪崩
│       ├── 过期时间随机化
│       ├── 熔断降级
│       └── 多级缓存
└── 缓存更新策略
    ├── Cache-Aside
    ├── Read-Through
    ├── Write-Through
    └── Write-Behind

高可用方案 ⭐⭐⭐⭐⭐
├── 主从复制 ⭐⭐⭐⭐
│   ├── 全量同步
│   ├── 增量同步
│   ├── 无盘复制
│   └── 主从延迟
├── 哨兵（Sentinel）⭐⭐⭐⭐⭐
│   ├── 监控
│   ├── 故障转移
│   ├── 通知
│   └── 配置中心
└── 集群（Cluster）⭐⭐⭐⭐⭐
    ├── 16384个槽（slot）
    ├── 一致性哈希
    ├── 客户端路由
    └── 集群扩缩容

应用场景 ⭐⭐⭐⭐⭐
├── 分布式锁
│   ├── SETNX实现
│   ├── Redlock算法
│   └── 超时释放
├── 限流器 ⭐⭐⭐⭐⭐
│   ├── 计数器
│   ├── 滑动窗口
│   └── 令牌桶
├── 排行榜
│   └── Sorted Set
├── 消息队列
│   ├── List（LPUSH + BRPOP）
│   └── Stream
├── 布隆过滤器
└── HyperLogLog（基数统计）

性能优化
├── Pipeline（批量操作）
├── Lua脚本（原子性）
├── 慢查询日志
└── bigkey问题
```

##### 4.3.2 MongoDB ⭐⭐⭐
```
文档数据库
├── BSON格式
├── 灵活Schema
├── 聚合框架
└── 分片集群

索引
├── 单字段索引
├── 复合索引
├── 多键索引
└── 地理空间索引

副本集
├── 主从复制
├── 选举机制
└── Oplog
```

#### 4.4 消息队列 ⭐⭐⭐⭐

##### 4.4.1 Kafka ⭐⭐⭐⭐⭐
```
核心概念
├── Topic（主题）
├── Partition（分区）
├── Producer（生产者）
├── Consumer（消费者）
└── Consumer Group

高性能原理 ⭐⭐⭐⭐⭐
├── 顺序写磁盘
│   └── 比随机写内存快
├── Zero Copy ⭐⭐⭐⭐⭐
│   ├── sendfile
│   └── 减少数据拷贝
├── 批量发送
├── 数据压缩
└── 分区并行

可靠性保证 ⭐⭐⭐⭐
├── ISR机制
│   ├── In-Sync Replicas
│   └── 副本同步
├── acks参数
│   ├── acks=0（不等待）
│   ├── acks=1（Leader确认）
│   └── acks=all（所有ISR确认）
└── 幂等性与事务

消费模型
├── 消费者组
├── 分区分配策略
│   ├── Range
│   ├── RoundRobin
│   └── Sticky
└── offset管理

应用场景
├── 日志收集
├── 流处理（Kafka Streams）
├── 事件溯源
└── 消息解耦
```

##### 4.4.2 RabbitMQ ⭐⭐⭐
```
核心概念
├── Exchange（交换机）
│   ├── Direct
│   ├── Topic
│   ├── Fanout
│   └── Headers
├── Queue（队列）
└── Binding（绑定）

可靠性
├── 消息确认（ACK）
├── 持久化
├── 镜像队列
└── 死信队列（DLQ）

应用场景
├── 任务队列
├── RPC调用
└── 延迟队列
```

---

### 5. 后端开发进阶

#### 5.1 API设计 ⭐⭐⭐⭐
```
RESTful API ⭐⭐⭐⭐⭐
├── REST原则
│   ├── 资源（Resource）
│   ├── 表现层（Representation）
│   └── 状态转移（State Transfer）
├── HTTP方法语义
│   ├── GET（查询）
│   ├── POST（创建）
│   ├── PUT（完整更新）
│   ├── PATCH（部分更新）
│   └── DELETE（删除）
├── URL设计
│   ├── 名词复数
│   ├── 层级关系
│   └── 查询参数
├── 状态码 ⭐⭐⭐⭐⭐
│   ├── 2xx（成功）
│   ├── 3xx（重定向）
│   ├── 4xx（客户端错误）
│   └── 5xx（服务器错误）
├── 版本控制
│   ├── URL版本（/v1/users）
│   ├── Header版本
│   └── 内容协商
└── HATEOAS

GraphQL ⭐⭐⭐
├── Schema定义
├── Query/Mutation
├── 解决Over-fetching/Under-fetching
└── Apollo Server/Client

gRPC ⭐⭐⭐⭐
├── Protocol Buffers
│   ├── 二进制序列化
│   ├── 强类型
│   └── 向后兼容
├── HTTP/2
│   └── 多路复用
├── 四种RPC模式
│   ├── Unary RPC
│   ├── Server Streaming
│   ├── Client Streaming
│   └── Bidirectional Streaming
└── 应用场景（微服务通信）
```

#### 5.2 认证与授权 ⭐⭐⭐⭐⭐
```
认证方式
├── Basic Authentication
│   └── Base64编码（不安全）
├── Session/Cookie
│   ├── 服务端session
│   ├── 分布式session问题
│   └── Redis存储session
├── Token认证
│   └── 无状态
└── JWT ⭐⭐⭐⭐⭐
    ├── Header（算法）
    ├── Payload（声明）
    ├── Signature（签名）
    ├── 优点（无状态、跨域）
    ├── 缺点（无法撤销、payload不能太大）
    └── Refresh Token

授权框架
├── OAuth 2.0 ⭐⭐⭐⭐⭐
│   ├── 四种授权模式
│   │   ├── 授权码模式
│   │   ├── 简化模式
│   │   ├── 密码模式
│   │   └── 客户端模式
│   ├── Access Token
│   └── 第三方登录
├── RBAC（基于角色的访问控制）
│   ├── 用户-角色-权限
│   └── Casbin
└── ABAC（基于属性的访问控制）

Web安全 ⭐⭐⭐⭐⭐
├── 加密算法
│   ├── MD5（已不安全，用于校验）
│   ├── SHA-256
│   ├── bcrypt ⭐⭐⭐⭐⭐（密码哈希）
│   └── scrypt
├── 常见攻击与防护
│   ├── XSS（跨站脚本）
│   │   ├── 反射型XSS
│   │   ├── 存储型XSS
│   │   └── 防护（转义、CSP）
│   ├── CSRF（跨站请求伪造）
│   │   └── 防护（Token、SameSite Cookie）
│   ├── SQL注入
│   │   └── 防护（参数化查询、ORM）
│   ├── 点击劫持
│   │   └── X-Frame-Options
│   └── DDOS攻击
│       └── 防护（限流、CDN）
└── 安全头部
    ├── Content-Security-Policy
    ├── X-XSS-Protection
    └── Strict-Transport-Security
```

#### 5.3 测试 ⭐⭐⭐⭐
```
单元测试 ⭐⭐⭐⭐⭐
├── Go testing框架
│   ├── 测试函数（TestXxx）
│   ├── 断言
│   └── 测试覆盖率
├── Mock/Stub
│   ├── gomock
│   ├── testify/mock
│   └── 接口隔离
├── 表驱动测试
└── 测试金字塔

集成测试 ⭐⭐⭐⭐
├── 数据库测试
│   ├── 测试数据库
│   └── 事务回滚
├── API测试
│   └── httptest
└── 容器化测试（Testcontainers）

性能测试 ⭐⭐⭐⭐
├── Benchmark
│   ├── go test -bench
│   └── benchmem
├── 压测工具
│   ├── wrk
│   ├── ab（Apache Bench）
│   └── vegeta
└── 性能分析
    ├── pprof（CPU、内存、goroutine）
    └── trace

端到端测试
├── Selenium
└── Cypress
```

#### 5.4 设计模式 ⭐⭐⭐⭐
```
创建型模式
├── 单例模式
│   ├── 懒汉式（sync.Once）
│   └── 饿汉式
├── 工厂模式
│   ├── 简单工厂
│   ├── 工厂方法
│   └── 抽象工厂
├── 建造者模式
│   └── 链式调用
└── 原型模式

结构型模式
├── 适配器模式
│   └── 接口转换
├── 装饰器模式
│   └── 动态扩展功能
├── 代理模式
│   ├── 静态代理
│   └── 动态代理
├── 外观模式
└── 桥接模式

行为型模式
├── 观察者模式
│   └── 发布-订阅
├── 策略模式
│   └── 算法族
├── 责任链模式
│   └── 中间件
├── 模板方法模式
└── 状态模式

并发模式（Go特有）
├── Pipeline模式
├── Fan-in/Fan-out
├── Worker Pool
└── Context传播
```

#### 5.5 软件工程 ⭐⭐⭐⭐
```
代码质量
├── 代码审查（Code Review）
│   ├── Pull Request流程
│   ├── 审查checklist
│   └── 工具（GitHub PR、GitLab MR）
├── 代码规范
│   ├── Go风格指南
│   ├── 命名规范
│   └── 注释规范
├── 重构
│   ├── 坏味道识别
│   ├── 重构技巧
│   └── 测试保障
└── 静态分析
    ├── golangci-lint
    ├── sonarqube
    └── 代码覆盖率

架构设计
├── 分层架构
│   ├── Controller层
│   ├── Service层
│   └── DAO层
├── 六边形架构（端口适配器）
├── DDD（领域驱动设计）
│   ├── 实体（Entity）
│   ├── 值对象（Value Object）
│   ├── 聚合（Aggregate）
│   └── 领域服务
├── CQRS（命令查询职责分离）
└── Event Sourcing

技术债务管理
├── TODO管理
├── 技术债务记录
└── 定期重构
```

---

### 6. DevOps & 云原生 ⭐⭐⭐⭐⭐

#### 6.1 容器技术 ⭐⭐⭐⭐⭐
```
Docker ⭐⭐⭐⭐⭐
├── 容器原理
│   ├── Namespace（资源隔离）
│   │   ├── PID namespace
│   │   ├── NET namespace
│   │   ├── MNT namespace
│   │   └── UTS namespace
│   ├── Cgroups（资源限制）
│   │   ├── CPU限制
│   │   ├── 内存限制
│   │   └── IO限制
│   └── UnionFS（联合文件系统）
│       └── 镜像分层
├── 镜像
│   ├── Dockerfile
│   │   ├── FROM、RUN、COPY、CMD
│   │   ├── 多阶段构建
│   │   └── 最佳实践（减小镜像体积）
│   ├── 镜像仓库
│   │   ├── Docker Hub
│   │   ├── Harbor（私有仓库）
│   │   └── 阿里云/腾讯云镜像
│   └── 镜像优化
│       ├── 使用Alpine基础镜像
│       ├── 清理缓存
│       └── .dockerignore
├── 容器
│   ├── 生命周期管理
│   ├── 日志查看
│   ├── 进入容器（exec）
│   └── 资源限制
├── 网络
│   ├── Bridge模式
│   ├── Host模式
│   ├── Container模式
│   └── None模式
├── 存储
│   ├── Volume（数据卷）
│   ├── Bind Mount
│   └── tmpfs
└── Docker Compose
    ├── docker-compose.yml
    ├── 服务编排
    └── 多容器应用

Kubernetes ⭐⭐⭐⭐⭐（Week 12源码阅读）
├── 核心概念
│   ├── Pod ⭐⭐⭐⭐⭐
│   │   ├── 最小调度单元
│   │   ├── 容器组
│   │   ├── 共享网络和存储
│   │   └── 生命周期
│   ├── Service ⭐⭐⭐⭐⭐
│   │   ├── ClusterIP
│   │   ├── NodePort
│   │   ├── LoadBalancer
│   │   └── kube-proxy（iptables/IPVS）
│   ├── Deployment ⭐⭐⭐⭐⭐
│   │   ├── 滚动更新
│   │   ├── 回滚
│   │   └── 扩缩容
│   ├── StatefulSet
│   │   ├── 有状态应用
│   │   └── 稳定的网络标识
│   ├── DaemonSet
│   │   └── 每个节点运行一个Pod
│   ├── Job/CronJob
│   │   └── 批处理任务
│   ├── ConfigMap/Secret
│   │   └── 配置管理
│   └── Namespace
│       └── 资源隔离
├── 调度器 ⭐⭐⭐⭐⭐（Week 12重点）
│   ├── 调度流程
│   │   ├── 预选（Predicate）
│   │   │   ├── PodFitsHost
│   │   │   ├── PodFitsResources
│   │   │   └── NodeAffinity
│   │   └── 优选（Priority）
│   │       ├── LeastRequestedPriority
│   │       └── BalancedResourceAllocation
│   ├── 亲和性
│   │   ├── NodeAffinity
│   │   └── PodAffinity/PodAntiAffinity
│   ├── 污点与容忍
│   │   ├── Taint
│   │   └── Toleration
│   └── 资源配额
│       ├── requests
│       └── limits
├── 网络 ⭐⭐⭐⭐
│   ├── CNI（Container Network Interface）
│   │   ├── Flannel
│   │   ├── Calico
│   │   └── Weave
│   ├── Service网络
│   │   └── kube-proxy
│   └── Ingress ⭐⭐⭐⭐
│       ├── Nginx Ingress
│       ├── Traefik
│       └── 路由规则
├── 存储 ⭐⭐⭐⭐
│   ├── Volume
│   ├── PV（Persistent Volume）
│   ├── PVC（Persistent Volume Claim）
│   └── StorageClass
│       └── 动态供应
├── 高级特性
│   ├── HPA（水平自动扩展）
│   │   ├── 基于CPU/内存
│   │   └── 自定义指标
│   ├── VPA（垂直自动扩展）
│   ├── Operator模式
│   │   └── 自定义CRD
│   └── Service Mesh
│       └── Istio
└── 源码阅读重点（Week 12）
    ├── pkg/scheduler/
    │   ├── scheduler.go
    │   └── framework/
    ├── pkg/kubelet/
    └── pkg/controller/
```

#### 6.2 CI/CD ⭐⭐⭐⭐
```
持续集成
├── Jenkins ⭐⭐⭐⭐
│   ├── Pipeline（Groovy DSL）
│   ├── 插件生态
│   └── 分布式构建
├── GitLab CI ⭐⭐⭐⭐
│   ├── .gitlab-ci.yml
│   ├── Runner
│   └── Pipeline as Code
└── GitHub Actions ⭐⭐⭐⭐
    ├── Workflow
    ├── Marketplace
    └── 自托管Runner

持续部署
├── 部署策略
│   ├── 蓝绿部署
│   │   └── 无缝切换
│   ├── 金丝雀发布
│   │   └── 灰度流量
│   ├── 滚动更新
│   │   └── 逐步替换
│   └── A/B测试
├── GitOps
│   ├── ArgoCD
│   ├── Flux
│   └── Git作为唯一真相来源
└── 部署工具
    ├── Helm（K8s包管理）
    └── Kustomize

Pipeline设计
├── 阶段划分
│   ├── 构建（Build）
│   ├── 测试（Test）
│   │   ├── 单元测试
│   │   ├── 集成测试
│   │   └── 性能测试
│   ├── 安全扫描
│   │   ├── 代码扫描
│   │   └── 镜像扫描
│   └── 部署（Deploy）
├── 制品管理
│   ├── Nexus
│   └── Artifactory
└── 通知机制
    ├── 邮件
    ├── 钉钉/企业微信
    └── Slack
```

#### 6.3 监控与可观测性 ⭐⭐⭐⭐⭐
```
监控系统
├── Prometheus ⭐⭐⭐⭐⭐
│   ├── 数据模型
│   │   ├── 时间序列
│   │   └── 标签（Label）
│   ├── 指标类型
│   │   ├── Counter（计数器）
│   │   ├── Gauge（仪表盘）
│   │   ├── Histogram（直方图）
│   │   └── Summary（摘要）
│   ├── PromQL ⭐⭐⭐⭐⭐
│   │   ├── 查询语法
│   │   ├── 聚合操作
│   │   └── 函数
│   ├── 服务发现
│   │   ├── 静态配置
│   │   ├── Kubernetes SD
│   │   └── Consul SD
│   ├── 告警
│   │   ├── Alertmanager
│   │   ├── 告警规则
│   │   └── 路由分组
│   └── Exporter
│       ├── node_exporter（主机监控）
│       ├── mysqld_exporter
│       └── 自定义exporter
├── Grafana ⭐⭐⭐⭐⭐
│   ├── Dashboard设计
│   ├── 数据源（Prometheus、MySQL）
│   ├── 可视化面板
│   └── 变量与模板
└── Datadog ⭐⭐⭐
    └── SaaS监控服务

日志系统
├── ELK Stack
│   ├── Elasticsearch ⭐⭐⭐⭐
│   │   ├── 倒排索引
│   │   ├── 分片与副本
│   │   └── 查询DSL
│   ├── Logstash
│   │   ├── Input/Filter/Output
│   │   └── Grok解析
│   └── Kibana
│       └── 日志搜索与可视化
├── Loki（轻量级日志）⭐⭐⭐⭐
│   ├── 标签索引
│   ├── LogQL
│   └── Grafana集成
└── Fluentd/Fluent Bit
    └── 日志采集

链路追踪 ⭐⭐⭐⭐
├── OpenTelemetry
│   ├── 统一标准
│   └── SDK/API
├── Jaeger ⭐⭐⭐⭐
│   ├── Span
│   ├── Trace
│   └── 分布式追踪
└── Zipkin

应用性能监控（APM）
├── SkyWalking
│   ├── 字节码增强
│   └── 拓扑图
└── New Relic

可观测性三大支柱
├── 指标（Metrics）
├── 日志（Logs）
└── 链路（Traces）
```

#### 6.4 安全与密钥管理 ⭐⭐⭐⭐
```
密钥管理
├── Vault ⭐⭐⭐⭐
│   ├── 动态密钥
│   │   ├── 数据库凭证
│   │   └── AWS凭证
│   ├── 加密存储
│   ├── 访问控制策略
│   └── 密钥轮换
└── Sealed Secrets（K8s）

网络安全
├── 防火墙
│   ├── iptables
│   └── firewalld
├── VPN
│   ├── OpenVPN
│   └── WireGuard
└── SSH
    ├── 密钥认证
    ├── SSH配置加固
    └── 跳板机

容器安全
├── 镜像扫描
│   ├── Trivy
│   └── Clair
├── Runtime Security
│   └── Falco
└── 网络策略（NetworkPolicy）
```

#### 6.5 版本控制 ⭐⭐⭐⭐⭐
```
Git ⭐⭐⭐⭐⭐
├── 基础操作
│   ├── add/commit/push/pull
│   ├── branch/merge
│   ├── checkout/switch
│   └── log/diff
├── 高级操作
│   ├── rebase ⭐⭐⭐⭐
│   │   ├── 变基
│   │   ├── 交互式rebase
│   │   └── 黄金法则
│   ├── cherry-pick
│   │   └── 挑选提交
│   ├── stash
│   │   └── 暂存工作
│   ├── reset/revert
│   │   ├── --soft
│   │   ├── --mixed
│   │   └── --hard
│   └── submodule
│       └── 子模块管理
├── 工作流 ⭐⭐⭐⭐⭐
│   ├── Git Flow
│   │   ├── master/develop分支
│   │   ├── feature分支
│   │   ├── release分支
│   │   └── hotfix分支
│   ├── GitHub Flow
│   │   ├── 简化流程
│   │   └── PR驱动
│   └── GitLab Flow
│       └── 环境分支
├── 钩子（Hooks）
│   ├── pre-commit
│   ├── commit-msg
│   └── pre-push
└── 最佳实践
    ├── Commit Message规范
    │   └── Conventional Commits
    ├── .gitignore
    └── 代码审查流程
```

#### 6.6 云原生生态
```
服务网格（Service Mesh）⭐⭐⭐⭐
├── Istio
│   ├── 流量管理
│   ├── 安全（mTLS）
│   ├── 可观测性
│   └── Sidecar模式
└── Linkerd

Serverless ⭐⭐⭐
├── FaaS（Function as a Service）
│   ├── AWS Lambda
│   ├── Google Cloud Functions
│   └── Knative
└── 应用场景
    ├── 事件驱动
    └── 突发流量

消息总线
├── NATS
└── Pulsar
```

---

### 7. 分布式系统 ⭐⭐⭐⭐⭐

#### 7.1 分布式理论 ⭐⭐⭐⭐⭐
```
CAP理论 ⭐⭐⭐⭐⭐
├── Consistency（一致性）
├── Availability（可用性）
├── Partition Tolerance（分区容错性）
└── CAP不可能三角
    ├── CP系统（etcd、ZooKeeper）
    └── AP系统（Cassandra、DynamoDB）

BASE理论 ⭐⭐⭐⭐
├── Basically Available（基本可用）
├── Soft State（软状态）
└── Eventually Consistent（最终一致性）

一致性模型
├── 强一致性
├── 弱一致性
├── 最终一致性 ⭐⭐⭐⭐⭐
└── 因果一致性

分布式难题
├── 时钟同步
│   ├── NTP
│   └── TrueTime（Google Spanner）
├── 脑裂（Split-Brain）
└── 拜占庭将军问题
```

#### 7.2 分布式算法 ⭐⭐⭐⭐⭐
```
共识算法
├── Raft ⭐⭐⭐⭐⭐（Week 9: etcd源码）
│   ├── Leader选举
│   │   ├── 任期（Term）
│   │   ├── 超时机制
│   │   └── 投票规则
│   ├── 日志复制
│   │   ├── AppendEntries RPC
│   │   ├── 日志一致性
│   │   └── 提交规则
│   ├── 安全性
│   │   ├── Leader Completeness
│   │   └── State Machine Safety
│   └── 应用（etcd、Consul）
├── Paxos ⭐⭐⭐⭐
│   ├── Basic Paxos
│   ├── Multi-Paxos
│   └── 难以理解但经典
└── ZAB（ZooKeeper Atomic Broadcast）
    └── ZooKeeper专用

分布式锁 ⭐⭐⭐⭐⭐
├── Redis分布式锁
│   ├── SETNX + EXPIRE
│   ├── SET NX EX（原子操作）
│   ├── Redlock算法
│   └── 问题（时钟漂移、GC停顿）
├── etcd分布式锁
│   ├── Lease租约
│   ├── 自动续约
│   └── 可靠性更高
└── ZooKeeper锁
    └── 临时顺序节点

分布式ID生成
├── Snowflake算法
│   ├── 时间戳 + 机器ID + 序列号
│   └── 41bit时间 + 10bit机器 + 12bit序列
├── UUID
└── 数据库自增ID
```

#### 7.3 分布式存储 ⭐⭐⭐⭐⭐
```
分布式KV存储
├── etcd ⭐⭐⭐⭐⭐（Week 9源码阅读）
│   ├── Raft共识算法
│   │   ├── raft/raft.go
│   │   └── 选举与日志复制
│   ├── MVCC
│   │   ├── 多版本存储
│   │   └── 版本号管理
│   ├── Watch机制
│   │   └── 事件通知
│   ├── Lease租约
│   │   └── TTL管理
│   └── 应用场景
│       ├── K8s配置存储
│       ├── 服务发现
│       └── 分布式锁
├── RocksDB ⭐⭐⭐⭐⭐（Week 11源码阅读）
│   ├── LSM-Tree ⭐⭐⭐⭐⭐
│   │   ├── MemTable（内存）
│   │   ├── Immutable MemTable
│   │   ├── SSTable（磁盘）
│   │   └── 分层存储
│   ├── Compaction ⭐⭐⭐⭐⭐
│   │   ├── Minor Compaction
│   │   ├── Major Compaction
│   │   ├── Leveled Compaction
│   │   └── Universal Compaction
│   ├── Bloom Filter
│   │   └── 减少磁盘读取
│   ├── WAL（Write-Ahead Log）
│   │   └── 崩溃恢复
│   └── 性能特点
│       ├── 写优化
│       ├── 顺序写
│       └── 读放大/写放大/空间放大
└── LevelDB
    └── RocksDB前身

分布式文件系统
├── HDFS（Hadoop）
│   ├── NameNode + DataNode
│   ├── 块存储
│   └── 副本机制
├── GFS（Google File System）
│   └── 论文经典
├── Ceph
│   ├── CRUSH算法
│   └── 对象/块/文件存储
└── MinIO
    └── S3兼容对象存储

NewSQL ⭐⭐⭐⭐
├── TiDB
│   ├── TiKV（KV存储，基于RocksDB）
│   ├── PD（调度）
│   └── Raft复制
├── CockroachDB
│   └── Google Spanner开源实现
└── Spanner
    └── TrueTime时钟
```

#### 7.4 微服务架构 ⭐⭐⭐⭐⭐
```
服务发现 ⭐⭐⭐⭐⭐
├── etcd
│   └── K8s默认
├── Consul
│   ├── 服务注册
│   ├── 健康检查
│   └── KV存储
└── Nacos
    └── 阿里开源

负载均衡 ⭐⭐⭐⭐⭐
├── Nginx ⭐⭐⭐⭐⭐
│   ├── 反向代理
│   ├── 负载均衡算法
│   │   ├── 轮询（Round Robin）
│   │   ├── 加权轮询
│   │   ├── IP Hash
│   │   └── 最少连接
│   ├── 缓存服务器
│   ├── 限流
│   └── 性能优化
│       ├── Worker进程模型
│       └── epoll事件驱动
├── HAProxy
│   └── 四层/七层负载均衡
├── Envoy
│   └── Service Mesh数据平面
└── 客户端负载均衡
    └── Ribbon

API网关 ⭐⭐⭐⭐
├── Kong
│   ├── 插件生态
│   └── 基于Nginx
├── Traefik
│   └── 云原生网关
└── APISIX
    └── 高性能网关

配置中心 ⭐⭐⭐⭐
├── Apollo
│   ├── 携程开源
│   └── 配置发布/回滚
└── Nacos
    └── 配置 + 服务发现

限流与熔断 ⭐⭐⭐⭐⭐
├── 限流算法
│   ├── 计数器
│   ├── 滑动窗口 ⭐⭐⭐⭐⭐
│   │   └── Redis + Sorted Set
│   ├── 令牌桶 ⭐⭐⭐⭐⭐
│   │   └── 平滑突发流量
│   └── 漏桶
│       └── 固定速率
├── 熔断器（Circuit Breaker）⭐⭐⭐⭐
│   ├── 关闭状态
│   ├── 开启状态
│   ├── 半开状态
│   └── 实现（Hystrix、Sentinel）
└── 降级策略
    ├── 返回默认值
    └── 快速失败

RPC框架 ⭐⭐⭐⭐
├── gRPC（见5.1）
└── Dubbo
    ├── 服务注册
    └── 负载均衡

分布式事务 ⭐⭐⭐⭐
├── 两阶段提交（2PC）
│   ├── Prepare阶段
│   ├── Commit阶段
│   └── 问题（阻塞、单点）
├── 三阶段提交（3PC）
│   └── 增加CanCommit阶段
├── TCC（Try-Confirm-Cancel）
│   ├── 补偿机制
│   └── 业务侵入性强
├── Saga模式
│   ├── 长事务
│   └── 事件驱动
└── 本地消息表
    └── 最终一致性
```

---

### 8. 其他重要技能

#### 8.1 大数据基础 ⭐⭐⭐
```
Hadoop生态
├── HDFS（存储）
├── MapReduce（计算）
└── YARN（资源管理）

Spark
├── RDD
├── Spark SQL
└── 内存计算

数据处理
├── Hive（数据仓库）
├── Flink（流处理）
└── Kafka Streams
```

#### 8.2 机器学习基础 ⭐⭐⭐
```
基础概念（适合CV背景）
├── 监督学习 vs 无监督学习
├── 过拟合与正则化
└── 评估指标

深度学习框架
├── PyTorch
├── TensorFlow
└── ONNX

MLOps
├── 模型训练
├── 模型部署
└── 模型监控
```

#### 8.3 技术写作 ⭐⭐⭐⭐
```
文档编写
├── Markdown
├── API文档（Swagger/OpenAPI）
└── 系统设计文档

技术博客
├── 文章结构
├── 代码示例
└── 图表制作（Mermaid、PlantUML）

开源贡献
├── README编写
├── Contributing Guide
└── Issue/PR规范
```

---

## 📊 学习优先级说明

### ⭐⭐⭐⭐⭐ 必须精通（对应LEARNING_ROADMAP Week 1-12）
```
数据结构：LRU Cache、红黑树、跳表、B+树、布隆过滤器、LSM-Tree
算法：动态规划、图算法、字符串算法
Go底层：GPM模型、Channel、GC、Map/Slice内部实现
操作系统：I/O模型（epoll）、进程调度、内存管理
网络：TCP/IP、HTTP、HTTPS、epoll
数据库：MySQL索引、事务、Redis数据结构与持久化
分布式：Raft算法、etcd、RocksDB、Kubernetes调度器
```

### ⭐⭐⭐⭐ 需要掌握（工作必备）
```
Docker、K8s基础
CI/CD流程
Prometheus + Grafana监控
RESTful API、gRPC设计
JWT认证、OAuth2
单元测试、性能测试
Git工作流
Nginx负载均衡
微服务基础架构
```

### ⭐⭐⭐ 需要了解（扩展知识）
```
GraphQL
MongoDB、Cassandra
Service Mesh（Istio）
Serverless
大数据基础（Hadoop、Spark）
机器学习基础
```

---

## 🔗 与LEARNING_ROADMAP的对应关系

| 周次 | 学习内容 | 对应本文档章节 |
|-----|---------|--------------|
| Week 1-2 | 数据结构（LRU、跳表、Trie、红黑树、B+树） | 2.1 基础数据结构 |
| Week 3-4 | 算法专题（DP、图算法） | 2.2 算法专题 |
| Week 5-6 | Go底层（GPM、Channel、GC） | 1.1 Go语言底层原理 |
| Week 7 | 操作系统（进程、内存、I/O） | 3.1 操作系统 |
| Week 8 | 网络（TCP/IP、epoll） | 3.2 计算机网络 |
| Week 9 | etcd源码（Raft） | 7.3 分布式存储 - etcd |
| Week 10 | Redis源码（数据结构、持久化） | 4.3.1 Redis |
| Week 11 | RocksDB源码（LSM-Tree） | 7.3 分布式存储 - RocksDB |
| Week 12 | K8s源码（调度器） | 6.1 容器技术 - Kubernetes |

---

## 💼 面试重点

### 后端开发岗位（BAT/字节/美团）
**必考知识点**：
1. 数据结构与算法（LeetCode Medium/Hard）
2. MySQL索引原理、事务隔离级别、MVCC
3. Redis数据结构、持久化、缓存策略
4. Go并发编程（Goroutine、Channel）
5. TCP/IP三次握手、四次挥手、拥塞控制
6. 分布式系统（CAP、一致性算法）
7. 微服务架构（服务发现、限流熔断）

### DevOps/SRE岗位
**必考知识点**：
1. Linux系统管理（性能分析、故障排查）
2. Docker与Kubernetes
3. CI/CD流程设计
4. Prometheus + Grafana监控
5. 脚本编程（Bash、Python）
6. 网络协议（TCP/IP、HTTP）
7. 高可用架构设计

### 系统架构师/技术专家
**必考知识点**：
1. 分布式系统设计（CAP、BASE、一致性）
2. 高并发架构（负载均衡、缓存、限流）
3. 微服务架构（服务拆分、治理）
4. 性能优化（CPU、内存、IO、网络）
5. 源码阅读能力（etcd、Redis、K8s）
6. 技术选型能力
7. 团队管理与技术规划

---

**这个技术体系会随着你的学习不断完善！建议打印或收藏，定期review！**
