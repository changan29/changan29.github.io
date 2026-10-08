---
title: linux_init_and_systemd
date: '2018-04-29T14:30:39+08:00'
url: /2018/04/29/linux_init_and_systemd/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 一个问题 {#一个问题}

今天在重启 nginx 服务的时候，注意到有这样一条输出  
![image](http://otc9g6h5k.bkt.clouddn.com/redirct_systemctl.png)

还有在设置 nginx 开机启动，一般的做法是 把脚本放在 /etc/init.d/ 下然后 用 chkconfig 管理  
不过这些都是”相对过时”做法了，由 service+chkconfig 变成 systemctl，背后是Linux 用户态进程管理体系的一次更改

- service+chkconfig 是 system V 的做法
- systemctl 是systemd 的做法

### 什么是 System V {#什么是-System-V}

```plain
UNIX System V是Unix操作系统众多版本中的一支。
它最初由AT&T开发，在1983年第一次发布，因此也被称为AT&T System V。
一共发行了4个System V的主要版本：版本1、2、3和4。
System V Release 4，或者称为SVR4，是最成功的版本，成为一些UNIX共同特性的源头，
例如“SysV 初始化脚本”（/etc/init.d），用来控制系统启动和关闭，
System V Interface Definition（SVID）是一个System V如何工作的标准定义。
```

### 什么是 是systemd {#什么是-是systemd}

```plain
systemd是Linux电脑操作系统之下的一套中央化系统及设置管理程序（init），
包括有守护进程、程序库跟应用软件，
由Lennart Poettering带头开发。其开发目标是提供更优秀的框架以表示
系统服务间的依赖关系，并依此实现系统初始化时服务的并行启动，
同时达到降低Shell的系统开销的效果，最终代替现在常用的System V与BSD风格init程序。
目前绝大多数的Linux发行版都已采用systemd代替原来的System V。
```

[参考: https://zh.wikipedia.org/wiki/Systemd](https://zh.wikipedia.org/wiki/Systemd)

### 趋势 {#趋势}

systemd已纳入众多Linux发行版的软件源中，以下简表：  
默认init程序为systemd的发行版

- Fedora 15及后续版本
- openSUSE 12.1 及后续版本
- Red Hat Enterprise Linux 7及后续版本，包括其派生品CentOS、Scientific Linux、Oracle Linux等
- Arch Linux在2012年10月13日将systemd-sysvcompat纳入base软件组，自此Arch Linux默认安装完即以systemd为init程序[13]，同时也提供了与Arch自带启动脚本兼容用的systemd启动脚本包以方便用户，使用户能“开箱即用”
- Chakra GNU/Linux，在2012.10的光盘映像档发布后默认使用systemd。
- Debian GNU/Linux，在2014年的技术委员会的init系统投票中决定在Debian 8“Jessie”中以Linux为核心的版本转换到systemd[16]。
- Ubuntu 15.04及后续版本[17]

### 发展的过程 {#发展的过程}

发展的过程引用一片文章，讲的非常细，收藏

[LINUX PID 1 和 SYSTEMD](https://coolshell.cn/articles/17998.html)
