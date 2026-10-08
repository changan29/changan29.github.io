---
title: perf火焰图 wiki
date: '2020-11-17T15:40:39+08:00'
url: /2020/11/17/perf-flame-graph/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 火焰图 wiki {#火焰图-wiki}

火焰图的每个框代表一个栈里的一个函数，燃烧在火苗尖部的就是CPU正在执行的函数

- 纵轴代表调用栈的深度（栈桢数），火焰顶端显示正在运行的函数，这之下都是调用者。在下面的函数是上面函数的父函数
- 横轴代表消耗的时间，因为调用栈在横向会按照字母排序，并且同样的调用栈会做合并，所以一个格子的宽度越大越说明其可能是瓶颈，比如上图中的3个函数，消耗CPU非常多，要是能做优化，系统性能自然就上来了。

![image](https://twork-1251969284.cos.ap-guangzhou.myqcloud.com/perf/perf.16.41.svg)

### 使用perf 生成火焰图 {#使用perf-生成火焰图}

```plain
# PID 指进程id， -F 指采样频率，高负载时最好指定，不然生成的文件很大 
# -g 告诉 perf record 额外记录函数的调用关系 
# -- sleep 表示 采样时间 这里采样10s
root@localhost ~]#perf record -F 99 -g -p PID -- sleep 10 
# perf sript 对 perf record 生成的perf.data进行解析
# stackcollapse-perf.pl 对 perf script 结果中的符号进行折叠
# flamegraph.pl 生成svg格式火焰图
root@localhost ~]#perf script | ./stackcollapse-perf.pl | ./flamegraph.pl > perf.svg
```

#### 注意点 {#注意点}

- perf record 高负载采样时，最好指定采样频率或控制采样时间，特别是现网环境下，尽量将生成的perf.data控制在200M内，不然perf script 会生成大文件，后面生成火焰图就会跑满单核CPU很长一段时间，若现网调试，请多加小心。另外，还遇到系统高负载时perf top 分析把系统搞挂的情况（小概率），虽然perf record我还没遇到过，但也请多加注意。
- 有些模块的火焰图中会出现一些函数框，这其实不是火焰图的锅，是perf采样的时候就有了，它找不到相应的符号表，无法解析调用栈。假如框很少的话可以先不管；要是很多的话，可以在编译时加上编译选项-fno-omit-frame-pointer，重新编译程序，包括so文件，再重新采样生成火焰图。
- 假如火焰图的一些函数框里什么也没有，应该也是找不到符号信息，无法解析调用栈，所以在编译时记得加上-g参数，-O2也会优化掉调用栈信息，可以去掉，编译完成后也不要用strip命令清符号表信息。
