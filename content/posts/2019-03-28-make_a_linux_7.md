---
title: 写一个操作系统_07 设计自己的页表
date: '2019-03-28T21:08:45+08:00'
url: /2019/03/28/make_a_linux_7/
draft: false
categories:
- kernel
- OS
tags:
- OS
- kernel
---

### 设计自己的页表 {#设计自己的页表}

只针对32位的操作系统，设计一个二级页表，目的是构建一个简易的能跑起来的操作系统。对于4G的地址空间，每个页大小是4K,模仿Linux早期的做法，32位地址的前10位为页目录项，中间10位为页表，后面10位为偏移量。

```plain
|--目录项----|--页表-----|---地址偏移-|
|---10bit----|---10bit---|----12bit---|
```

#### 总体布局 {#总体布局}

我们把我们的内核放在物理地址1M以内，页表放在物理地址1M的地方，即 0x100000,目录项占 1K \* 4B = 4K,那么页表在 0x101000的地方。

![image](http://201904-1251969284.cossh.myqcloud.com/page_phy_addr.png)

备注:

- 我们模仿Linux，0-3G是用户空间，3G-4G是内核空间，又因为咱们写死了内核加载到物理地址1M以内，所以将0xc0000000 映射到物理地址的0-4M上
- 第一个页表也映射到0-4M这个空间，我们的内核用了这个地方，方便操作
- 最后一个目录项指向页目录表自己的地址，为了后面进程页表操作方便

```plain
setup_page:
;先把页目录占用的空间逐字节清0
   mov ecx, 4096
   mov esi, 0
.clear_page_dir:
   mov byte [PAGE_DIR_TABLE_POS + esi], 0
   inc esi
   loop .clear_page_dir

;开始创建页目录项(PDE)
.create_pde:				     
; 创建Page Directory Entry
   mov eax, PAGE_DIR_TABLE_POS
   add eax, 0x1000 			     ; 此时eax为第一个页表的位置及属性
   mov ebx, eax				     ; 此处为ebx赋值，是为.create_pte做准备，ebx为基址。

;   下面将页目录项0和0xc00都存为第一个页表的地址，
;   一个页表可表示4MB内存,这样0xc03fffff以下的地址
;   和0x003fffff以下的地址都指向相同的页表，
;   这是为将地址映射为内核地址做准备
   or eax, PG_US_U | PG_RW_W | PG_P	     ; 
   ;页目录项的属性RW和P位为1,US为1,表示用户属性,所有特权级别都可以访问.
   mov [PAGE_DIR_TABLE_POS + 0x0], eax       ; 
   ;第1个目录项,在页目录表中的第1个目录项写入第一个页表的位置(0x101000)及属性(7)
   mov [PAGE_DIR_TABLE_POS + 0xc00], eax     ; 
   ;一个页表项占用4字节,0xc00表示第768个页表占用的目录项,0xc00以上的目录项用于内核空间,
					     ;也就是页表的0xc0000000~0xffffffff共计1G属于内核,
					     ;0x0~0xbfffffff共计3G属于用户进程.
   sub eax, 0x1000
   mov [PAGE_DIR_TABLE_POS + 4092], eax	     ; 使最后一个目录项指向页目录表自己的地址
```
