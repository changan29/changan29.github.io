---
title: 进程通信 & PIPE
date: '2019-10-24T23:58:45+08:00'
url: /2019/10/24/IPC_and_pipe/
draft: false
categories:
- IPC
- pipe
tags:
- IPC
- pipe
---

### 历史

UNIX两大贡献者贝尔实验室和BSD，在进程之间通信侧重不同，前者基于内核对进程之间的通信手段进行了改进，形成了“System V IPC”，而后者则是基于网络形成了套接字。

而POSIX则是IEEE制定的标准，目的是为运行在不同操作系统上的软件提供统一的接口，实现者则是不同的操作系统内核开发人员。

System V 以及POSIX 对信号量、共享内存、消息队列等进程之间共享方式提供了自己的解决方案。

在观察使用进程间通信手段后，会发现在多线程中使用的基本是POSIX标准提供的接口函数，而多进程则是基于System V。

### 管道

pipe和fifo用的不多了，让我们从Nginx那里学一个 全双工的管道: socketpair  
![image](https://201910-1251969284.cos.ap-shanghai.myqcloud.com/socket_pair_pipe.bmp)

```plain
#include <iostream>
#include <unistd.h>
#include <sys/uio.h>
#include <sys/socket.h>
#include <stdlib.h>
#include <stdio.h>

ssize_t
sock_fd_write(int sock, void *buf, ssize_t buflen, int fd)
{
  ssize_t     size;
  struct msghdr   msg;
  struct iovec    iov;
  union {
    struct cmsghdr  cmsghdr;
    char        control[CMSG_SPACE(sizeof (int))];
  } cmsgu;
  struct cmsghdr  *cmsg;

  iov.iov_base = buf;
  iov.iov_len = buflen;

  msg.msg_name = NULL;
  msg.msg_namelen = 0;
  msg.msg_iov = &iov;
  msg.msg_iovlen = 1;

  if (fd != -1) {
    msg.msg_control = cmsgu.control;
    msg.msg_controllen = sizeof(cmsgu.control);

    cmsg = CMSG_FIRSTHDR(&msg);
    cmsg->cmsg_len = CMSG_LEN(sizeof (int));
    cmsg->cmsg_level = SOL_SOCKET;
    cmsg->cmsg_type = SCM_RIGHTS;

    printf ("passing fd %d\n", fd);
    *((int *) CMSG_DATA(cmsg)) = fd;
  } else {
    msg.msg_control = NULL;
    msg.msg_controllen = 0;
    printf ("not passing fd\n");
  }

  size = sendmsg(sock, &msg, 0);

  if (size < 0)
    perror ("sendmsg");
  return size;
}

ssize_t
sock_fd_read(int sock, void *buf, ssize_t bufsize, int *fd)
{
  ssize_t     size;

  if (fd) {
    struct msghdr   msg;
    struct iovec    iov;
    union {
      struct cmsghdr  cmsghdr;
      char        control[CMSG_SPACE(sizeof (int))];
    } cmsgu;
    struct cmsghdr  *cmsg;

    iov.iov_base = buf;
    iov.iov_len = bufsize;

    msg.msg_name = NULL;
    msg.msg_namelen = 0;
    msg.msg_iov = &iov;
    msg.msg_iovlen = 1;
    msg.msg_control = cmsgu.control;
    msg.msg_controllen = sizeof(cmsgu.control);
    size = recvmsg (sock, &msg, 0);
    if (size < 0) {
      perror ("recvmsg");
      exit(1);
    }
    cmsg = CMSG_FIRSTHDR(&msg);
    if (cmsg && cmsg->cmsg_len == CMSG_LEN(sizeof(int))) {
      if (cmsg->cmsg_level != SOL_SOCKET) {
        fprintf (stderr, "invalid cmsg_level %d\n",
                 cmsg->cmsg_level);
        exit(1);
      }
      if (cmsg->cmsg_type != SCM_RIGHTS) {
        fprintf (stderr, "invalid cmsg_type %d\n",
                 cmsg->cmsg_type);
        exit(1);
      }

      *fd = *((int *) CMSG_DATA(cmsg));
      printf ("received fd %d\n", *fd);
    } else
      *fd = -1;
  } else {
    size = read (sock, buf, bufsize);
    if (size < 0) {
      perror("read");
      exit(1);
    }
  }
  return size;
}

void
child(int sock)
{
  int fd;
  char    buf[16];
  ssize_t size;

  sleep(1);
  for (;;) {
    size = sock_fd_read(sock, buf, sizeof(buf), &fd);
    if (size <= 0)
      break;
    printf ("read %d\n", (int) size);
    if (fd != -1) {
      write(fd, "hello, world\n", 13);
      close(fd);
    }
  }
}

void
parent(int sock)
{
  ssize_t size;
  int i;
  int fd;

  fd = 1;
  size = sock_fd_write(sock, (void *) "1", 1, 1);
  printf ("wrote %d\n", (int) size);
}

int
main(int argc, char **argv)
{
  int sv[2];
  int pid;

  if (socketpair(AF_UNIX, SOCK_DGRAM, 0, sv) < 0) {
    perror("socketpair");
    exit(1);
  }
  switch ((pid = fork())) {
  case 0:
    close(sv[0]);
    child(sv[1]);
    break;
  case -1:
    perror("fork");
    exit(1);
  default:
    close(sv[1]);
    parent(sv[0]);
    break;
  }
  return 0;
}
```
