---
title: 写一个操作系统_05 调试MBR
date: '2019-03-23T21:08:45+08:00'
url: /2019/03/23/make_a_linux_5/
draft: false
categories:
- kernel
- OS
tags:
- OS
- kernel
---

### 熟悉MBR引导

MBR的512字节里面也是代码，他被BIOS加载执行，自身执行的时候加载操作系统的loader，我们在没有操作系统loader的情况下先写一段测试程序，体验一下。

实现的功能如下:

BIOS 加载 MBR中的代码，执行，用BIOS 中断 INT 0x10显示字符串

#### MBR的代码

```plain
;主引导程序 
; 起始编译地址是 0x7c00
SECTION MBR vstart=0x7c00         
   mov ax,cs      
   mov ds,ax
   mov es,ax
   mov ss,ax
   mov fs,ax
   mov sp,0x7c00

; 使用BIOS 显示服务（Video Service）--INT 10H 
;;;;;;;;;    下面这三行代码是获取光标位置    ;;;;;;;;;
;.get_cursor获取当前光标位置,在光标位置处打印字符.
   mov ah, 3		; 输入: 3号子功能是获取光标位置,需要存入ah寄存器
   mov bh, 0		; bh寄存器存储的是待获取光标的页号

   int 0x10		; 输出: ch=光标开始行,cl=光标结束行
			; dh=光标所在行号,dl=光标所在列号

;;;;;;;;;    获取光标位置结束    ;;;;;;;;;;;;;;;;

;;;;;;;;;     打印字符串    ;;;;;;;;;;;
   ;还是用10h中断,不过这次是调用13号子功能打印字符串
   mov ax, message 
   mov bp, ax		; es:bp 为串首地址, es此时同cs一致，
			; 开头时已经为sreg初始化

   ; 光标位置要用到dx寄存器中内容,cx中的光标位置可忽略
   mov cx, 8		; cx 为串长度,不包括结束符0的字符个数
   mov ax, 0x1301	; 子功能号13是显示字符及属性,要存入ah寄存器,
			; al设置写字符方式 ah=01: 显示字符串,光标跟随移动
   mov bx, 0x2		; bh存储要显示的页号,此处是第0页,
			; bl中是字符属性, 改颜色，要不看不见
   int 0x10		; 执行BIOS 0x10 号中断
;;;;;;;;;      打字字符串结束	 ;;;;;;;;;;;;;;;

   jmp $		; ; 死循环

   message db "TEST MBR"	; 字符串定义
   ; 保证 MBR  512B 
   times 510-($-$$) db 0
   db 0x55,0xaa
```

#### 编译成二进制

```plain
nasm -o mbr.bin mbr.S
```

#### 将二进制文件存储到MBR扇区

```plain
dd if=./mbr.bin  of=./hd30M.img bs=512 count=1 conv=notrunc
```

#### 调试 MBR 代码

在 bochs的配置文件中配好 MBR扇区对应的磁盘，启动bochs

```plain
ata0-master: type=disk, mode=flat, path="/root/data/bochs/bin/hd30M.img"

./bochs -f bochsrc-sample.txt
```

#### 在虚拟机中调试

![image](http://201904-1251969284.cossh.myqcloud.com/MBR_DEBUG.png)
