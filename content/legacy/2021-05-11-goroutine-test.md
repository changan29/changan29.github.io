---
title: 从一个小例子看go的调度
date: '2021-05-11T19:40:39+08:00'
url: /2021/05/11/goroutine-test/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 从一个小例子看go的调度 {#从一个小例子看go的调度}

### 问题 {#问题}

```plain
func main() {
	runtime.GOMAXPROCS(1)
	ch := make(chan int)
	go count(ch, 10000)
	go count(ch, 10001)
	time.Sleep(10000 * time.Millisecond)
	fmt.Printf("exit\n")
}

func count(r chan int, who int) {
	for {
		if who%2 == 0 {
			r <- who
			fmt.Printf("|write <- who|%d\n", who)
		} else {
			<-r
			fmt.Printf("| <-r recv|%d\n", who)
		}
	}
}
```

```plain
输出是
| <-r recv|10001
| <-r recv|10001
|write <- who|10000
|write <- who|10000
......
为什么不是一个一个交替的形式? 
| <-r recv|10001
|write <- who|10000
| <-r recv|10001
|write <- who|10000
......
```

### 解释 {#解释}

- 首先,runtime.GOMAXPROCS(1) 这种情况下两个 goroutine 都靠同一个 chan 形成同步，是完全稳定的输出

#### 一句话解释 {#一句话解释}

因为是非抢占式调度。阻塞会立刻转移数据但不会提前取走调度权。每次某个 goroutine 先运行一轮满足另一 goroutine 的阻塞，再运行一轮到自己再阻塞，这才能顺利移交调度权。

#### 长解释 {#长解释}

数字太长了我就写 0 和 1 了。先起了两个 go routine：0 和 1 但不会运行，因为 main() 还没交出调度权，直到 time.Sleep 才交出调度权。  
因为某种实现细节，playground 先执行了 1 （竟然是先入后出？），运行到 <- r，阻塞，交出调度权。  
这时候 0 开始运行，到 r<-who 塞进去了以后这个数据立刻被 1 拿走，但这时候 0 还没交出调度权，继续运行，  
运行到第二次 r<-who 的时候阻塞，交出调度权。  
这时候发现 1 可以运行了，就运行 1，先把上次阻塞的 <- r 处理了，运行一遍，再拿掉 0 阻塞着的那个 who，再运行一遍再阻塞。

阻塞的读取方可以立即读到写入通道的数据（但不会同时移交调度给读取方），所以每一轮其实可以发送两个数据，一个直接给阻塞的读取方读了，一个阻塞在通道，如果通道缓存为 1，那么就是一轮三个，以此类推。

要在通道后进行手动移交调度，才能达到你的想法。
