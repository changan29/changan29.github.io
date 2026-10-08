---
title: std::shared_ptr
date: '2020-12-16T15:40:39+08:00'
url: /2020/12/16/shared_ptr/
draft: false
categories:
- c++
tags:
- c++
---

### shared\_ptr

```plain
template< class T > class shared_ptr; (C++11 起)
```

多个shared\_ptr管理同一个指针，仅当最后一个shared\_ptr析构时，指针才被delete。这是怎么实现的呢？答案是：引用计数（reference counting）。引用计数指的是，所有管理同一个裸指针（raw pointer）的shared\_ptr，都共享一个引用计数器，每当一个shared\_ptr被赋值（或拷贝构造）给其它shared\_ptr时，这个共享的引用计数器就加1，当一个shared\_ptr析构或者被用于管理其它裸指针时，这个引用计数器就减1，如果此时发现引用计数器为0，那么说明它是管理这个指针的最后一个shared\_ptr了，于是我们释放指针指向的资源。

#### shared\_ptr的简单实现

```plain
#include<iostream>
#include<mutex>
#include<thread>
using namespace std;

template<class T>
class Shared_Ptr{
public:
	Shared_Ptr(T* ptr = nullptr)
		:_pPtr(ptr)
		, _pRefCount(new int(1))
		, _pMutex(new mutex)
	{}
	~Shared_Ptr()
	{
		Release();
	}
	Shared_Ptr(const Shared_Ptr<T>& sp)
		:_pPtr(sp._pPtr)
		, _pRefCount(sp._pRefCount)
		, _pMutex(sp._pMutex)
	{
		AddRefCount();
	}
	Shared_Ptr<T>& operator=(const Shared_Ptr<T>& sp)
	{
		//if (this != &sp)
		if (_pPtr != sp._pPtr)
		{
			// 释放管理的旧资源
			Release();
			// 共享管理新对象的资源，并增加引用计数
			_pPtr = sp._pPtr;
			_pRefCount = sp._pRefCount;
			_pMutex = sp._pMutex;
			AddRefCount();
		}
		return *this;
	}
	T& operator*(){
		return *_pPtr;
	}
	T* operator->(){
		return _pPtr;
	}
	int UseCount() { return *_pRefCount; }
	T* Get() { return _pPtr; }
	void AddRefCount()
	{
		_pMutex->lock();
		++(*_pRefCount);
		_pMutex->unlock();
	}
private:
	void Release()
	{
		bool deleteflag = false;
		_pMutex->lock();
		if (--(*_pRefCount) == 0)
		{
			delete _pRefCount;
			delete _pPtr;
			deleteflag = true;
		}
		_pMutex->unlock();
		if (deleteflag == true)
			delete _pMutex;
	}
private:
	int *_pRefCount;
	T* _pPtr;
	mutex* _pMutex;
};
```

### 错误用法

#### 1. 循环引用

```plain
struct ListNode
{
	int _data;
	shared_ptr<ListNode> _prev;
	shared_ptr<ListNode> _next;
	~ListNode(){ cout << "~ListNode()" << endl; }
};
int main()
{
	shared_ptr<ListNode> node1(new ListNode);
	shared_ptr<ListNode> node2(new ListNode);
	cout << node1.use_count() << endl;
	cout << node2.use_count() << endl;
	node1->_next = node2;
	node2->_prev = node1;
	cout << node1.use_count() << endl;
	cout << node2.use_count() << endl;
	system("pause");
	return 0;
}
```

解决方案：在引用计数的场景下，把节点中的\_prev和\_next改成weak\_ptr就可以了

#### 2. 多个无关的shared\_ptr管理同一裸指针

只能通过复制构造或复制赋值其值给另一 shared\_ptr ，将对象所有权与另一 shared\_ptr 共享。用另一 shared\_ptr 所占有的底层指针创建新的 shared\_ptr 导致未定义行为。  

```plain
int *a = new int;
std::shared_ptr<int> p1(a);
std::shared_ptr<int> p2(a);
```

#### 3. 直接用new构造多个shared\_ptr作为实参

```plain
// 声明
void f(A *p1, B *p2);
// 使用
f(new A, new B);       <--  ×
```

```plain
// 声明
void f(shared_ptr<A> p1, shared_ptr<B> p2);
// 使用
f(shared_ptr<A>(new A), shared_ptr<B>(new B));              <-- √
```

#### tips

- 用shared\_ptr，不用new
- 使用weak\_ptr来打破循环引用
- 用make\_shared来生成shared\_ptr

### ref

- <https://heleifz.github.io/14696398760857.html>
- <https://zh.cppreference.com/w/cpp/memory/shared_ptr>

blog: <http://www.oneyearago.me>
