---
title: concurrent overview
date: '2018-09-16T01:01:39+08:00'
url: /2018/09/16/concurrent_overview/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 并发 {#并发}

#### 并发模型 {#并发模型}

**多进程**

- 简单
- 开销比较大

–例子: apache

> Apache的server为process-based server ,也就是基于多进程的HTTP服务器，它需要对每个用户请求创建一个子进程进行响应，这样的缺点是，如果并发的请求非常多(在大型门户网站很常见)，就会需要非常多的进程，从而占用极多的cpu资源和内存。因此对于并发处理不是Apache的强项。

**多线程**

- 开销依旧比较大
- 我本来有一个问题，用多程线了后，现在两个题问了有我 (哈哈哈)

–例子: MySQL

> 一个连接一个线程模型：适用场景，连接少，且逻辑复杂。例如mysql采用此模型，一个连接一个线程。模型的一些小变体是线程采用线程池，避免创建销毁线程的开销

**基于回调的非阻塞/异步IO**

> 阻塞调用是指调用结果返回之前，当前线程会被挂起。调用线程只有在得到结果之后才会返回。  
> 非阻塞调用指在不能立刻得到结果之前，该调用不会阻塞当前线程。

- 比较难理解，流程是不顺序的

– 例子: Nginx

> Nginx 的Worker进程数一般设置与CPU核数相同,每进来一个request，会有一个worker进程去处理。但不是全程的处理，处理到什么程度呢？处理到可能发生阻塞的地方，比如向上游（后端）服务器转发request，并等待请求返回。那么，这个处理的worker不会这么傻等着，他会在发送完请求后，注册一个事件：“如果upstream返回了，告诉我一声，我再接着干”。于是他就休息去了。此时，如果再有request 进来，他就可以很快再按这种方式处理。而一旦上游服务器返回了，就会触发这个事件，worker才会来接手，这个request才会接着往下走。这是因为webserver刚好属于网络io密集型应用，不算是计算密集型。

> memcache也是类似的，它的多线程完全是为了利用多个cpu能力.

###### tips: {#tips}

nginx 使用的 epoll 是非阻塞的，但不是异步IO。  
glibc 的 aio 有 bug , kernel 的 aio 只能以 O\_DIRECT 方式做直接 IO , libeio 也是 beta 阶段。epoll 是成熟的，但是 epoll 本身是同步的。Linux 上目前没有像 IOCP 这样的成熟异步 IO 实现。

**协程**

- 开销小
- 需要自己实现调度器

全异步模型是性能最好，同时也是开发难度最高的模型。为了追求更好的性能，许多语言例如C++,C#,GO,nodejs,python都尝试简化此模型的编程，推出了支持异步编程的语言特性。

C++的future，promise

C#的async/await

GO的goroutine

nodejs的Promise=>generator=>async/await

python的yield

在这些特性的支持下，全异步编程可以做到与同步编程非常接近，可读性良好。其中GoLang是近年来新推出的语言，专门为高并发设计了goroutine，对此模型的支持最佳，使用体验也最好。

这里理解是这样的，你自己的用户态进程，如果能自己控制遇到系统调用也不会切换到 内核，而是继续执行你自己的指令；对于语言来说，**==让编译器知道这个函数就是阻塞的==**，在用户态自己调度就好了 参考了轮子哥的观点: [如何深入理解多进程，多线程，非阻塞/异步IO(callback) 以及Coroutine模型？](https://www.zhihu.com/question/21896633)

##### 挖坑 {#挖坑}

[golang的调度](https://morsmachine.dk/go-scheduler)

### 附: 阻塞非阻塞与同步异步的区别 {#附-阻塞非阻塞与同步异步的区别}

##### 先说结论 {#先说结论}

阻塞非阻塞 都是 同步io 不需要用户态进程任何阻塞的才是 异步IO  
这里参考 陈硕的回答: [怎样理解阻塞非阻塞与同步异步的区别？](https://www.zhihu.com/question/19732473)  
![image](http://otc9g6h5k.bkt.clouddn.com/io.png)

— @Sun Sep 16 01:02:39 CST 2018
