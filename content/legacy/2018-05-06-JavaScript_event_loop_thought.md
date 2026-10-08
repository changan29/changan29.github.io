---
title: JavaScript事件随想
date: '2018-05-06T01:30:39+08:00'
url: /2018/05/06/JavaScript_event_loop_thought/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 问题起因 {#问题起因}

最近想学习下web前端，原因是这样的，当前中国(公元2018年)的网络环境，表现形式基本是web+移动端，像微信的小程序只要写js应该就可以了，移动端的app很多也是以web的形式展示，应用程序内是html+css+js，不仅方便发布，而且做起来也快，写完页面样式，更新资源，就好了，类似热更

### 问题 {#问题}

javascript是怎么捕获到鼠标点击事件的？

js有很多事件，鼠标的、页面加载完成的、点击按钮的；etc..

### 理清这个思路 {#理清这个思路}

以我 windows环境下 chrome 浏览器为例

##### 首先几个概念 {#首先几个概念}

- 鼠标这种硬件产生的信号肯定是操作系统处理，操作系统转发给对应的进程
- 在windows中，也有鼠标按下,moveover 这类的消息
- 在windows中,一个消息，是系统定义的一个32位的值，他唯一的定义了一个事件，向 Windows发出一个通知，告诉应用程序某个事情发生了。例如，单击鼠标、改变窗口尺寸、按下键盘上的一个键都会使Windows发送一个消息给应用程序的消息队列,然后应用程序再从消息队列中取出消息并进行相应的响应
- Windows操作系统也会给应用程序“发送消息”，而所谓的发送消息——–实际上就是操作系统调用程序中的一个专门负责处理消息的函数，这个函数称为窗口过程(应用程序给出的一个函数)
- 消息队列可以分成系统消息队列和线程消息队列。系统消息队列由Windows维护，线程消息队列则由每个GUI线程自己进行维护
- 对于队列消息，最常见的是鼠标和键盘触发的消息，例如WM\_MOUSERMOVE,WM\_CHAR等消息，还有一些其它的消息，例如：WM\_PAINT、 WM\_TIMER和WM\_QUIT。当鼠标、键盘事件被触发后，相应的鼠标或键盘驱动程序就会把这些事件转换成相应的消息，然后输送到系统消息队列，由 Windows系统去进行处理。Windows系统则在适当的时机，从系统消息队列中取出一个消息，根据前面我们所说的MSG消息结构确定消息是要被送往那个窗口，然后把取出的消息送往创建窗口的线程的相应队列，下面的事情就该由线程消息队列操心了，Windows开始忙自己的事情去了。线程看到自己的消息队列中有消息，就从队列中取出来，通过操作系统发送到合适的窗口过程去处理。

![image](http://otc9g6h5k.bkt.clouddn.com/Windows_event.jpg)  
![image](http://otc9g6h5k.bkt.clouddn.com/Windows_event2.gif)

##### windows 程序 最基本的窗口 {#windows-程序-最基本的窗口}

```plain
//一个简单的Win32应用程序
//通过这个简单的实例讲解Windows消息是如何传递的
#include <windows.h>
//声明窗口过程函数
LRESULT CALLBACK WndProc(HWND,UINT,WPARAM,LPARAM);
//定义一个全局变量，作为窗口类名
TCHAR szClassName[] = TEXT("SimpleWin32");
//应用程序主函数
int WINAPI WinMain (HINSTANCE hInstance,
                                  HINSTANCE hPrevInstance,
                                  LPSTR szCmdLine,
                                  int iCmdShow)
{
    //****1.设计一个窗口类****
    WNDCLASS wndclass;
    //****2.注册窗口类
    if(!RegisterClass (&wndclass))
    {
           MessageBox (NULL, TEXT ("RegisterClass Fail!"), 
                   szClassName, MB_ICONERROR);
           return 0;
    }
    //****3.创建一个窗口
    HWND hwnd;
    hwnd = CreateWindow(szClassName,//窗口类名称
           TEXT ("The Simple Win32 Application"),//窗口标题 
           WS_OVERLAPPEDWINDOW,//窗口风格,即通常我们使用的windows窗口样式
           CW_USEDEFAULT,//指定窗口的初始水平位置,即屏幕坐标系的窗口的左上角的X坐标
           CW_USEDEFAULT,//指定窗口的初始垂直位置,即屏幕坐标系的窗口的左上角的Y坐标
           CW_USEDEFAULT,//窗口的宽度
           CW_USEDEFAULT,//窗口的高度
           NULL,//父窗口句柄
           NULL,//窗口菜单句柄
           hInstance,//实例句柄
           NULL);
    //****4.显示窗口
    ShowWindow(hwnd,iCmdShow);
    //消息循环
    MSG msg;
    while(GetMessage(&msg,NULL,0,0))//从消息队列中取消息 
    {
           TranslateMessage (&msg);              //转换消息
           DispatchMessage (&msg);               //派发消息
    }
    return msg.wParam;
}

//消息处理函数
//参数:窗口句柄，消息，消息参数，消息参数
LRESULT CALLBACK WndProc(HWND hwnd, UINT message, WPARAM wParam, LPARAM lParam)
{
    //处理感兴趣的消息
    switch (message)
    {
    case WM_DESTROY:
           //当用户关闭窗口，窗口销毁，程序需结束，发退出消息，以退出消息循环
           PostQuitMessage(0);
           return 0;
    // 鼠标消息
    }
    //其他消息交给由系统提供的缺省处理函数
    return ::DefWindowProc (hwnd, message, wParam, lParam);
}
```

##### 浏览器处理消息 {#浏览器处理消息}

在Windows上，浏览器作为一个Windows上的应用程序，也会收到鼠标点击的事件，所以收到以后的事件不是 JavaScript 事件，而是浏览器进程和js解释器共同处理的事件

浏览器访问网页是个网络io,渲染html，css 以及执行 js的操作，有些操作比如说获取远程数据、I/O操作等，他们都很耗时，如果采用同步的方式，那么进程在执行这些操作时就会因为耗时而等待，就像上面那样，下面的任务也只能等待，这样效率并不高。那浏览器是怎么做的呢?

**为了协调事件，用户交互，脚本，渲染，网络等，用户代理必须使用事件循环。**  
**事件循环的主要机制就是任务队列机制:**

- 一个事件循环有一个或者多个任务队列（task queues）。任务队列是task的有序列表，task是调度Events，Parsing，Callbacks，Using a resource，Reacting to DOM manipulation这些任务的算法;
- 每个任务都来自一个特定的任务源（task source）(比如鼠标键盘事件)。来自同一个特定任务源且属于特定事件循环的任务必须被加入到同一个任务队列中，来自不同任务源的任务可以放在不同的任务队列中;
- 浏览器调用这些队列中的任务时采取这样的做法: 相同队列中的任务按照先进先出的顺序, 不同的队列按照提前设置的队列优先级来调用. 例如，用户代理可以有一个用于鼠标和键盘事件的任务队列（用户交互任务源），另一个用于其他任务。然后，用户代理75%概率调用键盘和鼠标事件任务队列，25%调用其他队列, 这样的话就保持界面响应而且不会饿死其他任务队列. 但是相同队列中的任务要按照先进先出的顺序。也就是说单独的任务队列中的任务总是按先进先出的顺序执行，但是不保证多个任务队列中的任务优先级，具体实现可能会交叉执行

在调用任务的过程中, 会产生新的任务, 浏览器就会不断执行任务, 因此称为事件循环.

![image](http://otc9g6h5k.bkt.clouddn.com/browser-event.png)

##### 参考 {#参考}

- [1] <https://segmentfault.com/a/1190000012748907>
- [2] <https://github.com/JaffarL/weeiry/blob/master/2018-2-11/Nodejs%E7%9A%84%E4%BA%8B%E4%BB%B6%E5%BE%AA%E7%8E%AF.md>
- [3] <https://blog.csdn.net/liulianglin/article/details/14449577>
