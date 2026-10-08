---
title: TCP的几个问题
date: '2019-11-25T01:40:39+08:00'
url: /2019/11/25/net_shake_hand/
draft: false
categories:
- tcp
tags:
- tcp
---

### IP header {#IP-header}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/ip_header.jpg)

### UDP header {#UDP-header}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/UDP-Header-Format.png)

### TCP header {#TCP-header}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/TCP-Header-Format.png)

### MTU & MSS {#MTU-amp-MSS}

MSS=MTU-IP包头长-TCP包头长

### 三次握手 {#三次握手}

交换初始序列号 交换TCP窗口大小信息

### 四次挥手 {#四次挥手}

可以三次挥手吗，可以，ack+fin一起发就是三次，启用延迟确认就有三次挥手

### 2MSL等待状态 {#2MSL等待状态}

TIME\_WAIT状态也称为2MSL等待状态。每个具体TCP实现必须选择一个报文段最大生存时间MSL（Maximum Segment Lifetime）。它是任何报文段被丢弃前在网络内的最长时间。我们知道这个时间是有限的，因为TCP报文段以IP数据报在网络内传输，而IP数据报则有限制其生存时间的TTL字段。

对一个具体实现所给定的MSL值，处理的原则是：当TCP执行一个主动关闭，并发回最后一个ACK，该连接必须在TIME\_WAIT状态停留的时间为2倍的MSL。这样可让TCP再次发送最后的ACK以防这个ACK丢失（另一端超时并重发最后的FIN）。

这种2MSL等待的另一个结果是这个TCP连接在2MSL等待期间，定义这个连接的插口（客户的IP地址和端口号，服务器的IP地址和端口号）不能再被使用。这个连接只能在2MSL结束后才能再被使用。

### Nagle算法 {#Nagle算法}

Nagle算法是时代的产物，因为当时网络带宽有限。而当前的局域网、广域网的带宽则宽裕得多，所以目前的TCP/IP协议栈默认将Nagle算法关闭，即通过SO\_NODELAY = 1

### 滑动窗口 {#滑动窗口}

性能: 停止等待协议 -> 滑动窗口  
协议: GBN and SR

### 流量控制 {#流量控制}

发送方不能发的太快，接收不过来

### 拥塞控制 {#拥塞控制}

成因: 中间路由器缓存有限，发送带宽有限，更多没必要的重传，多跳  
影响: 丢包 & 时延  
方案: 观察 时延 丢包判断是否拥塞(TCP采用)  
策略: 慢启动，先指数增，遇到丢包 /2 再线性增

### 状态图 {#状态图}

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/tcp_all_stat)
