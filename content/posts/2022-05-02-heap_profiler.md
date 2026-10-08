---
title: C++程序内存泄露工具
date: '2022-05-02T21:40:39+08:00'
url: /2022/05/02/heap_profiler/
draft: false
categories:
- C++
tags:
- C++
---

在使用 gperftools 的 heap profiler 分析 C++ 程序的内存使用情况时，您可以生成多个内存分配描述文件（heap profiles），这些文件记录了程序在特定时间点的内存分配状态。通过对比这些描述文件（即进行 “diff” 操作），可以发现在两个时间点之间的内存分配差异。

具体来说，`pprof` 工具可以用来查看和分析这些描述文件。通过 `pprof` 的 `--text`、`--list` 和其他命令行选项，可以输出具体的内存分配详情，包括具体的函数、调用栈以及分配的内存量等信息。

## 详细步骤

### 1. 启动内存分析

首先，在程序中启动 heap profiler。这通常通过设置环境变量 `HEAPPROFILE` 来实现，这个环境变量指定了描述文件的输出路径和前缀。

```
HEAPPROFILE=/tmp/my_program_heap
export HEAPPROFILE
./my_program  # 运行你的程序
```

### 2. 生成内存描述文件

程序运行过程中，heap profiler 将定期生成内存描述文件（例如 `my_program_heap.0001.heap`、`my_program_heap.0002.heap` 等），每个文件都代表一个内存分配的快照。

### 3. 分析和对比描述文件

使用 `pprof` 工具来分析和对比这些文件。例如，要查看两个时间点之间的差异，可以使用以下命令：

```
pprof --text --base=my_program_heap.0001.heap my_program my_program_heap.0002.heap
```

### 4. 理解输出

`pprof` 的输出会列出内存分配的详细信息，包括增加或减少的内存分配。如果使用 `--list=<function_name>` 选项，可以看到某个特定函数内的详细内存分配情况。

## 原理

### 原理和操作的底层函数

拦截内存分配函数：  
Heap profiler 通过覆盖标准的内存分配和释放函数来工作，如 malloc, free, new, delete 等。当程序调用这些函数时，它实际上是调用了 profiler 提供的替代版本。

### 记录内存活动：

每次内存分配或释放时，heap profiler 记录关键信息，包括分配的大小、时间戳和调用栈。这些数据被用来跟踪内存使用的历史和发展趋势。

### 生成堆描述文件：

分配的数据被定期写入到堆描述文件（heap profiles），这些文件反映了程序在特定时间点的内存使用情况。这些描述文件可以用于进一步的分析，如比较两个时间点的内存使用差异。

### 调用栈捕获：

为了提供内存分配的上下文信息，heap profiler 会捕获当前的调用栈。这通常通过使用 backtrace 或类似机制来实现，这允许 profiler 记录哪个函数请求了内存。

### 如何实现

在底层，gperftools 的 heap profiler 修改了程序的动态链接（通过 LD\_PRELOAD 或其他机制），使得所有内存分配调用都被重定向到 profiler 自己的实现。这个实现中包括了额外的逻辑来记录和分析内存使用数据。这些记录的数据包括：

- 内存分配的大小和地址
- 分配和释放的时间戳
- 发生分配和释放时的调用栈
