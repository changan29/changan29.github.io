---
title: 写一个操作系统_04 内核调试环境搭建
date: '2019-03-20T21:08:45+08:00'
url: /2019/03/20/make_a_linux_4/
draft: false
categories:
- kernel
- OS
tags:
- OS
- kernel
---

### 环境准备工作

有一台win10的笔记本，调试想选择在Linux环境下，选择配置虚拟机，然后安装bochs调试内核  
需要准备的软件:

- VirtualBox和增加工具包
- 安装CentOS 6.8minimal
- 设置CentOS的网卡，使之能上网
- 用yum安装开发工具和GNOME
- 安装bochs
- 进入GNOME，在GNOME中运行bochs

#### 虚拟机

##### CentOS配置

- 虚拟机设置中连接方式为NAT
- 因为安装的CentOS是minimal版本，所以网卡没有默认设置为启动

  ```plain
  vi /etc/sysconfig/network-scripts/ifcfg-eth0
  修改:

  ONBOOT="yes"  
  MM_Controlled="no"
  # ifconfig eth0 192.168.0.2 netmask 255.255.255.0 //修改ip地址、子网掩码
  # route add default gw 192.168.0.1 dev eth0　　	  //修改网关  
  # hostname centos 　　					          //修改计算机名  
  # service network restart  				          //重启网络
  # ping 192.168.1.1					              //ping网关应正常
  # ping www.163.com				                  //ping外网应正常

  保证能上外网
  ```

##### 需要的编译软件

```plain
# yum groupinstall "Development Tools"	//gcc, gcc-c++等常用软件都会装上
# yum -y install vim*
# yum install ncurses-devel zlib-devel texinfo gtk2-devel qt-devel
# yum install tcl-devel   tk-devel kernel-headers kernel-devel				
# yum install gtk*			//前一行的gtk2-dvel已经安装，此行可不执行
# yum install build-essential nasm
```

##### 安装Gnome

```plain
# yum grouplist ｜more  					//察看Gnome安装包的名称
# yum groupinstall "X Window System"		// 安装基本的X系统组件
# yum groupinstall "Desktop "		
# startx									//进入Gnome
```

##### 编译安装bochs

我是下载了bochs-2.6.7.tar.gz的包通过虚拟机的共享目录转到CentOS下，然后在linux下操作

```plain
./configure  --prefix=/data/bochs/ --enable-debugger --enable-disasm
./configure --enable-iodebug --enable-x86-debugger --with-x --with-x11 
make 
make install  // sudo
```

##### bochs运行报错

遇到两个错误:

- PANIC ../share/doc/bochs/bochsrc-sample.txt:187: cpu directive malformed.

    解决的办法: 配置里面的CPU cpu: model=core2\_penryn\_t9600目前bochs不支持，需要找支持的CPU

```plain
./bochs -help cpu
Supported CPU models:

bx_generic
pentium
pentium_mmx
amd_k6_2_chomper
p2_klamath
p3_katmai
p4_willamette
core_duo_t2400_yonah
atom_n270
```

- PANIC bochsrc.txt:917: Bochs is not compiled with lowlevel sound support

   声卡不支持，那先不需要声卡，把配置中的声卡先去掉不用

##### 从硬盘启动

![image](http://201903-1251969284.cosgz.myqcloud.com/bochs_harddisk.png)

这个报错跟上面的不一样，接下来要从启动盘启动，没有找到磁盘，接下来需要我们制作启动盘了。

##### 制作启动盘

执行 bximage 按照步骤可以制作一个启动盘，假如制作的启动盘为 hd30M.img  
修改配置如下  

```plain
# 设置硬盘
ata0: enabled=1,ioaddr1=0x1f0, ioaddr2=0x3f0, irq=14

# 启动盘信息
ata0-master: type=disk, path="/data/bochs/hd30M.img", mode=flat
```

再次启动:  
![image](http://201903-1251969284.cosgz.myqcloud.com/bochs_create.png)

出现这用错误的原因是因为我们现在的硬盘还只是一个空的硬盘，没有任何数据，又如何能够运行呢，CPU一下就跑没影了。

##### Next

接下来我们要加上这个启动设备，拉起内核。期待… :-)

![image](http://201903-1251969284.cosgz.myqcloud.com/boy.gif)
