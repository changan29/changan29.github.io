---
title: 使用 C++11 编写 Linux 多线程程序
date: '2022-02-18T22:40:39+08:00'
url: /2022/02/18/muiti-thread-by-cpp/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 前言 {#前言}

在这个多核时代，如何充分利用每个 CPU 内核是一个绕不开的话题，从需要为成千上万的用户同时提供服务的服务端应用程序，到需要同时打开十几个页面，每个页面都有几十上百个链接的 web 浏览器应用程序，从保持着几 t 甚或几 p 的数据的数据库系统，到手机上的一个有良好用户响应能力的 app，为了充分利用每个 CPU 内核，都会想到是否可以使用多线程技术。这里所说的“充分利用”包含了两个层面的意思，一个是使用到所有的内核，再一个是内核不空闲，不让某个内核长时间处于空闲状态。在 C++98 的时代，C++标准并没有包含多线程的支持，人们只能直接调用操作系统提供的 SDK API 来编写多线程程序，不同的操作系统提供的 SDK API 以及线程控制能力不尽相同，到了 C++11，终于在标准之中加入了正式的多线程的支持，从而我们可以使用标准形式的类来创建与执行线程，也使得我们可以使用标准形式的锁、原子操作、线程本地存储 (TLS) 等来进行复杂的各种模式的多线程编程，而且，C++11 还提供了一些高级概念，比如 promise/future，packaged\_task，async 等以简化某些模式的多线程编程。

### 线程间的数据交互和数据争用 {#线程间的数据交互和数据争用}

同一个进程内的多个线程之间多是免不了要有数据互相来往的，队列和共享数据是实现多个线程之间的数据交互的常用方式，封装好的队列使用起来相对来说不容易出错一些，而共享数据则是最基本的也是较容易出错的，因为它会产生数据争用的情况，即有超过一个线程试图同时抢占某个资源，比如对某块内存进行读写等，如下例所示：

```plain
// thread_data_race.cc
static void
inc(int *p ){
 for(int i = 0; i < COUNT; i++){
 (*p)++;
 }
}
void threadDataRacing(void){
 int a = 0;
 thread ta( inc, &a);
 thread tb( inc, &a);
 ta.join();
 tb.join();
 cout << "a=" << a << endl;
}
```

这是简化了的极端情况，我们可以一眼看出来这是两个线程在同时对&a 这个内存地址进行写操作，但是在实际工作中，在代码的海洋中发现它并不一定容易。从表面看，两个线程执行完之后，最后的 a 值应该是 COUNT  *2，但是实际上并非如此，因为简单如 (*p)++这样的操作并不是一个原子动作，要解决这个问题，对于简单的基本类型数据如字符、整型、指针等，C++提供了原子模版类 atomic，而对于复杂的对象，则提供了最常用的锁机制，比如互斥类 mutex，门锁 lock\_guard，唯一锁 unique\_lock，条件变量 condition\_variable 等。

现在我们使用原子模版类 atomic 改造上述例子得到预期结果：

```plain
// thread_atomic.cc
static void
inc(atomic<int> *p ){
 for(int i = 0; i < COUNT; i++){
 (*p)++;
 }
}
void threadDataRacing(void){
 atomic<int> a(0) ;
 thread ta( inc, &a);
 thread tb( inc, &a);
 ta.join();
 tb.join();
 cout << "a=" << a << endl;
}
```

我们也可以使用 lock\_guard，lock\_guard 是一个范围锁，本质是 RAII(Resource Acquire Is Initialization)，在构建的时候自动加锁，在析构的时候自动解锁，这保证了每一次加锁都会得到解锁。即使是调用函数发生了异常，在清理栈帧的时候也会调用它的析构函数得到解锁，从而保证每次加锁都会解锁，但是我们不能手工调用加锁方法或者解锁方法来进行更加精细的资源占用管理，使用 lock\_guard 示例如下：

```plain
// thread_lock_guard.cc
static mutex g_mutex;
static void
inc(int *p ){
 for(int i = 0; i < COUNT; i++){
 lock_guard<mutex> _(g_mutex);
 (*p)++;
 }
}
void threadLockGuard(void){
 int a = 0;
 thread ta( inc, &a);
 thread tb( inc, &a);
 ta.join();
 tb.join();
 cout << "a=" << a << endl;
}
```

如果要支持手工加锁，可以考虑使用 unique\_lock 或者直接使用 mutex。unique\_lock 也支持 RAII，它也可以一次性将多个锁加锁；如果使用 mutex 则直接调用 mutex 类的 lock, unlock, trylock 等方法进行更加精细的锁管理：

```plain
//thread_mutex.cc
static mutex g_mutex;
static void
inc(int *p ){
 thread_local int i; // TLS 变量
 for(; i < COUNT; i++){
 g_mutex.lock();
 (*p)++;
 g_mutex.unlock();
 }
}
void threadMutex(void){
 int a = 0;
 thread ta( inc, &a);
 thread tb( inc, &a);
 ta.join();
 tb.join();
 cout << "a=" << a << endl;
}
```

在上例中，我们还使用了线程本地存储 (TLS) 变量，我们只需要在变量前面声明它是 thread\_local 即可。TLS 变量在线程栈内分配，线程栈只有在线程创建之后才生效，在线程退出的时候销毁，需要注意不同系统的线程栈的大小是不同的，如果 TLS 变量占用空间比较大，需要注意这个问题。TLS 变量一般不能跨线程，其初始化在调用线程第一次使用这个变量时进行，默认初始化为 0。

对于线程间的事件通知，C++11 提供了条件变量类 condition\_variable，可视为 pthread\_cond\_t 的封装，使用条件变量可以让一个线程等待其它线程的通知 (wait，wait\_for，wait\_until)，也可以给其它线程发送通知 (notify\_one，notify\_all)，条件变量必须和锁配合使用，在等待时因为有解锁和重新加锁，所以，在等待时必须使用可以手工解锁和加锁的锁，比如 unique\_lock，而不能使用 lock\_guard，示例如下：

```plain
//thread_cond_var.cc
#include <thread>
#include <iostream>
#include <condition_variable>
using namespace std;
mutex m;
condition_variable cv;
void threadCondVar(void){
# define THREAD_COUNT 10
 thread** t = new thread*[THREAD_COUNT];
 int i;
 for(i = 0; i < THREAD_COUNT; i++){
 t[i] = new thread( [](int index){
 unique_lock<mutex> lck(m);
 cv.wait_for(lck, chrono::hours(1000));
 cout << index << endl;
 }, i );
 this_thread::sleep_for( chrono::milliseconds(50));
 }
 for(i = 0; i < THREAD_COUNT; i++){
 lock_guard<mutex> _(m);
 cv.notify_one();
 }
 for(i = 0; i < THREAD_COUNT; i++){
 t[i]->join();
 delete t[i];
 }
 delete t;
}
```

从上例的运行结果也可以看到，条件变量是不保证次序的，即首先调用 wait 的不一定首先被唤醒。

### 几个高级概念 {#几个高级概念}

C++11 提供了若干多线程编程的高级概念：promise/future, packaged\_task, async，来简化多线程编程，尤其是线程之间的数据交互比较简单的情况下，让我们可以将注意力更多地放在业务处理上。

promise/future 可以用来在线程之间进行简单的数据交互，而不需要考虑锁的问题，线程 A 将数据保存在一个 promise 变量中，另外一个线程 B 可以通过这个 promise 变量的 get\_future() 获取其值，当线程 A 尚未在 promise 变量中赋值时，线程 B 也可以等待这个 promise 变量的赋值：

```plain
// thread_promise_future.cc
promise<string> val;
static void
threadPromiseFuture(){
 thread ta([](){
 future<string> fu = val.get_future();
 cout << "waiting promise->future" << endl;
 cout << fu.get() << endl;
 });
 thread tb([](){
 this_thread::sleep_for( chrono::milliseconds(100) );
 val.set_value("promise is set");
 });
 ta.join();
 tb.join();
}
```

一个 future 变量只能调用一次 get()，如果需要多次调用 get()，可以使用 shared\_future，通过 promise/future 还可以在线程之间传递异常。

如果将一个 callable 对象和一个 promise 组合，那就是 packaged\_task，它可以进一步简化操作：

```plain
// thread_packaged_task.cc
static mutex g_mutex;
static void
threadPackagedTask(){
 auto run = [=](int index){ 
 {
 lock_guard<mutex> _(g_mutex);
 cout << "tasklet " << index << endl;
 }
 this_thread::sleep_for( chrono::seconds(10) );
 return index * 1000;
 };
 packaged_task<int(int)> pt1(run);
 packaged_task<int(int)> pt2(run);
 thread t1([&](){pt1(2);} );
 thread t2([&](){pt2(3);} );
 int f1 = pt1.get_future().get();
 int f2 = pt2.get_future().get();
 cout << "task result=" << f1 << endl;
 cout << "task result=" << f2 << endl;
 t1.join();
 t2.join();
}
```

我们还可以试图将一个 packaged\_task 和一个线程组合，那就是 async() 函数。使用 async() 函数启动执行代码，返回一个 future 对象来保存代码返回值，不需要我们显式地创建和销毁线程等，而是由 C++11 库的实现决定何时创建和销毁线程，以及创建几个线程等，示例如下：

```plain
thread_async.cc
static long
do_sum(vector<long> *arr, size_t start, size_t count){
 static mutex _m;
 long sum = 0;
 for(size_t i = 0; i < count; i++){
 sum += (*arr)[start + i];
 }
 {
 lock_guard<mutex> _(_m);
 cout << "thread " << this_thread::get_id() 
 << ", count=" << count
 << ", sum=" << sum << endl;
 }
 return sum;
}
static void
threadAsync(){
# define COUNT 1000000
 vector<long> data(COUNT);
 for(size_t i = 0; i < COUNT; i++){
 data[i] = random() & 0xff;
 }
 //
 vector< future<long> > result;
 size_t ptc = thread::hardware_concurrency() * 2;
 for(size_t batch = 0; batch < ptc; batch++){
 size_t batch_each = COUNT / ptc;
 if (batch == ptc - 1){
 batch_each = COUNT - (COUNT / ptc * batch);
 }
 result.push_back(async(do_sum, &data, batch * batch_each, batch_each));
 }
 long total = 0;
 for(size_t batch = 0; batch < ptc; batch++){
 total += result[batch].get();
 }
 cout << "total=" << total << endl;
}
```

如果是在多核或者多 CPU 的环境上面运行上述例子，仔细观察输出结果，可能会发现有些线程 ID 是重复的，这说明重复使用了线程，也就是说，通过使用 async() 还可达到一些线程池的功能。

### 几个需要注意的地方 {#几个需要注意的地方}

thread 同时也是棉线、毛线、丝线等意思，我想大家都能体会面对一团乱麻不知从何处查找头绪的感受，不要忘了，线程不是静态的，它是不断变化的，请想像一下面对一团会动态变化的乱麻的情景。所以，使用多线程技术的首要准则是我们自己要十分清楚我们的线程在哪里？线头（线程入口和出口）在哪里？先安排好线程的运行，注意不同线程的交叉点（访问或者修改同一个资源，包括内存、I/O 设备等），尽量减少线程的交叉点，要知道几条线堆在一起最怕的是互相打结。

当我们的确需要不同线程访问一个共同的资源时，一般都需要进行加锁保护，否则很可能会出现数据不一致的情况，从而出现各种时现时不现的莫名其妙的问题，加锁保护时有几个问题需要特别注意：一是一个线程内连续多次调用非递归锁 (non-recursive lock) 的加锁动作，这很可能会导致异常；二是加锁的粒度；三是出现死锁 (deadlock)，多个线程互相等待对方释放锁导致这些线程全部处于罢工状态。

第一个问题只要根据场景调用合适的锁即可，当我们可能会在某个线程内重复调用某个锁的加锁动作时，我们应该使用递归锁 (recursive lock)，在 C++11 中，可以根据需要来使用 recursive\_mutex，或者 recursive\_timed\_mutex。

第二个问题，即锁的粒度，原则上应该是粒度越小越好，那意味着阻塞的时间越少，效率更高，比如一个数据库，给一个数据行 (data row) 加锁当然比给一个表 (table) 加锁要高效，但是同时复杂度也会越大，越容易出错，比如死锁等。

对于第三个问题我们需要先看下出现死锁的条件：

- 资源互斥，某个资源在某一时刻只能被一个线程持有 (hold)；
- 吃着碗里的还看着锅里的，持有一个以上的互斥资源的线程在等待被其它进程持有的互斥资源；
- 不可抢占，只有在某互斥资源的持有线程释放了该资源之后，其它线程才能去持有该资源；
- 环形等待，有两个或者两个以上的线程各自持有某些互斥资源，并且各自在等待其它线程所持有的互斥资源。

我们只要不让上述四个条件中的任意一个不成立即可。在设计的时候，非常有必要先分析一下会否出现满足四个条件的情况，特别是检查有无试图去同时保持两个或者两个以上的锁，当我们发现试图去同时保持两个或者两个以上的锁的时候，就需要特别警惕了。下面我们来看一个简化了的死锁的例子：

```plain
// thread_deadlock.cc
static mutex g_mutex1, g_mutex2;
static void
inc1(int *p ){
 for(int i = 0; i < COUNT; i++){
 g_mutex1.lock();
 (*p)++;
 g_mutex2.lock();
 // do something.
 g_mutex2.unlock();
 g_mutex1.unlock();
 }
}
static void
inc2(int *p ){
 for(int i = 0; i < COUNT; i++){
 g_mutex2.lock();
 g_mutex1.lock();
 (*p)++;
 g_mutex1.unlock();
 // do other thing.
 g_mutex2.unlock();
 }
}
void threadMutex(void){
 int a = 0;
 thread ta( inc1, &a);
 thread tb( inc2, &a);
 ta.join();
 tb.join();
 cout << "a=" << a << endl;
}
```

在这个例子中，g\_mutex1 和 g\_mutex2 都是互斥的资源，任意时刻都只有一个线程可以持有（加锁成功），而且只有持有线程调用 unlock 释放锁资源的时候其它线程才能去持有，满足条件 1 和 3，线程 ta 持有了 g\_mutex1 之后，在释放 g\_mutex1 之前试图去持有 g\_mutex2，而线程 tb 持有了 g\_mutex2 之后，在释放 g\_mutex2 之前试图去持有 g\_mutex1，满足条件 2 和 4，这种情况之下，当线程 ta 试图去持有 g\_mutex2 的时候，如果 tb 正持有 g\_mutex2 而试图去持有 g\_mutex1 时就发生了死锁。在有些环境下，可能要多次运行这个例子才出现死锁，实际工作中这种偶现特性让查找问题变难。要破除这个死锁，我们只要按如下代码所示破除条件 3 和 4 即可：

```plain
// thread_break_deadlock.cc
static mutex g_mutex1, g_mutex2;
static voi
inc1(int *p ){
 for(int i = 0; i < COUNT; i++){
 g_mutex1.lock();
 (*p)++;
 g_mutex1.unlock();
 g_mutex2.lock();
 // do something.
 g_mutex2.unlock();
 }
}
static void
inc2(int *p ){
 for(int i = 0; i < COUNT; i++){
 g_mutex2.lock();
 // do other thing.
 g_mutex2.unlock();
 g_mutex1.lock();
 (*p)++;
 g_mutex1.unlock();
 }
}
void threadMutex(void){
 int a = 0;
 thread ta( inc1, &a);
 thread tb( inc2, &a);
 ta.join();
 tb.join();
 cout << "a=" << a << endl;
}
```

在一些复杂的并行编程场景，如何避免死锁是一个很重要的话题，在实践中，当我们看到有两个锁嵌套加锁的时候就要特别提高警惕，它极有可能满足了条件 2 或者 4。

#### 编译与调试 {#编译与调试}

- 使用 g++编译链接时设置 -Wl,–no-as-needed 传给链接器，有些版本的 g++需要这个设置；
- 设置宏定义 -D\_REENTRANT，有些库函数是依赖于这个宏定义来确定是否使用多线程版本的。
- 在用 gdb 调试多线程程序的时候，可以输入命令 info threads 查看当前的线程列表，通过命令 thread n 切换到第 n 个线程的上下文，这里的 n 是 info threads 命令输出的线程索引数字，例如，如果要切换到第 2 个线程的上下文，则输入命令 thread 2。
