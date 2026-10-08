---
title: redis数据冗余与分片
date: '2021-04-19T19:40:39+08:00'
url: /2021/04/19/redis-replica-and-redundancy/
draft: false
categories:
- redis
tags:
- redis
---

### redis数据冗余与分片

#### 数据冗余-主从复制

Redis 提供了主从库模式，以保证数据副本的一致，主从库之间采用的是读写分离的方式。  
![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/rds-commu.jpg)

- 读操作：主库、从库都可以接收；
- 写操作：首先到主库执行，然后，主库将写操作同步给从库。

#### 主从库间如何进行第一次同步？

```plain
replicaof  172.16.19.3  6379
```

![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/redis-replicaof.jpg)

#### 断网增量更新

当主从库断连后，主库会把断连期间收到的写操作命令，写入 replication buffer，同时也会把这些操作命令也写入 repl\_backlog\_buffer 这个缓冲区。

![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/redis-diff-push.jpg)

#### 主从全量同步使用RDB而不使用AOF的原因

RDB文件内容是经过压缩的二进制数据（不同数据类型数据做了针对性优化），文件很小。而AOF文件记录的是每一次写操作的命令，写操作越多文件会变得很大，其中还包括很多对同一个key的多次冗余操作。

#### 哨兵机制

Redis-Sentinel是Redis官方推荐的高可用性(HA)解决方案，当用Redis做Master-slave的高可用方案时，假如master宕机了，Redis本身(包括它的很多客户端)都没有实现自动进行主备切换，而Redis-sentinel本身也是一个独立运行的进程，它能监控多个master-slave集群，发现master宕机后能进行自动切换。  
![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/rds-sentinel.png)  
![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/rds-sentinel-util.png)

##### 如何选定新主库？

- 优先级最高的从库得分高。
- 和旧主库同步程度最接近的从库得分高。
- ID 号小的从库得分高。

### 数据分片

```plain
redis-cli -h 172.16.19.3 –p 6379 cluster addslots 0,1
redis-cli -h 172.16.19.4 –p 6379 cluster addslots 2,3
redis-cli -h 172.16.19.5 –p 6379 cluster addslots 4
```

![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/redis-cluster.jpg)

#### 客户端如何定位数据？

![image](https://blog21-1251969284.cos.ap-shanghai.myqcloud.com/redis/redis-cluster-choose.jpg)
