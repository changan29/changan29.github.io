---
title: 文件描述符限制
date: '2019-03-30T16:08:45+08:00'
url: /2019/03/30/file_handler_limit/
draft: false
categories:
- OS
tags:
- OS
---

### 文件描述符限制 {#文件描述符限制}

#### 文件描述符 {#文件描述符}

Linux的核心思想之一就是”一切皆文件”。即Linux中所有的内容都是以文件的形式保存和管理的，它为不同类型的文件提供了统一的操作接口，对于不同类型的文件，我们都可以使用fopen()/fclose()/fwrite()/fread()等对这些文件进行读写处理。在Linux中，普通文件、目录、链接文件、字符设备、块设备以及网络套接字等等都以文件的形式存在。

文件描述符（文件句柄）是Linux内核为了高效管理已被打开的文件所创建的索引，它是一个非负整数，用于指代被打开的文件。文件打开后，所有对该文件的操作都通过文件描述符进行操作。

文件描述符是系统的重要资源，当打开过多文件时会出现把系统内存耗尽的情况，所以Linux提供了多种维度对文件描述符的使用上限进行限制，以避免该问题。

#### 文件描述符限制 {#文件描述符限制-1}

##### file-max {#file-max}

file\_max用于限制整个系统能够分配的文件描述符，即当前系统下所有用户下的所有进程，其分配的所有文件描述符之和不能超过该值。  

```plain
dailybuild@9.77.10.46:~>cat /proc/sys/fs/file-max 
3043564
# 临时修改，机器重启后失效
sysctl -w fs.file-max=102400

# 永久修改（直接在/etc/sysctl.conf文件最后添加一行fs.file-max=102400），执行下面命令使配置生效。
sysctl -p
```

##### file-nr {#file-nr}

```plain
dailybuild@9.77.10.46:~>cat /proc/sys/fs/file-nr
1952    0       3043564
```

file-nr是一个只读文件,它记录了当前系统中文件描述符的使用情况，第一个（6144）是系统已经分配的文件描述符数；第二个（0）是系统已经分配但未使用的文件描述符数；第三个是内核最大能分配的文件描述符数，与file-max一致。我们可以通过查看此数据判断文件描述符是够已经将要用尽和是否需要扩file-max字段。

##### nr\_open {#nr-open}

nr\_open用于限制单个进程可以分配的最大文件描述符，该限制对于系统上所有用户下的所有进程都生效，查看当时如下：  

```plain
dailybuild@9.77.10.46:~>cat /proc/sys/fs/nr_open
1048576
```

##### soft limit && hard limit {#soft-limit-amp-amp-hard-limit}

Soft limit 和 Hard limit# 在Linux的系统中对于进程(Process)会有一些限制，这就所谓的limit，在实际应用中最常见的就是对打开文件(Open Files)的限制，在配置web 服务如nginx时就会用到。在linux中这些限制是分为软限制(soft limit)和硬限制(hard limit)的。他们的区别就是软限制可以在程序的进程中自行改变(突破限制)，而硬限制则不行(除非程序进程有root权限)

```plain
ulimit -Sn 查看的是软限制

ulimit -Hn 查看的是硬限制
```

在这个文件中我们可以对cpu, 内存等等设置限制

```plain
sudo vim /etc/security/limit.conf
```
