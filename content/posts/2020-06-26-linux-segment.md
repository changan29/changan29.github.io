---
title: Linux中的段
date: '2020-06-26T15:40:39+08:00'
url: /2020/06/26/linux-segment/
draft: false
categories:
- kernel
tags:
- kernel
---

### Linux中的段

Intel 微处理器的段机制是从8086 开始提出的， 那时引入的段机制解决了从CPU 内部  
16 位地址到20 位实地址的转换。为了保持这种兼容性，386 仍然使用段机制，但比以前复杂。  
因此，Linux 内核的设计并没有全部采用Intel 所提供的段方案，仅仅有限度地使用  
了一下分段机制。这不仅简化了Linux 内核的设计，而且为把Linux 移植到其他平台创造了  
条件，因为很多RISC 处理器并不支持段机制。但是，对段机制相关知识的了解是进入Linux  
内核的必经之路。

从2.2 版开始，Linux 让所有的进程（或叫任务）都使用相同的逻辑地址空间，因此就  
没有必要使用局部描述符表LDT。但内核中也用到LDT，那只是在VM86 模式中运行Wine 时，  
即在Linux 上模拟运行Windows 软件或DOS 软件的程序时才使用。

#### linux的GDT

Linux 在启动的过程中设置了段寄存器的值和全局描述符表GDT 的内容，段的定义在include/asm-i386/segment.h 中：

```plain
#define __KERNEL_CS     0x10    /* 内核代码段, index=2,TI=0,RPL=0 */
 
#define __KERNEL_DS     0x18    /* 内核数据段, index=3,TI=0,RPL=0 */
 
#define __USER_CS       0x23    /* 用户代码段, index=4,TI=0,RPL=3 */
 
#define __USER_DS       0x2B    /* 用户数据段, index=5,TI=0,RPL=3 */
```

从定义看出，没有定义堆栈段，实际上，Linux 内核不区分数据段和堆栈段，这也体现  
了Linux 内核尽量减少段的使用。因为没有使用LDT，因此，TI=0，并把这4 个段都放在GDT  
中, index 就是某个段在GDT 表中的下标。内核代码段和数据段具有最高特权，因此其RPL  
为0，而用户代码段和数据段具有最低特权，因此其RPL 为3。可以看出，Linux 内核再次简  
化了特权级的使用，使用了两个特权级而不是4 个。

全局描述符表的定义在arch/i386/kernel/head.S 中：

```plain
ENTRY(gdt_table)
    .quad 0x0000000000000000    /* NULL descriptor */
    .quad 0x0000000000000000    /* not used */
    .quad 0x00cf9a000000ffff    /* 0x10 kernel 4GB code at 0x00000000 */
    .quad 0x00cf92000000ffff    /* 0x18 kernel 4GB data at 0x00000000 */
    .quad 0x00cffa000000ffff    /* 0x23 user   4GB code at 0x00000000 */
    .quad 0x00cff2000000ffff    /* 0x2b user   4GB data at 0x00000000 */
    .quad 0x0000000000000000    /* not used */
    .quad 0x0000000000000000    /* not used */
    /*
     * The APM segments have byte granularity and their bases
     * and limits are set at run time.
     */
    .quad 0x0040920000000000    /* 0x40 APM set up for bad BIOS's */
    .quad 0x00409a0000000000    /* 0x48 APM CS    code */
    .quad 0x00009a0000000000    /* 0x50 APM CS 16 code (16 bit) */
    .quad 0x0040920000000000    /* 0x58 APM DS    data */
    .fill NR_CPUS*4,8,0     /* space for TSS's and LDT's */
```

从代码可以看出，GDT 放在数组变量gdt\_table 中。按Intel 规定，GDT 中的第一项为  
空，这是为了防止加电后段寄存器未经初始化就进入保护模式而使用GDT 的。第二项也没用。  
从下标2～5 共4 项对应于前面的4 种段描述符值。对照图2.10，从描述符的数值可以得出：

- 段的基地址全部为0x00000000；
- 段的上限全部为0xffff；
- 段的粒度G 为1，即段长单位为4KB；
- 段的D 位为1，即对这4 个段的访问都为32 位指令；
- 段的P 位为1，即4 个段都在内存。

由此可以得出，每个段的逻辑地址空间范围为0～4GB。读者可能对此不太理解，但这种设置既简单又巧妙。因为每个段的基地址为0，因此，逻辑地  
址到线性地址映射保持不变，也就是说，偏移量就是线性地址，我们以后所提到的逻辑地址  
（或虚拟地址）和线性地址指的也就是同一地址。看来，Linux 巧妙地把段机制给绕过去了，  
而完全利用了分页机制。

从逻辑上说，Linux 巧妙地绕过了逻辑地址到线性地址的映射，但实质上还得应付Intel  
所提供的段机制。只不过，Linux 把段机制变得相当简单，它只把段分为两种：用户态（RPL  
＝3）的段和内核态（RPL=0）的段。另外，用户段和内核段的区别也仅仅在其  
RPL 不同，因此内核根本无需访问描述符投影寄存器，当然也无需访问GDT，而仅从段寄存器  
的最低两位就可以获取RPL 的信息。Linux 这样设计所带来的好处是显而易见的，Intel 的分  
段部件对Linux 性能造成的影响可以忽略不计。

按Intel 的规定，每个进程有一个任务状态段（TSS）和局部描述符表LDT，但Linux 也  
没有完全遵循Intel 的设计思路。如前所述，Linux 的进程没有使用LDT，而对TSS 的使用也  
非常有限，每个CPU 仅使用一个TSS。

通过上面的介绍可以看出，Intel 的设计可谓周全细致，但Linux 的设计者并没有完全  
陷入这种沼泽，而是选择了简洁而有效的途径，以完成所需功能并达到较好的性能为目标。

### 段寄存器与段选择子

段寄存器CS 、DS 、ES 、FS 、GS 、SS ，在实模式下时，段中存储的是段基地址，即内存段的起始地址。  
而在保护模式下时，由于段基址已经存入了段描述符中，所以段寄存器中再存放段基址是没有意义的，在段寄  
存器中存入的是一个叫作选择子的东西– selector。用此索引值在段描述符表中索引相应的段描述符，这样，便在段描述符  
中得到了内存段的起始地址和段界限值等相关信息。

由于段寄存器是16 位，所以选择子也是16 位，在其低2 位即第O～ 1 位，  
用来存储RPL，即请求特权级，可以表示0、1 、2 、3 四种特权级。在选  
择子的第2 位是TI位，即Table Indicator，用来指示选择子是在GDT 中，还是LDT 中索引描述符。TI  
为0 表示在GDT 中索引描述符， TI 为1 表示在LDT 中索引描述符。选择子的高13 位，即第3～ 15 位是  
描述符的索引值，用此值在GDT 中索引描述符。

![image](https://202002-1251969284.cos.ap-shanghai.myqcloud.com/linux_seg_desc_table.png)

![image](https://202002-1251969284.cos.ap-shanghai.myqcloud.com/linux_seg_desc_selector.png)

### 学而不思则罔

计算机系统从最初发展到现在，硬件、软件都在发展着，而且二者的发展有些地方是目的相同的。例如多任务这个方向，内存管理是实现多任务的必要条件，硬件在这方面的发展是MMU去支持内存管理，而软件的发展就是内核去配合MMU，从而才能实现内存管理的虚拟空间、分页。

分析新处理器上的内核代码时，可以分析处理器新功能带来的更新。例如arm处理器，由于arm处理器架构比较灵活，一般说arm寄存器并没有内存管理相关的寄存器，只说7种模式下的16个寄存器，其实arm的内存管理需要协处理器CP15去支持。另外arm处理器的内存管理也没有段机制，直接就是页机制。

### 参考

- [linux-segment](http://mrhopehub.github.io/2014/11/22/linux-segment.html)
