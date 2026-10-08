---
title: Elasticsearch实践
date: '2018-09-07T20:54:32+08:00'
url: /2018/09/07/elasticsearch/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 概览 {#概览}

全文搜索属于最常见的需求，开源的 Elasticsearch （以下简称 Elastic）是目前全文搜索引擎的首选。

它可以快速地储存、搜索和分析海量数据。维基百科、Stack Overflow、Github 都采用它。

没有写具体的安装过程以及集群，这些资料很多，参考:

- [全文搜索引擎 Elasticsearch 入门教程](http://www.ruanyifeng.com/blog/2017/08/elasticsearch.html)
- [集群搭建](https://www.jianshu.com/p/149a8da90bbc)

##### Elasticsearch 是什么 {#Elasticsearch-是什么}

- 一个分布式的实时文档存储，每个字段 可以被索引与搜索
- 一个分布式实时分析搜索引擎
- 能胜任上百个服务节点的扩展，并支持 PB 级别的结构化或者非结构化数据

##### 面向文档 {#面向文档}

Elasticsearch 是 面向文档 的，意味着它存储整个对象或 文档\_。Elasticsearch 不仅存储文档，而且 \_索引 每个文档的内容使之可以被检索。在 Elasticsearch 中，你 对文档进行索引、检索、排序和过滤–而不是对行列数据。这是一种完全不同的思考数据的方式，也是 Elasticsearch 能支持复杂全文检索的原因。

##### 分析数据 {#分析数据}

Elasticsearch 有一个功能叫聚合（aggregations），允许我们基于数据生成一些精细的分析结果。聚合与 SQL 中的 GROUP BY 类似但更强大。

##### 分布式特征 {#分布式特征}

Elasticsearch 尽可能地屏蔽了分布式系统的复杂性。这里列举了一些在后台自动执行的操作：

分配文档到不同的容器 或 分片 中，文档可以储存在一个或多个节点中  
按集群节点来均衡分配这些分片，从而对索引和搜索过程进行负载均衡  
复制每个分片以支持数据冗余，从而防止硬件故障导致的数据丢失  
将集群中任一节点的请求路由到存有相关数据的节点  
集群扩容时无缝整合新节点，重新分配分片以便从离群节点恢复

### 一些概念 {#一些概念}

##### Node 与 Cluster {#Node-与-Cluster}

Elastic 本质上是一个分布式数据库，允许多台服务器协同工作，每台服务器可以运行多个 Elastic 实例。

单个 Elastic 实例称为一个节点（node）。一组节点构成一个集群（cluster）。

##### Index {#Index}

Elastic 会索引所有字段，经过处理后写入一个反向索引（Inverted Index）。查找数据的时候，直接查找该索引。

所以，Elastic 数据管理的顶层单位就叫做 Index（索引）。它是单个数据库的同义词。每个 Index （即数据库）的名字必须是小写。

```plain
curl -X GET 'http://localhost:9200/_cat/indices?v'
```

##### Document {#Document}

Index 里面单条的记录称为 Document（文档）。许多条 Document 构成了一个 Index。

### 集群 {#集群}

- 集群是由一个或者多个拥有相同cluster.name配置的节点组成，它们共同承担数据和负载的压力。
- 我们往 Elasticsearch 添加数据时需要用到 索引 —— 保存相关数据的地方。 索引实际上是指向一个或者多个物理 分片 的 逻辑命名空间 。
- 一个 分片 是一个底层的 工作单元 ，它仅保存了 全部数据中的一部分。一个分片是一个 Lucene 的实例
- Elasticsearch 是利用分片将数据分发到集群内各处的。分片是数据的容器，文档保存在分片内，分片又被分配到集群内的各个节点里。

### 上手试一试 {#上手试一试}

ES跑起来还是很消耗资源的，自己的VPS 1G内存跑起来很吃力，用组里的几台机器搭了一套 ES集群，前端配置了 kibana,后台日志使用 filebeat,没用 logstash 可以看到已经有 7000w+ 条记录了， 按照用户id 和 游戏的 gameid 查询能做到秒查  
![image](http://otc9g6h5k.bkt.clouddn.com/es_cluster.png)

### 坑 {#坑}

- 多网卡的机器不能集群

如果你配置了 network.host是0.0.0.0，那么非localhost是可以访问的，但是这将不能与同网段的node 集群，如果既要外网访问，又要集群，那么可以按照下面这样配置

```plain
network.bind_host: "0.0.0.0"
network.publish_host: 123.5.6.7

vs 

network.host: 0.0.0.0
```

- vm.max\_map\_count 限制

需要设置 vm.max\_map\_count=262144

max\_map\_count文件包含限制一个进程可以拥有的VMA(虚拟内存区域)的数量。虚拟内存区域是一个连续的虚拟地址空间区域。在进程的生命周期中，每当程序尝试在内存中映射文件，链接到共享内存段，或者分配堆空间的时候，这些区域将被创建。调优这个值将限制进程可拥有VMA的数量。限制一个进程拥有VMA的总数可能导致应用程序出错，因为当进程达到了VMA上线但又只能释放少量的内存给其他的内核进程使用时，操作系统会抛出内存不足的错误。如果你的操作系统在NORMAL区域仅占用少量的内存，那么调低这个值可以帮助释放内存给内核用。

### TODO {#TODO}

- 查询语法

ES 有自己的查询语法，基于Apache Lucene，查询语法单独汇总

- 结合ES查询的后台服务例子

补个 demo

#### @Fri Sep 7 20:54:43 CST 2018 SHENZHEN {#Fri-Sep-7-20-54-43-CST-2018-SHENZHEN}
