---
title: gdb指令级别单步调试
date: '2021-05-18T21:40:39+08:00'
url: /2021/05/18/gdb-step-debug/
draft: false
categories:
- gdb
tags:
- gdb
---

### gdb单步调试

#### 原来服务的的启动方式

```plain
/data/home/2333/xx/matchsvr/bin/matchsvr --noloadconf --conf-file=../conf/matchsvr_conf.xml 
    --log-file=../log --tlogconf=../conf/tlogconf.xml --id=28.233.0.1 
    --business-id=0  -D restart
```

#### gdb拉起，设置命令行参数

```plain
$ gdb matchsvr
set args --noloadconf --conf-file=../conf/matchsvr_conf.xml 
    --log-file=../log --tlogconf=../conf/tlogconf.xml --id=28.233.0.1 --business-id=0  -D restart
```

#### 处理守护进程

对于守护进程，实际服务的进程是fork两次后的子进程，我们在父进程打断点是起不到作用的

#### set follow-fork-mode [parent|child]

parent: fork之后继续调试父进程，子进程不受影响。

child: fork之后调试子进程，父进程不受影响。

如果需要调试子进程，在启动gdb后：

(gdb) set follow-fork-mode child

```plain
(gdb) set detach-on-fork on
(gdb) set follow-fork-mode child
```

#### 打断点的几种方式

```plain
(gdb) b tbuspphook::hook_tbuspp::Peek   
(gdb) b ua::TBusppChannel::OnRecv
(gdb) b hook_tbuspp.h:108 
(gdb) r 
(gdb) c
(gdb) p pCall
$1 = (int (tbuspphook::hook_tbuspp::*)(tbuspphook::hook_tbuspp * const, int *, const char **...
```

#### 设置单步调试

```plain
(gdb) set disassemble-next-line on
```

按照指令级别 单步调试  

```plain
(gdb) si 
0x00007ffff7556aa2      108     in server/hook/hook_tbuspp/hook_tbuspp.h
   0x00007ffff7556a9e <hook_tbuspp::Peek(int*)+64>:     48 8b 45 f0     mov    -0x10(%rbp),%rax
=> 0x00007ffff7556aa2 <hook_tbuspp::Peek(int*)+68>:     83 e0 01        and    $0x1,%eax
   0x00007ffff7556aa5 <hook_tbuspp::Peek(int*)+71>:     48 85 c0        test   %rax,%rax
   0x00007ffff7556aa8 <hook_tbuspp::Peek(int*)+74>:     75 06   jne    0x7ffff7556ab0 <hook_tbuspp::Peek(int*)+82>
   0x00007ffff7556aaa <hook_tbuspp::Peek(int*)+76>:     48 8b 45 f0     mov    -0x10(%rbp),%rax
   0x00007ffff7556aae <hook_tbuspp::Peek(int*)+80>:     eb 1f   jmp    0x7ffff7556acf <hook_tbuspp::Peek(int*)+113>
   0x00007ffff7556ab0 <hook_tbuspp::Peek(int*)+82>:     48 8b 45 f8     mov    -0x8(%rbp),%rax
   0x00007ffff7556ab4 <hook_tbuspp::Peek(int*)+86>:     48 89 c2        mov    %rax,%rdx
```

##### 看崩溃在哪一条指令

![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/work/gdb-step-debug.png)
