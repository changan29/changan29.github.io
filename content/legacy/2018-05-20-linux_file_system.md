---
title: linux_file_system
date: '2018-05-20T18:07:39+08:00'
url: /2018/05/20/linux_file_system/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 引言 {#引言}

在学校的时候泛泛读过一遍 apue,其中的部分知识只是有个大概印象，其实我个人对底层技术还是有热情和追求的 哈哈，打算把经典的书籍结合遇到的场景重读一遍，先拿 Linux 文件系统练习下。代码参考的是Linux早期的代码，没有现代内核的高级特性，VFS这部分只有介绍。

### 主要思路 {#主要思路}

写自己的总结之前在网上找了一些别人的总结，很多人很喜欢从宏观着手，上来就介绍 VFS,讲文件系统的分层然后具体到 ext2/ext3/ext4 文件系统，讲这部分文件系统是如何结构化磁盘的以方便文件的管理，再带一部分磁盘的格式化，inode节点，超级块结构等，这是一部分人；另一部分是反过来，从磁盘讲起，到VFS

以上两种方式各有优点，不过会有一种流水账的感觉，如果有具体的例子，会印象更深刻一些。我的思路是从 代码的角度出发，操作文件必经的 操作是 open 系统调用，然后从一个进程的角度看文件系统，这样会涉及到 内核处理文件的细节，自然会知道描述文件的各种结构，这种顺序的思路 印象也相对深刻

### 准备工作 {#准备工作}

- 内核源码

  查看系统调用内部 文件系统处理的过程需要看内核的代码，现代的Linux2.6以上的内核已经很复杂了，而且经过了多轮优化，不一定能看懂。。决定拿比较早期的内核 Linux0.11 版本的入手，简单而且资料多。

  代码在这里: [linux-0.11](https://github.com/karottc/linux-0.11)
- 系统调用

  以前写过一篇系统调用的: <http://www.oneyearago.me/2018/05/08/apue-again-system-call-and-std/>  
   系统调用以中断的方式进行，Linux的系统调用通过int 80h实现，用系统调用号来区分入口函数。
- Linux 一切皆文件

  - 首先通常在windows中是文件的东西,它们在linux中也是文件
  - 其次一些在windows中不是文件的东西, 比如进程, 磁盘, 也被抽象成了文件. 你可以使用访问文件的方法访问它们获得信息.
  - 再其次,一些很离谱的东西, 比如管道, 比如/dev/zero(一个可以读出无限个0的文件) /dev/null(一个重定向进去之后就消失了的文件). 它们也是文件
  - 再再其次, 类似于socket这样的东西, 使用的接口跟文件接口也是一致的.

    带来的好处就是, 你可以使用同一套api(read, write)和工具(cat , 重定向, 管道)来处理unix中大多数的资源.这就使得组合了简单的命令和字符处理工具(awk, sed)之后, shell脚本就能发挥出强大的功能.
- Linux文件类型：

  - 1.普通文件 # xxx.log
  - 2.目录 # /usr/ /home/
  - 3.字符设备文件 # /dev/tty的属性是 crw-rw-rw- ，注意前面第一个字符是 c ，这表示字符设备文件,比如猫等串口设备
  - 4.块设备文件 # /dev/hda1 的属性是 brw-r—– ，注意前面的第一个字符是b，这表示块设备，比如硬盘，光驱等设备
  - 5.套接字文件 # /var/lib/mysql/mysql.sock srwxrwxrwx
  - 6.管道 # pipe
  - 7.符号链接文件 # softlink…

### 文件操作分析 {#文件操作分析}

- open -> sys\_open

  打开一个文件不论哪种语言都会有个 open(),在编译和解释器执行的时候一定会调用系统调用 open(),所以系统调用一定是实现 这个open() 的，我们来找一下，在代码 linux-0.11-master/lib/open.c

  ```plain
  int open(const char * filename, int flag, ...)
  {
  	register int res;
  	va_list arg;

  	va_start(arg,flag);
  	__asm__("int $0x80"
  		:"=a" (res)
  		:"0" (__NR_open),"b" (filename),"c" (flag),
  		"d" (va_arg(arg,int)));
  	if (res>=0)
  		return res;
  	errno = -res;
  	return -1;
  }
  ```

  0x80 是系统调用对应的终端指令，\_\_NR\_open 是 对应的调用号，定义在linux-0.11-master/include/unistd.h

  ```plain
  #define __NR_setup	0	/* used only by init, to get system going */
  #define __NR_exit	1
  #define __NR_fork	2
  #define __NR_read	3   
  #define __NR_write	4
  #define __NR_open	5   // <- open() call 
  #define __NR_close	6
  #define __NR_waitpid	7
  #define __NR_creat	8
  #define __NR_link	9
  #define __NR_unlink	10
  #define __NR_execve	11
  #define __NR_chdir	12
  #define __NR_time	13
  #define __NR_mknod	14
  #define __NR_chmod	15
  ......
  ```

  与这些中断调用号对应是 一个函数指针数组:

  ```plain
  fn_ptr sys_call_table[] = { sys_setup, sys_exit, sys_fork, sys_read,
  sys_write, sys_open, sys_close, sys_waitpid, sys_creat, sys_link,
  sys_unlink, sys_execve, sys_chdir, sys_time, sys_mknod, sys_chmod,
  sys_chown, sys_break, sys_stat, sys_lseek, sys_getpid, sys_mount,
  sys_umount, sys_setuid, sys_getuid, sys_stime, sys_ptrace, sys_alarm,
  sys_fstat, sys_pause, sys_utime, sys_stty, sys_gtty, sys_access,
  sys_nice, sys_ftime, sys_sync, sys_kill, sys_rename, sys_mkdir,
  sys_rmdir, sys_dup, sys_pipe, sys_times, sys_prof, sys_brk, sys_setgid,
  sys_getgid, sys_signal, sys_geteuid, sys_getegid, sys_acct, sys_phys,
  sys_lock, sys_ioctl, sys_fcntl, sys_mpx, sys_setpgid, sys_ulimit,
  sys_uname, sys_umask, sys_chroot, sys_ustat, sys_dup2, sys_getppid,
  sys_getpgrp, sys_setsid, sys_sigaction, sys_sgetmask, sys_ssetmask,
  sys_setreuid,sys_setregid };
  ```

```
可以看到 sys_open 正好是在 第6个，必须要对应上的，所以说，我们 open一个文件，实际上最后是交给了 sys_open()
```

- 内核操作打开文件 (进程中维护文件指针数组)

  我们来看下 sys\_open

  ```plain
  int sys_open(const char * filename,int flag,int mode)
  {
  	struct m_inode * inode;
  	struct file * f;
  	int i,fd;

      // 首先对参数进行处理。将用户设置的文件模式和屏蔽码相与，产生许可的文件模式
      // 为了为打开文件建立一个文件句柄，需要搜索进程结构中文件结构指针数组，以查
      // 找一个空闲项。空闲项的索引号fd即是文件句柄值。若已经没有空闲项，则返回出错码。
  	mode &= 0777 & ~current->umask;
  	for(fd=0 ; fd<NR_OPEN ; fd++)
  		if (!current->filp[fd])
  			break;
  	if (fd>=NR_OPEN)
  		return -EINVAL;
      // 然后我们设置当前进程的执行时关闭文件句柄(close_on_exec)位图，复位对应的
      // bit位。close_on_exec是一个进程所有文件句柄的bit标志。每个bit位代表一个打
      // 开着的文件描述符，用于确定在调用系统调用execve()时需要关闭的文件句柄。当
      // 程序使用fork()函数创建了一个子进程时，通常会在该子进程中调用execve()函数
      // 加载执行另一个新程序。此时子进程中开始执行新程序。若一个文件句柄在close_on_exec
      // 中的对应bit位被置位，那么在执行execve()时应对应文件句柄将被关闭，否则该
      // 文件句柄将始终处于打开状态。当打开一个文件时，默认情况下文件句柄在子进程
      // 中也处于打开状态。因此这里要复位对应bit位。   	current->close_on_exec &= ~(1<<fd);
      // 然后为打开文件在文件表中寻找一个空闲结构项。我们令f指向文件表数组开始处。
      // 搜索空闲文件结构项(引用计数为0的项)，若已经没有空闲文件表结构项，则返回
      // 出错码。
  	f=0+file_table;
  	for (i=0 ; i<NR_FILE ; i++,f++)
  		if (!f->f_count) break;
  	if (i>=NR_FILE)
  		return -EINVAL;
      // 此时我们让进程对应文件句柄fd的文件结构指针指向搜索到的文件结构，并令文件
      // 引用计数递增1。然后调用函数open_namei()执行打开操作，若返回值小于0，则说
      // 明出错，于是释放刚申请到的文件结构，返回出错码i。若文件打开操作成功，则
      // inode是已打开文件的i节点指针。
  	(current->filp[fd]=f)->f_count++;
  	if ((i=open_namei(filename,flag,mode,&inode))<0) {
  		current->filp[fd]=NULL;
  		f->f_count=0;
  		return i;
  	}
      // 根据已打开文件的i节点的属性字段，我们可以知道文件的具体类型。对于不同类
      // 型的文件，我们需要操作一些特别的处理。如果打开的是字符设备文件，那么对于
      // 主设备号是4的字符文件(例如/dev/tty0)，如果当前进程是组首领并且当前进程的
      // tty字段小于0(没有终端)，则设置当前进程的tty号为该i节点的子设备号，并设置
      // 当前进程tty对应的tty表项的父进程组号等于当前进程的进程组号。表示为该进程
      // 组（会话期）分配控制终端。对于主设备号是5的字符文件(/dev/tty)，若当前进
      // 程没有tty，则说明出错，于是放回i节点和申请到的文件结构，返回出错码(无许可)。
      /* ttys are somewhat special (ttyxx major==4, tty major==5) */
  	if (S_ISCHR(inode->i_mode)) {
  		if (MAJOR(inode->i_zone[0])==4) {
  			if (current->leader && current->tty<0) {
  				current->tty = MINOR(inode->i_zone[0]);
  				tty_table[current->tty].pgrp = current->pgrp;
  			}
  		} else if (MAJOR(inode->i_zone[0])==5)
  			if (current->tty<0) {
  				iput(inode);
  				current->filp[fd]=NULL;
  				f->f_count=0;
  				return -EPERM;
  			}
  	}
      /* Likewise with block-devices: check for floppy_change */
      // 如果打开的是块设备文件，则检查盘片是否更换过。若更换过则需要让高速缓冲区
      // 中该设备的所有缓冲块失败。
  	if (S_ISBLK(inode->i_mode))
  		check_disk_change(inode->i_zone[0]);
      // 现在我们初始化打开文件的文件结构。设置文件结构属性和标志，置句柄引用计数
      // 为1，并设置i节点字段为打开文件的i节点，初始化文件读写指针为0.最后返回文
      // 件句柄号。
  	f->f_mode = inode->i_mode;
  	f->f_flags = flag;
  	f->f_count = 1;
  	f->f_inode = inode;
  	f->f_pos = 0;
  	return (fd);
  }
  ```

  解释一下这段代码，current 是指当前进程的 task\_struct 一个进程的 PCB，NR\_OPEN 是一个进程最多打开的文件个数，0.11版本的Linux最多只能打开20个。上面的这个 task\_struct 结构非常重要，他是一个进程的描述单位，在linux-0.11-master/include/linux/sched.h：

  ```plain
  struct task_struct {
  /* these are hardcoded - don't touch */
  	long state;	/* -1 unrunnable, 0 runnable, >0 stopped */
  	long counter;
  	long priority;
  	long signal;
  	struct sigaction sigaction[32];
  	long blocked;	/* bitmap of masked signals */
  /* various fields */
  	int exit_code;
  	unsigned long start_code,end_code,end_data,brk,start_stack;
  	long pid,father,pgrp,session,leader;
  	unsigned short uid,euid,suid;
  	unsigned short gid,egid,sgid;
  	long alarm;
  	long utime,stime,cutime,cstime,start_time;
  	unsigned short used_math;
  /* file system info */
  	int tty;		/* -1 if no tty, so it must be signed */
  	unsigned short umask;
  	struct m_inode * pwd;
  	struct m_inode * root;
  	struct m_inode * executable;
  	unsigned long close_on_exec;
  	struct file * filp[NR_OPEN];        // <-  see it , file pointer array
  /* ldt for this task 0 - zero 1 - cs 2 - ds&ss */
  	struct desc_struct ldt[3];
  /* tss for this task */
  	struct tss_struct tss;
  };
  ```

  也就是说，每个进程会维护一个打开文件的数组 struct file \* filp[NR\_OPEN]; 打开，把这个fd 传给用户空间，那么，这个file 结构又是如何组织的呢？
- 每个文件的信息是如何组织的  
  从进程中的 file 结构出发，我们看下文件结构是如何组织的 linux-0.11-master/include/linux/fs.h :

```plain
struct file {
	unsigned short f_mode;
	unsigned short f_flags;
	unsigned short f_count;
	struct m_inode * f_inode;
	off_t f_pos;
};
struct m_inode {
	unsigned short i_mode;
	unsigned short i_uid;
	unsigned long i_size;
	unsigned long i_mtime;
	unsigned char i_gid;
	unsigned char i_nlinks;
	unsigned short i_zone[9];
/* these are in memory also */
	struct task_struct * i_wait;
	unsigned long i_atime;
	unsigned long i_ctime;
	unsigned short i_dev;
	unsigned short i_num;
	unsigned short i_count;
	unsigned char i_lock;
	unsigned char i_dirt;
	unsigned char i_pipe;
	unsigned char i_mount;
	unsigned char i_seek;
	unsigned char i_update;
};
```

这里看出，每个文件描述指针中有一个指向 inode (i 节点)的指针，i节点的描述如下:  
![image](http://otc9g6h5k.bkt.clouddn.com/5-fs-inode.jpg)

所以从进程到每个文件的描述，就有了这样一张图(apue第三章):  
![image](http://otc9g6h5k.bkt.clouddn.com/fs_inode.png)

图中显示的是 V 节点作为索引部分，i节点作为数据部分，不过linux只用了i节点，有数据部分和索引部分，还有一点，这里的inode只是一个代称，Linux使用ext2/ext3/ext4文件系统，用inode组织磁盘，像ntfs文件系统是不用inode这种形式的，为了支持多个文件系统，Linux实现了 虚拟文件系统

### VFS {#VFS}

计算机中出现的问题，绝大多数都能通过添加中间层的方式实现，这句话真是有道理啊。  
更高版本的Linux内核不断抽象了文件系统，不仅支持磁盘文件，块设备，字符设备，甚至socket也可以看做是一个文件处理，也就是那句经典的“Linux一切皆文件”  
![image](http://otc9g6h5k.bkt.clouddn.com/vfs.png)

高版本内核文件系统引入的 cache 和 支持 socket 挖坑以后再填。

### 引用 {#引用}

- 1. <https://arkingc.github.io/2017/08/18/linux-code-vfs/>
- 1. <http://bean-li.github.io/vfs-inode-dentry/>
- 1. <https://compas.cs.stonybrook.edu/~nhonarmand/courses/fa14/cse506.2/slides/vfs.pdf>
- 1. <http://www.ruanyifeng.com/blog/2011/12/inode.html>
- 1. <https://www.zhihu.com/question/30420304/answer/48348738>

– @Sun May 20 18:04:13 CST 2018
