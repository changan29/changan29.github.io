---
title: tcp socket api
date: '2019-12-11T00:40:39+08:00'
url: /2019/12/11/socket_api/
draft: false
categories:
- socket
tags:
- socket
---

### socket是网络协议栈的抽象

### socket api在tcp中的体现

![image](https://201911-1251969284.cos.ap-shanghai.myqcloud.com/tcp-sync-queue-and-accept-queue.jpg)

TCP三次握手后有个accept队列，进到这个队列才能从Listen变成accept，默认backlog 值是50。

如上图所示，这里有两个队列：syns queue(半连接队列）；accept queue（全连接队列）  
三次握手中，在第一步server收到client的syn后，把相关信息放到半连接队列中，同时回复syn+ack给client，第三步的时候server收到client的ack，如果这时全连接队列没满，那么从半连接队列拿出相关信息放入到全连接队列中。

当接收到3次握手中的ACK分组时，将它们移动到accept队列。 显而易见，accept系统调用只是简单地从完成队列中取出连接。 在这种情况下，listen syscall的backlog参数表示完成队列的大小。

### code

```plain
/*
tcpechosvr.c
*/
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <errno.h>
#include <string.h>

#define SRVIP "127.0.0.1"
#define SRVPORT 10005
#define MAX_NUM 1024

int main()
{
   int serverSock = socket(AF_INET,SOCK_STREAM,IPPROTO_TCP);
    if(serverSock < 0)
    {
        printf("socket creation failed\n");
        exit(-1);
    }
    printf("socket create successfully.\n");

    struct sockaddr_in serverAddr;
    memset(&serverAddr,0,sizeof(serverAddr));
    serverAddr.sin_family=AF_INET;
    serverAddr.sin_port = htons((u_short)SRVPORT);
    serverAddr.sin_addr.s_addr = inet_addr(SRVIP);

    if(bind(serverSock, (struct sockaddr*)&serverAddr, sizeof(struct sockaddr))==-1)
    {
        printf("Bind error.IP[%s], Port[%d]\n", SRVIP, serverAddr.sin_port);
        exit(-1);
    }
    printf("Bind successful.IP[%s], Port[%d]\n", SRVIP, serverAddr.sin_port);

        /*  把sock链接进入监听哈希表listening_hash & 创建半连接队列的实例，初始化全连接队列 */
    if(listen(serverSock,10)==-1)
    {
        printf("Listen error!\n");
    }
    printf("Listening on port[%d]\n", serverAddr.sin_port);

    char recvBuf[MAX_NUM]={0};
    while(1)
    {
        struct sockaddr clientAddr;
        int size = sizeof(clientAddr);

        /* 从全连接队列中取  */
        int clientSock = accept(serverSock, &clientAddr, &size);
        printf("***SYS***    New client touched.\n");
        while(1)
        {
            //一直接收客户端socket的send操作
            if(recv(clientSock, recvBuf, MAX_NUM, 0) == -1)
            {
                printf("read error.\n");
            }
            else
            {
                printf("receiv from client:%s\n",recvBuf);
            }

            // 向客户端发送套接字
            if(send(clientSock, recvBuf , sizeof(recvBuf), MSG_NOSIGNAL) == -1)
            {
                printf("Send error!\n");
            }
            else
            {
                printf("Send to client:%s\n", recvBuf);
            }
            if( strcmp(recvBuf, "bye") == 0)
            {
                break;
            }
            memset(recvBuf, 0, sizeof(recvBuf));
        }
        close(clientSock);
    }
    close(serverSock);
    return 0;
}
```

### REF

- [关于TCP 半连接队列和全连接队列](http://jm.taobao.org/2017/05/25/525-1/)
- [How TCP backlog works in Linux](http://veithen.io/2014/01/01/how-tcp-backlog-works-in-linux.html)
