---
title: hook的几种方式及原理学习
date: '2020-09-27T01:17:39+08:00'
url: /2020/09/27/hook_and_internal/
draft: false
categories:
- hook
- os
tags:
- os
- hook
---

### 概述 {#概述}

对于大型的工程项目，依赖许多人的配合，包含大量不同的代码库与服务，有的我们能够访问程序的源代码，有的可以访问程序的可重定位文件，有的可以访问到可执行文件及其环境，假如我们想在在不同的层面改变或者添加一些逻辑，操作系统、编译器以及程序语言、代码库等都提供了 一些机制使得 开发者可以 方便的 增加或替换代码逻辑，对于逻辑调试、测试、性能分析、版本兼容等都有比较好的效果。

### 编译器支持 {#编译器支持}

#### Function Attribute {#Function-Attribute}

GNU C 使用attribute 可以设置函数属性（Function Attribute ）、变量属性（Variable Attribute ）和类型属性（Type Attribute ）。 **attribute**前后都有两个下划线，并且后面会紧跟一对原括弧，括弧里面是相应的**attribute**参数。

**attribute**语法格式为：**attribute** ( ( attribute-list ) )

比如常用的constructor属性，则会使函数在main（）函数执行之前被自动的执行。

```plain
#include <stdio.h>
#include <stdlib.h>
static int * g_count = NULL;
__attribute__((constructor)) void load_file()
{
    printf("Constructor is called.\n");
    g_count = (int *)malloc(sizeof(int));
    if (g_count == NULL)
    {
    fprintf(stderr, "Failed to malloc memory.\n");
    }
}
__attribute__((destructor)) void unload_file()
{
    printf("destructor is called.\n");
    if (g_count)
    free(g_count);
}
int main()
{
    return 0;
}
```

参考: <https://gcc.gnu.org/onlinedocs/gcc-4.7.0/gcc/Function-Attributes.html>

### 库打桩机制 {#库打桩机制}

linux 链接器支持一个很强大的技术，称为库打桩机制，它允许你截获对共享库函数的调用，取而代之执行自己的代码。使用打桩机制，你可以追踪对某个特殊库函数的调用次数、验证和追踪它的输入和输出，甚至可以把它替换成一个完全不同的实现。

#### 编译时 {#编译时}

```plain
/main.c/
#include <stdio.h>
#include <malloc.h>

int main()
{
    int*p = malloc(32);

    free(p);

    return 0;

}
```

```plain
/*  malloc.h  */
#define malloc(size) my_malloc(size)
#define free(ptr)    my_free(ptr)


void * my_malloc(size_t size);
void * my_free(void* ptr);
```

```plain
/*  my_malloc.c  */
#ifdef COMPILE_TIME

#include <stdio.h>
#include <malloc.h>

void *my_malloc(size_t size)
{
    printf("enter my_malloc \n");
    void * ptr = malloc(size);
    printf("malloc(%d) = %p \n", (int ) size , ptr);

    return ptr;
}


void my_free( void* ptr )
{
    printf("enter my_free \n");
    free(ptr);
    printf("free (%p) \n" , ptr);
}

#endif
```

```plain
gcc -DCOMPILE_TIME -c my_malloc.c 
gcc -I. -o a.out main.c my_malloc.o
```

由于有 I. 参数，所以会进行打桩，它告诉C预处理器，在搜索通常的系统目录之前，现在当前的目录查找malloc.h

#### 链接时 {#链接时}

linux的静态链接器支持使用 –wrap f标志进行连接时打桩，这个标志告诉链接器，把对符号 f 的引用 解析成 **wrap\_f 还要把对符号** real\_f的 引用解析成 f 。

```plain
/*  link.c  */
#ifdef LINK_TIME

#include <stdio.h>

void * __real_malloc(size_t size);
void * __real_free(void * ptr);

void * __wrap_malloc(size_t size)
{
    printf("enter __wrap_malloc\n");
    void * ptr = __real_malloc(size);
    printf("malloc(%d) = %p \n", (int ) size , ptr);
    return ptr;
}


void __wrap_free(void *ptr)
{
    printf("enter __wrap_free\n");
    __real_free(ptr);
    printf("free (%p) \n" , ptr);
}

#endif
```

```plain
gcc -DLINK_TIME -c link.c
gcc -c main.c 
gcc -Wl,--wrap,malloc -Wl,--wrap,free -o a.out main.o  link.o
```

#### 运行时 {#运行时}

编译时打桩需要能够访问程序的源代码，链接时打桩需要能够访问程序的可重定位文件。不过，有一种机制能够在运行时打桩，只需能够访问可执行目标文件。 运行时打桩基于动态链接器的 LD\_PRELOAD 环境变量。

如果 LD\_PRELOAD 环境变量 被设置成为 共享库路径名的列表,当执行和加载程序的时候，当需要解析未定义的引用时，动态链接器会先搜索 LD\_PRELOAD 库，然后才搜索其他的库。

```plain
/*random.c   */

#include <stdio.h>
#include <stdlib.h>
#include <time.h>

int main(){
  srand(time(NULL));
  int i = 10;
  while(i--) printf("%d\n",rand()%100);
  return 0;
}

/*unrandom.c */
int rand(){
    return 42; //the most random number in the universe
}
```

```plain
gcc -shared -fPIC unrandom.c -o unrandom.so
export LD_PRELOAD=your_path
```

### 动态库加载特性 - got 替换 {#动态库加载特性-got-替换}

#### ELF文件格式 {#ELF文件格式}

ELF格式通常有linking view和execution view，即编译时和运行时，一般链接时统称 section , 运行时称segment，segment是运行时把权限相同的section合并了加载到内存，从视图上看，两个视图数据是一样的，只不过有两种形态。

```plain
             +-----------------+
        +----| ELF File Header |----+
        |    +-----------------+    |
        v                           v
+-----------------+      +-----------------+
| Program Headers |      | Section Headers |
+-----------------+      +-----------------+
     ||                               ||
     ||                               ||
     ||                               ||
     ||   +------------------------+  ||
     +--> | Contents (Byte Stream) |<--+
          +------------------------+
          
       +-------------------------------+
       | ELF File Header               |
       +-------------------------------+
       | Program Header for segment #1 |
       +-------------------------------+
       | Program Header for segment #2 |
       +-------------------------------+
       | ...                           |
       +-------------------------------+
       | Contents (Byte Stream)        |
       | ...                           |
       +-------------------------------+
       | Section Header for section #1 |
       +-------------------------------+
       | Section Header for section #2 |
       +-------------------------------+
       | ...                           |
       +-------------------------------+
       | ".shstrtab" section           |
       +-------------------------------+
       | ".symtab"   section           |
       +-------------------------------+
       | ".strtab"   section           |
       +-------------------------------+
```

ELF文件是连接编译链接与运行的数据存在，其中里面的 .text （代码段）、 .rodata（只读数据段） 、 .data （数据段）、 .symtab（符号表） 我们都耳熟能详，我们这里关心的是 下面几个段：

```plain
# 测试的 test 源码在下面
[root@VM_8_16_centos ~/hook]# readelf -S test | grep -E "plt|got|dyn" --color
  [ 5] .dynsym           DYNSYM           00000000004002b8  000002b8            // 动态符号表
  [ 6] .dynstr           STRTAB           0000000000400498  00000498            // 动态字符串
  [ 9] .rela.dyn         RELA             0000000000400598  00000598            // 数据有关的重定位表
  [10] .rela.plt         RELA             00000000004005b0  000005b0            // 函数有关的重定位表
  [12] .plt              PROGBITS         00000000004007a0  000007a0            // plt段
  [21] .dynamic          DYNAMIC          0000000000601e28  00001e28            // 动态信息表
  [22] .got              PROGBITS         0000000000601ff8  00001ff8            // 数据的全局偏移表
  [23] .got.plt          PROGBITS         0000000000602000  00002000            // 函数的全局偏移表
```

先直观的看一下有这几个表，他们基本都与重定位有关，需要看下重定位的概念。  
![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/dynlink.png)

#### 重定位与动态链接 {#重定位与动态链接}

当多个 .o 文件链接或 运行时需要动态库的时候，都有重定位的概念，在链接的时候，多个.o之间 相互依赖的变量和函数 要找到实际的地址， 同样运行时依赖动态库中的函数，一般是记录在全局偏移表中，运行之前或运行时 找到实际地址，记录到偏移表，运行的时候通过 全局偏移表找到实际地址，从而执行。

###### 重定位表: {#重定位表}

重定位表是”.rel.dyn”和”.rel.plt”，它们分别相当于静态链接中的”.rel.data”和”.rel.text”。”.rel.dyn”实际上是对数据引用的修正，它所修正的位置相当于”.got “以及数据段；而”.rel.plt”则是对函数引用的修正，所修正的位置位于”.got.plt”。使用”readelf -r”命令，查看重定位表

```plain
[root@VM_8_16_centos ~/hook]# readelf -r test

Relocation section '.rela.dyn' at offset 0x598 contains 1 entries:
  Offset          Info           Type           Sym. Value    Sym. Name + Addend
000000601ff8  000c00000006 R_X86_64_GLOB_DAT 0000000000000000 __gmon_start__ + 0

Relocation section '.rela.plt' at offset 0x5b0 contains 19 entries:
  Offset          Info           Type           Sym. Value    Sym. Name + Addend
000000602018  000100000007 R_X86_64_JUMP_SLO 0000000000000000 puts + 0
000000602020  000200000007 R_X86_64_JUMP_SLO 0000000000000000 readlink + 0
000000602028  000300000007 R_X86_64_JUMP_SLO 0000000000000000 getpid + 0
000000602030  000400000007 R_X86_64_JUMP_SLO 0000000000000000 fclose + 0
000000602038  000500000007 R_X86_64_JUMP_SLO 0000000000000000 printf + 0
000000602040  000600000007 R_X86_64_JUMP_SLO 0000000000000000 snprintf + 0
000000602048  000700000007 R_X86_64_JUMP_SLO 0000000000000000 __assert_fail + 0
000000602050  000800000007 R_X86_64_JUMP_SLO 0000000000000000 __libc_start_main + 0
000000602058  000900000007 R_X86_64_JUMP_SLO 0000000000000000 memcmp + 0
000000602060  000a00000007 R_X86_64_JUMP_SLO 0000000000000000 fgets + 0
000000602068  000b00000007 R_X86_64_JUMP_SLO 0000000000000000 strcmp + 0
000000602070  000c00000007 R_X86_64_JUMP_SLO 0000000000000000 __gmon_start__ + 0
000000602078  000d00000007 R_X86_64_JUMP_SLO 0000000000000000 mprotect + 0
000000602080  000e00000007 R_X86_64_JUMP_SLO 0000000000000000 fopen + 0
000000602088  000f00000007 R_X86_64_JUMP_SLO 0000000000000000 strtok + 0
000000602090  001000000007 R_X86_64_JUMP_SLO 0000000000000000 strtoul + 0
000000602098  001100000007 R_X86_64_JUMP_SLO 0000000000000000 getpagesize + 0
0000006020a0  001200000007 R_X86_64_JUMP_SLO 0000000000000000 strstr + 0
0000006020a8  001300000007 R_X86_64_JUMP_SLO 0000000000000000 rand + 0
```

###### GOT 及 PLT 表 {#GOT-及-PLT-表}

在Linux下，GOT被拆分成”.got”和”.got.plt”2个表。其中”.got”用来保存全局变量引用的地址，”.got.plt”用来保存函数引用的地址  
GOT表项还保留了３个公共表项，也即got的前３项，分别保存：

got[0]: 本ELF动态段(.dynamic段）的装载地址  
got[1]：本ELF的link\_map数据结构描述符地址  
got[2]：\_dl\_runtime\_resolve函数的地址

操作系统设计了一段比较精巧的指令来实现延迟重定位，历史的版本应该是，进程运行的时候，如果依赖动态库，那么运行之前，需要把 程序依赖的动态库里面的每个变量和函数都 初始化 GOT表，这样的后果就是 如果依赖比较多，加载缓慢； 后来通过 PLT 设计了 延迟加载的功能， 主要思想是 第一次运行的时候，通过一段跳转指令， 转去动态链接器中的\_dl\_runtime\_resolve 函数查找，查找后写入GOT， 第二次的时候便可以直接访问 GOT，直接地址访问。 下面有使用 gdb 动态调试的过程， 实际过程中，可能直接 `disas _dl_runtime_resolve` 发现没有效果或者找不到函数， 按照内存查 可知 最新的函数名字上有所改变。

查看plt的内容，实际就是代码：

```plain
objdump -d test
```

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/plt_code.png)

gdb调试 plt懒加载过程：  
![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/pltgot1.png)

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/pltgot2.png)

#### 重定位类型及偏移表 {#重定位类型及偏移表}

我们如何计算GOT表应该偏移多少呢，又有哪些偏移的类型呢？ 参考： <http://www.ucw.cz/~hubicka/papers/abi/node19.html>

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/got_rel_type.png)

比如rand , 我们直接用 rel表的地址

#### GOT表项替换 {#GOT表项替换}

全局符号表(GOT表)hook实际是通过解析SO文件，将待hook函数在got表的地址替换为自己函数的入口地址，这样目标进程每次调用待hook函数时，实际上是执行了我们自己的函数。

导入表的hook有两种方法：

- 方法一：

　　通过解析elf格式，分析Section header table找出静态的.got表的位置，并在内存中找到相应的.got表位置，这个时候内存中.got表保存着导入函数的地址，读取目标函数地址，与.got表每一项函数入口地址进行匹配，找到的话就直接替换新的函数地址，这样就完成了一次导入表的Hook操作了。

- 方法二

　　通过分析program header table查找got表。导入表对应在动态链接段.got.plt（DT\_PLTGOT）指向处，但是每项的信息是和GOT表中的表项对应的，因此，在解析动态链接段时，需要解析DT\_JMPREL、DT\_SYMTAB，前者指向了每一个导入表表项的偏移地址和相关信息，包括在GOT表中偏移，后者为GOT表。

方法二的测试：

详细代码: <https://github.com/changan29/playcpp/tree/master/hook/got_hook>

### 内核调试接口 {#内核调试接口}

#### ptrace 系统调用 {#ptrace-系统调用}

有很多大家所常用的工具都基于ptrace来实现，如strace和gdb。

ptrace系统调从名字上看是用于进程跟踪的，它提供了父进程可以观察和控制其子进程执行的能力，并允许父进程检查和替换子进程的内核镜像(包括寄存器)的值。其基本原理是: 当使用了ptrace跟踪后，所有发送给被跟踪的子进程的信号(除了SIGKILL)，都会被转发给父进程，而子进程则会被阻塞，这时子进程的状态就会被系统标注为TASK\_TRACED。而父进程收到信号后，就可以对停止下来的子进程进行检查和修改，然后让子进程继续运行。

使用ptrace 可以动态调试进程，可以做到自定义gdb的某些功能，参考：  
<https://www.cnblogs.com/tangr206/articles/3094358.html>

### 跳转代码修改 {#跳转代码修改}

#### inline hook {#inline-hook}

详细代码: <https://github.com/changan29/playcpp/tree/master/hook/inline-hook>

inline-hook大致的原理：一般是在具体的代码函数头加一段跳转指令，跳转的地址在runtime的时候指定，然后执行该方法的时候，就跳到指定的函数，执行hook。

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/inlinehook_0.png)

更直观的理解是 比如找个函数，从直观的字节上理解：  
![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/inlinehook_1.png)

一般，我们使用jmp跳转来实现inline hook, 获取代码地址- 修改 函数内容 - 实现自定义跳转

```plain
void hooker::HookerX64::doHook(void *func,void *newAddr,void **origFunc) const {
  
      char *f = (char *)func;
      if (origFunc) {
          // find the return instruction retq: c3
          int index = 0;
          while (true) {
              if (static_cast<uint8_t>(f[index++]) == 0xc3 || index >= 1024) {
                  break;
              }
          }
  
          void *old = malloc(index);
         printf("HookerX64::doHookbackup : %p \n", old);
         if (old != nullptr)
         {
              memcpy(old, func, index);
              changeCodeAttrs(old, CODE_READ_ONLY);
              *origFunc = old;
          }
      }
  
      /*
      x64位下使用的跳转是
  
      jmp 或者 call模式。
      jmp共使用14个字节，0xFF2500000000为6个字节，目标地址为00000000`00000000为8字节。
  
      call模式
      0xff1500000002为6个字节，目标地址为00000000`00000000为8字节。
      */
  
  
      *(uint16_t *)&f[0] = 0x25ff;
      *(int *)&f[2] = 0x00000000;
      *(long *)&f[6] = (long)newAddr;
 
      printf("HookerX64::doHook newAddr: %p\n " ,  newAddr);
  }
```

#### inline-hook的注意点 {#inline-hook的注意点}

具体代码见附件。  
执行完inline hook , 保存了原函数，只不过，此时的原函数内容 被拷贝到了其他的地方，再次调用原函数的时候，有的时候会core

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/inline-hook-core.png)

为什么在开头调用一个函数就会core呢？ gdb 打印拷贝后的函数与原函数 对比，发现 调整后的指令如果有 %rip relative的地址（offset），那么这个地址是需要对应调整的。

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/inlinehook_adjust.png)

### 其他 {#其他}

- 基于虚拟机提供能力、Java语言特性、消息hook(特别操作系统支持)等机制，Windows及Android上常用的hook机制。
- Android的 xposed 、 jni hook等

这一部分最近没有使用，后面有用到再实践。

### 引用 {#引用}

- <http://www.ucw.cz/~hubicka/papers/abi/node19.html>
- <http://www.wireghost.cn/2015/04/01/ELF%E6%96%87%E4%BB%B6%E7%BB%93%E6%9E%84%E8%AF%A6%E8%A7%A3/>
- <https://www.cnblogs.com/goodhacker/p/9306997.html>
- <https://stevens.netmeister.org/631/elf.html>
- <https://github.com/liuyx/inline-hook>
- <https://github.com/zhougy0717/inject_got>
- 《程序员的自我修养》
- 《CSAPP 第3、7章》
