---
title: 写一个操作系统_08 内核与C语言
date: '2019-03-29T21:08:45+08:00'
url: /2019/03/29/make_a_linux_8/
draft: false
categories:
- kernel
- OS
tags:
- OS
- kernel
---

### 内核与C语言

#### 写内核能用C标准库吗？

我们开发用户应用程序的时候，有标准库可以用，最典型的就是GUN C库，标准库一般是系统调用的封装，表面上是通过标准库访问系统资源，实际上是通过系统调用实现的。Linux的系统调用一般是先往eax寄存器写入系统调用号，然后通过0x80中断来实现。中断向量号为0x80称为系统中断门，更多的中断参考中断描述符表。

所以自己写内核的话可以用C语言，毕竟可以编译成汇编，跟直接写汇编差不多，但是不能用C标准库，理由如上。

#### 编译32位程序

默认的GCC编译选项会编译出与系统一致的输出，例如64位直接使用gcc或g++编译出的为64位程序或库，而32位系统编译的是对应32位的。  
可以通过在编译时添加选项-m32或-m64来指定编译生成的相应版本。

问题的关键时多数时候会提示缺少库，这里以64位下编译32位程序为例，给出ubuntu和centos下相关依赖包。

```plain
sudo apt-get install build-essential module-assistant
sudo apt-get install gcc-multilib g++-multilib

yum install glibc-devel.i686 libstdc++-devel.i686
```

#### C运行时库

C运行时库与标准库不一样，可以理解为与操作系统无关的部分的汇编封装。  

```plain
// 32b.c 
 int main()
 {
     while(1);
     return 0;
 }
```

```plain
# gcc -c -o m32b.o 32b.c 
# ld m32b.o -Ttext 0xc0001500  -o m32.bin  // 默认链接 c-runtime-lib crt.o _start 
ld: warning: cannot find entry symbol _start; defaulting to 00000000c0001500
# ld m32b.o -Ttext 0xc0001500 -e main -o m32.bin   # 指定开始的地址
```

#### ELF文件

ELF是Linux上的可执行文件，其中Entry point address相当于操作系统与文件的约定，操作系统加载这个可执行文件的时候，去这个地方作为入口点。  

```plain
[root@VM_80_251_centos ~]# readelf -e m32.bin 
ELF Header:
  Magic:   7f 45 4c 46 02 01 01 00 00 00 00 00 00 00 00 00 
  Class:                             ELF64
  Data:                              2's complement, little endian
  Version:                           1 (current)
  OS/ABI:                            UNIX - System V
  ABI Version:                       0
  Type:                              EXEC (Executable file)
  Machine:                           Advanced Micro Devices X86-64
  Version:                           0x1
  Entry point address:               0xc0001500                 ## 起始地址 <--
  Start of program headers:          64 (bytes into file)
  Start of section headers:          5792 (bytes into file)
  Flags:                             0x0
  Size of this header:               64 (bytes)
  Size of program headers:           56 (bytes)
  Number of program headers:         2
  Size of section headers:           64 (bytes)
  Number of section headers:         7
  Section header string table index: 6
```

#### 如何在没有操作系统的时候调试

我们前面写好了loader, 把他加载到磁盘中，然后根据启动的流程，启动mbr,调用loader等过程，我们写完了C语言程序，设定入口后，需要将入口的虚拟地址映射到一块可以访问的物理内存，将编译完的内核也加载到磁盘，然后从磁盘加载到内存，跳转到入口执行。
