---
title: c++虚函数表
date: '2020-12-22T15:40:39+08:00'
url: /2020/12/22/cpp-class-model/
draft: false
categories:
- c++
tags:
- c++
---

### 概述 {#概述}

首先,相较于C语言，C++语言并没有额外增加内存消耗(确切说，在没有虚函数情况下)。 对于一个C++类对象,每个对象有独立的数据成员(非static),但是内存中成员函数只有一份,该类的所有对象共享成员函数。

编译器在编译阶段,进行函数的重构，即将成员函数进行非成员化。通过将this指针作为函数的第一个参数,通过this指针即可以找到对象的数据成员

### 使用GDB调试 C++ 虚函数 {#使用GDB调试-C-虚函数}

```plain
class Base
{
public:
        int a;
        int b;
        virtual void function(){}
        virtual void function001(){}

        Base(){}
        void NotVirtFunc(){}
};

class Derive : public Base
{
public:
    Derive(){}
    int c;
    virtual void function(){}
    virtual void function006(){}
};

class Test
{
    public:
    int vv;
    virtual void bar(){}
    virtual void foo(){}

    Test(){}
};

class Derive2 : public Derive, public Test
{
public:
    Derive2(){}
    int d2;
    virtual void function(){}
    virtual void function006(){}
    virtual void bar(){}
};

int main()
{
    Base b;                         // 没有继承
    b.function();

    Derive d;                       // 继承
    d.function();

    Derive2* d2 = new Derive2();    //多继承
    d2->function();
    delete d2;
    return 0;
}
```

```plain
(gdb) p b
$1 = {_vptr.Base = 0x400a90 <vtable for Base+16>, a = 0, b = 0}
(gdb) info vtbl b
vtable for 'Base' @ 0x400a90 (subobject @ 0x7fffffffe2e0):
[0]: 0x40081e <Base::function()>
[1]: 0x400828 <Base::function001()>
```

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/post-cpp-no-inherit.png)

```plain
(gdb) p d
$2 = {<Base> = {_vptr.Base = 0x400a50 <vtable for Derive+16>, a = 0, b = 0}, c = 1}
(gdb) info vtbl d
vtable for 'Derive' @ 0x400a50 (subobject @ 0x7fffffffe2c0):
[0]: 0x40086e <Derive::function()>
[1]: 0x400828 <Base::function001()>
[2]: 0x400878 <Derive::function006()>
```

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/post-cpp-single-inherit.png)

```plain
(gdb) p *d2        
$5 = {<Derive> = {<Base> = {_vptr.Base = 0x4009d0 <vtable for Derive2+16>, a = 0, b = 0}, c = 0}, 
        <Test> = {_vptr.Test = 0x400a00 <vtable for Derive2+64>, vv = 0}, d2 = 0}
(gdb) info vtbl d2
vtable for 'Derive2' @ 0x4009d0 (subobject @ 0x603010):
[0]: 0x4008ee <Derive2::function()>
[1]: 0x400828 <Base::function001()>
[2]: 0x4008f8 <Derive2::function006()>
[3]: 0x400902 <Derive2::bar()>

vtable for 'Test' @ 0x400a00 (subobject @ 0x603028):
[0]: 0x40090c <non-virtual thunk to Derive2::bar()>
[1]: 0x40088c <Test::foo()>
```

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/post-cpp-multi-inherit.png)

#### 构造函数与虚函数表 {#构造函数与虚函数表}

虚函数表创建时机是在编译期间。 编译期间编译器就为每个类确定好了对应的虚函数表里的内容。 所以在程序运行时，编译器会把虚函数表的首地址赋值给虚函数表指针，所以，这个虚函数表指针就有值了。

![image](https://blog2020-1251969284.cos.ap-shanghai.myqcloud.com/intel/c%2B%2B_class_model.png)

### ref {#ref}

- <https://tangocc.github.io/2018/03/20/cpp-class-memory-struct/>

### TODO {#TODO}

菱形继承于虚继承这里没写，使用gdb也可以很快找到，编译器的规则而已，同理，后面这两种知道规则就可以了，在语言的框架下理解就可以了。
