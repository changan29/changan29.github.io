---
title: 从hook的并发症理解x64指令格式
date: '2020-11-20T15:40:39+08:00'
url: /2020/11/20/hook-complications-and-x64-instrcution-format/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### hook的并发症 {#hook的并发症}

可以理解成一个有意思的问题，假如地址 addr1 上有一个函数func1，长度为len, 将这个函数 整体换一个位置，挪到 addr2, 移动之后的函数成为func2

```plain
memcpy(addr2 ,addr1,len );
```

原来调用 func1 语法是:

```plain
func1(arg1 , arg2 , arg3);
```

问:  
现在用如下的方式调用 func2 是否会发生异常？

```plain
func2(arg1 , arg2 , arg3);
```

假如把 func1 的开头 修改为特定的 shellcode，改成 跳转+ 目标跳转地址的格式，就是传统的 inline hook， 但是这种hook将 原来函数的指令 挪了位置，再次调用有的时候会崩溃， 有的时候称这种情况为 hook的 并发症。

### 指令格式 {#指令格式}

指令包括可选的指令前缀 (in any order)，主要操作码字节 (up to three bytes)，由ModR / M字节以及有时由SIB（Scale-Index-Base）组成的寻址形式说明符 (if required) ，位移字段 (if required)和立即数据字段 (if required)。

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/op-summery.png)

#### Instruction Prefixes {#Instruction-Prefixes}

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/Instruction-format-Prefixes.png)

### opcode {#opcode}

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/Instruction-format-opcode.png)

### TODO {#TODO}

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/harbin-snow-20201120.jpg)
