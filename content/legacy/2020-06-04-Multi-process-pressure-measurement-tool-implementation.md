---
title: 多进程游戏压测工具的实现
date: '2020-06-04T22:31:45+08:00'
url: /2020/06/04/Multi-process-pressure-measurement-tool-implementation/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 概述 {#概述}

![image](https://202003-1251969284.cos.ap-shanghai.myqcloud.com/game_c2s.png)

首先对于游戏的业务，一般是玩家登陆到大厅，有一些任务、物品、好友、排行榜、聊天这种交互，其次是玩家与玩家之前的匹配与对局。以Moba游戏为例，玩家主要的行为就是登陆后进行匹配，匹配到水平差不多的10个人，分为两队，每组5个人创建对局进行pvp战斗，玩家的操作以指令的方式由客户端发到服务器。 大厅中客户端与服务器的连接是TCP连接，对局中玩家的操作更关注实时性，一般的用可靠UDP进行通信。

##### 接入层 {#接入层}

对于大厂而言，客户端接入服务器一般是连接一个接入层，经接入层转发请求到游戏服务器，接入层的目的也是标准化（工业化）的一部分，使得每个游戏都可以复用，只需要写业务逻辑从而不用重写网络连接的代码。我们模拟玩家的行为，也就是往游戏服务器对应的接入层发包。

##### 内存池缓存数据 {#内存池缓存数据}

对于游戏，玩家身上有很多属性，比如各种物品、分数、赛季信息、抽奖信息、VIP等级、头像框id、注册时间、各种活动的数据等。而游戏在开服或者有一些活动的时候，也是玩家集中登陆集中操作的时候，服务器的做法一般是在玩家登陆的服务器 建立内存池，将玩家的数据缓存到内存，纯内存的操作处理玩家数据比较高效。

##### 通信协议 {#通信协议}

王者荣耀和吃鸡应该也是使用Protocol Buffers  
<https://developers.google.com/protocol-buffers>

#### 服务器逻辑 {#服务器逻辑}

服务器主循环：  
![image](https://202003-1251969284.cos.ap-shanghai.myqcloud.com/tapp_svr_loop.png)

#### 压测概述与思路 {#压测概述与思路}

![image](https://202003-1251969284.cos.ap-shanghai.myqcloud.com/gameperf.png)

我们通过模拟真实玩家发起压力测试，有的场景比较简单，比如查询排行榜，只要构造了排行榜的数据，发起查询请求即可；但是也有比较复杂的场景，比如巅峰赛观战，比如需要8000人在对局，与此同时，有40000人在观战，正在对局中的人需要模拟玩家的行为，移动、一技能、装备、聊天、弹幕等操作，此时我们需要保持着和多个进程的连接。

同时，每个进程保持的连接可能是有限的，比如接入层sdk的限制，连接层和逻辑层需要做到可扩展。

#### 共享内存通道 {#共享内存通道}

为支持逻辑层和连接层的可扩展，每增加一个逻辑层或者连接层，可以增加一个共享内存通道，  
![image](https://202003-1251969284.cos.ap-shanghai.myqcloud.com/bus_channel.png)

详细的代码在这里，可编译运行：

<https://github.com/changan29/playcpp/tree/master/bus_channel>

#### 定时器 {#定时器}

一个Timer的实现需要具备以下几个行为:

- StartTimer(Interval, ExpiryAction)

注册一个时间间隔为 Interval 后执行 ExpiryAction 的定时器实例，其中，返回 TimerId 以区分在定时器系统中的其他定时器实例。

- StopTimer(TimerId)

根据 TimerId 找到注册的定时器实例并执行 Stop 。

- PerTickBookkeeping()

在一个 Tick 时间粒度内，定时器系统需要执行的动作，它最主要的行为，就是检查定时器系统中，是否有定时器实例已经到期。

具体的代码实现思路就是：在StartTimer的时候，把 当前时间 + Interval 作为key放入一个容器，然后在Loop的每次Tick里，从容器里面选出一个最小的key与当前时间比较，如果key小于当前时间，则这个key代表的timer就是expired，需要执行它的ExpiryAction(一般为回调)。

##### 链表的实现 {#链表的实现}

- 精度是 1ms
- 最长时间是10min,延长时间可以增加 slot数量,slot时间的间隔是 1ms
- 通过继承Timer父类，在子类重写timeout实现 超时回调
- 每次都需要遍历超过时间的所有链表，时间复杂度为O(n)

详细的代码在这里，可编译运行：

<https://github.com/changan29/codeLib/tree/master/timer>

#### 协程 {#协程}

协程框架使用的是 ucontext，参考之前写的这篇：  
<http://www.oneyearago.me/2020/05/26/ucontext_01/>

实现的机制是：

1. 每个玩家抽象成机器人，有一个Robot类，Robot类继承 协程类
2. 每个Robot有自己的行为，logic进程的代码就是robot的行为
3. 把每个玩家抽象成一个协程，方便写逻辑，相互独立，等待调度
4. 初始化的时候创建一批玩家（协程），按照配置等待一定的时间开始执行（比如10ms）,结合定时器，可以按照固定的频率发包（比如1秒发送200个）
5. 调度程序 调度程序比较简单，比如有了响应包或者定时器事件到了，根据timer\_id或者pkg\_head\_src标记等 resume 协程
