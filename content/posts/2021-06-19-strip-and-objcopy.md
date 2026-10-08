---
title: 用strip和objcopy精简二进制
date: '2021-06-19T02:40:39+08:00'
url: /2021/06/19/strip-and-objcopy/
draft: false
categories:
- objcopy
tags:
- strip
- objcopy
---

#### 问题: 设置到 LD\_PRELOAD的动态库 找不到符号

在设置了优先加载的 动态库以后，先加载的库里面的函数与 后加载的库里面的函数可能是有冲突的，比如遇到的这个例子，先加在的动态库中其实不存在完整的函数实现，只有符号表，并不能执行，但是它被设置成先加载了。

![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/os/link-error-strip.png)

对于有多个动态库中存在同一个符号的，一定程度上，我们可以从某个库 中 裁剪去这个符号（函数），去加载另外一个库里面的函数。

#### agent\_api\_init 在静态库和动态库中都存在

```plain
# nm 3rdparty/libs/libtbus.a | grep agent_api_init
                 U agent_api_init
# nm build/build/libhook.so | grep agent_api_init
                 U agent_api_init
```

#### 使用strip 可以指定删减 符号表

去除前后的比较  

```plain
[root@VM_8_16_centos ~]# strip -R agent_api_init libtbus.a -o nn.a
[root@VM_8_16_centos ~]# nm libtbus.a | grep agent_api_init
                 U agent_api_init
[root@VM_8_16_centos ~]# nm nn.a | grep agent_api_init
nm: tbus.o: no symbols
nm: tbus_addr_templet.o: no symbols
......
nm: tbus_route.o: no symbols
```

#### objcopy 不能去除 重定位的符号表

```plain
[root@VM_8_16_centos ~]# objcopy --strip-symbol agent_api_init libtbus.a nnn.a
objcopy: not stripping symbol `agent_api_init' because it is named in a relocatio
```

#### 也可以直接去掉动态库的符号

```plain
# strip -R agent_api_init libhook.so -o new_hook.so 
# nm new_hook.so | grep agent_api_init
  nm: new_hook.so: no symbols
```

去掉libhook.so里面的 无用符号以后，加载器找函数执行的时候就能找到正确的so， 不再有运行时链接错误。
