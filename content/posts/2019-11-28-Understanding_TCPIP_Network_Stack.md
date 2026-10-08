---
title: Understanding TCPIP Network Stack
date: '2019-11-28T23:58:45+08:00'
url: /2019/11/28/Understanding_TCPIP_Network_Stack/
draft: false
categories:
- TCPIP
tags:
- TCPIP
---

### 概述 {#概述}

TCP & socket 总结，

### 网卡 network interface card {#网卡-network-interface-card}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/network_interface.png)

网卡需要有驱动才能工作，驱动是加载到内核中的模块，负责衔接网卡和内核的网络模块，驱动在加载的时候将自己注册进网络模块，当相应的网卡收到数据包时，网络模块会调用相应的驱动程序处理数据。

内核如何从网卡接受数据，传统的经典过程：

- 1、数据到达网卡；
- 2、网卡产生一个中断给内核；
- 3、内核使用I/O指令，从网卡I/O区域中去读取数据；

但是，这一种方法，有一种重要的问题，就是大流量的数据来到，网卡会产生大量的中断，内核在中断上下文中，会浪费大量的资源来处理中断本身。所以，一个问 题是，“可不可以不使用中断”，这就是轮询技术，所谓NAPI技术，说来也不神秘，就是说，内核屏蔽中断，然后隔一会儿就去问网卡，“你有没有数据 啊？”

另一个问题，就是从网卡的I/O区域，包括I/O寄存器或I/O内存中去读取数据，这都要CPU去读，也要占用CPU资源，“CPU从I/O区域 读，然后把它放到内存（这个内存指的是系统本身的物理内存，跟外设的内存不相干，也叫主内存）中”。于是自然地，就想到了DMA技术——让网卡直接从主内 存之间读写它们的I/O数据。

#### 优化后 {#优化后}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/ring-buffer.png)

NIC (network interface card) 在系统启动过程中会向系统注册自己的各种信息，系统会分配 Ring Buffer 队列也会分配一块专门的内核内存区域给 NIC 用于存放传输上来的数据包。struct sk\_buff 是专门存放各种网络传输数据包的内存接口，在收到数据存放到 NIC 专用内核内存区域后，sk\_buff 内有个 data 指针会指向这块内存。

Ring Buffer 队列内存放的是一个个 Packet Descriptor ，其有两种状态： ready 和 used 。初始时 Descriptor 是空的，指向一个空的 sk\_buff，处在 ready 状态。当有数据时，DMA 负责从 NIC 取数据，并在 Ring Buffer 上按顺序找到下一个 ready 的 Descriptor，将数据存入该 Descriptor 指向的 sk\_buff 中，并标记槽为 used。因为是按顺序找 ready 的槽，所以 Ring Buffer 是个 FIFO 的队列。

#### 1、首先，内核在主内存中为收发数据建立一个环形的缓冲队列（通常叫DMA环形缓冲区） {#1、首先，内核在主内存中为收发数据建立一个环形的缓冲队列（通常叫DMA环形缓冲区）}

Linux内核中，用skb来描述一个缓存，所谓分配，就是建立一定数量的skb，然后把它们组织成一个双向链表。sk\_buff结构或skb结构代表一个数据包,它们也变得更复杂了。

sk\_buff\_structure

##### 包含数据和元数据 {#包含数据和元数据}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/sk_buff_structure.png)

一些必要的信息比如头和内容长度被保存在元数据区。例如，在图6中，mac\_header、network\_header和transport\_header都有相应的指针，指向链路头、IP头和TCP头的起始地址。这种方式让TCP协议处理过程变得简单。

##### 如何增加或删除头 {#如何增加或删除头}

数据包在网络栈的各层中上升或下降时会增加或删除数据头。为了更有效率的处理而使用了指针。例如，要删除链路头只需要修改head pointer的值。

##### 如何合并或切分数据包 {#如何合并或切分数据包}

为了更有效率的执行把数据包增到或从socket缓冲区中删除这类操作而使用了链表，或者叫数据包链。next和prev指针用于这个场景。

##### 快速分配和释放 {#快速分配和释放}

无论何时创建数据包都会分配一个数据结构，此时会用到快速分配器。比如，如果数据通过10Gb的以太网传输，每秒会有超过一百万个对象被创建和销毁。

#### 2、内核将这个缓冲区通过DMA映射，把这个队列交给网卡 {#2、内核将这个缓冲区通过DMA映射，把这个队列交给网卡}

内核操作，双向映射

#### 3、网卡收到数据，就直接放进这个环形缓冲区了——也就是直接放进主内存了；然后，向系统产生一个中断 {#3、网卡收到数据，就直接放进这个环形缓冲区了——也就是直接放进主内存了；然后，向系统产生一个中断}

硬件行为

#### 4、内核收到这个中断，就取消DMA映射，这样，内核就直接从主内存中读取数据 {#4、内核收到这个中断，就取消DMA映射，这样，内核就直接从主内存中读取数据}

### TCP tcp control block {#TCP-tcp-control-block}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/tcp_control_block.png)

TCP 协议的所有状态信息都保存在 tcp\_sock 中，比如序列号、接受窗口、拥塞控制、和重传计时器都保存在 tcp\_sock。

socket 发送缓存和接收缓存就是 sk\_buff 列表，它们也保存了 tcp\_sock 信息。dst\_entry 和 IP 路由结果是为了避免频繁地进行路由。dst\_entry 允许快速搜索 ARP 结果，也就是目的 MAC 地址。dst\_entry 是路由表的一部分，路由表的结构非常复杂，这篇文章不会讨论。报文传输要使用的网络设备也能通过 dst\_entry 搜索到，网络设备对应的结构体是 net\_device。

因此，通过 file 结构体和各级指针就能找到处理 TCP 报文需要的结构体（从文件一直到网络驱动），各种结构体的大小之和也就是 TCP 连接要占用的内存大小，这个值在几 KB（当然不包括报文的数据）。对着更多的功能加进来，这个内存使用也会逐渐增加。

最后，我们来看看 TCP 连接查找表（lookup table），这是一个哈希表，用来搜索接收到的报文属于哪个 TCP 连接。哈希值是通过报文的  四元组和 Jenkins 哈希算法计算的，据说使用这个算法是为了应对对哈希表的攻击。

### REF {#REF}

- [理解TCP](https://cizixs.com/2017/07/27/understand-tcp-ip-network-stack/)
- [Understanding TCP/IP Network Stack](https://www.cubrid.org/blog/understanding-tcp-ip-network-stack)
- [link](http://blog.2baxb.me/archives/1190)
