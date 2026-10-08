---
title: 写一个操作系统_09 C语言和汇编语言
date: '2019-04-02T21:08:45+08:00'
url: /2019/04/02/make_a_linux_9/
draft: false
categories:
- kernel
- OS
tags:
- OS
- kernel
---

#### 函数调用约定 {#函数调用约定}

编译器一般使用堆栈实现函数调用，每个进程都有自己的栈，用栈来传递参数，会带来以下问题：

- 参数存储在栈中，那么谁来负责回收参数所占的栈空间，是调用者？还是被调用者
- 当参数很多的时候后，主调函数将参数以什么样的顺序传递

高级语言有调用的约定:  
![image](http://201904-1251969284.cossh.myqcloud.com/func_call_types.jpg)

举例:

```plain
int sub(int a , int b){ // 被调
    return a-b;
}
int s = sub(3,2);       // 调用

调用:
push 2
push 3
call sub

被调:
push ebp
mov ebp,esp
mov eax,[ebp+8]
sub eax,[ebp+12]
pop ebp         // 恢复 ebp
```

#### C & ASM 混合 {#C-amp-ASM-混合}

C和汇编混合编程有两种方式:

- 用汇编和C单独写单独编译，然后链接在一起
- C代码嵌入汇编

C代码嵌入汇编:  

```plain
char* str = "C & ASM Code\n";
int count = 0;
int main()
{
    asm("pusha;\
            movl $4,%eax;\
            movl $1,%ebx;\
            movl str,%ecx;\
            movl $13,%edx;\
            int $0x80;\
            movl %eax,count;\
            popa;\
            ");
}
// gcc -m32 casm.c
```

#### 内联汇编及其扩展 {#内联汇编及其扩展}

上面的例子是内联汇编的一个例子，基本的内敛汇编功能比较薄弱，在此基础上，GCC添加了扩展功能，在C语言里面用扩展汇编真的不舒服，相当于多了一种语言，需要依据其规则。
