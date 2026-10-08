---
title: 右值引用和构造
date: '2020-12-18T17:00:39+08:00'
url: /2020/12/18/rvalue_constructor/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### copy constructor/assign & move constructor/assign {#copy-constructor-assign-amp-move-constructor-assign}

```plain
class MClass
{
public:
    int* pIntdata;
    int sz;
public:
    MClass(int size) : sz(size)
    {
        pIntdata = new int[size];
        for(int i=0;i<sz;i++)
        {
            pIntdata[i] = i;
        }
        std::cout<< "constructor|addr:" << this << std::endl;
    }
    
    MClass& operator=(const MClass& h)
    {
        std::cout<< "copy assign  operator= " << std::endl;
        this->sz = h.sz;
        if(  this->pIntdata ){ delete [] this->pIntdata ; this->pIntdata = NULL; } // instance get own ptr, size maybe change, del previous
        
        pIntdata = new int[sz];
        for(int i=0;i<sz;i++)
        {
            pIntdata[i] = h.pIntdata[i];
        }   
        return *this;
    }
    
    MClass(const MClass &h) : sz(h.sz ) // copy constructor must pass its first argument by reference 
    {
        this->sz = h.sz; 
        pIntdata = new int[sz];
        for(int i=0;i<sz;i++)
        {
            pIntdata[i] = h.pIntdata[i];
        }
        std::cout<< "copy constructor " << this << std::endl;
    }
    
    MClass( MClass && h ): sz(h.sz ),pIntdata(h.pIntdata)  // move constructor 
    {
        std::cout<< "move constructor " << this << std::endl;
        h.pIntdata = nullptr;               // 处理关联的资源
    }
    MClass& operator=( MClass&& h)
    {
        std::cout<< "move MClass& operator=(const MClass&&) " << this << std::endl;
        this->sz = h.sz;
        this->pIntdata = h.pIntdata;

        h.pIntdata = nullptr;
    }
    ~MClass()
    {
        if( pIntdata  )
        {
            delete [] pIntdata;
            pIntdata = NULL;
        } 
        std::cout<< "destructor " << this << std::endl;
    }
};

MClass GetOne()
{
    MClass mc = MClass(42);
    std::cout << "GetOne()::&mc " <<  &mc  << std::endl;
    return mc; 
}

int func(MClass&& mc)
{
    std::cout << "int func(MClass&& mc)|&mc:" << &mc << std::endl;
    return 0; 
}

int func2(MClass mc)
{   
    std::cout << "int func1(MClass&& mc)|&mc:" << &mc << std::endl;
    return 0;
}

int main()
{
    MClass one = GetOne();  // -fno-elide-constructors 编译器优化
    std::cout << "main::&one " <<  &one  << std::endl;
    MClass two = func( std::move(one) );

    MClass t3(5);
    t3 = std::move(one);
    return 0;
}
```

### C++ 的移动 move 是怎么运作的？ {#C-的移动-move-是怎么运作的？}

引入移动构造是为了避免多余的堆申请

std::move在运行期不做任何事情（不生成任何机器码），在编译期只做一件事情，就是把入参cast成对应类型的rvalue，从而影响其他函数调用的重载决议。你可以理解成std::move其实应该叫做比如说cast\_to\_rvalue，但是标准委员会认为这个破名字太长不好记。搞清楚这点你就理解了std::move了。

[C++ 的移动 move 是怎么运作的？](https://www.zhihu.com/question/277908001)  
 C++的move基本啥也没干， 逻辑是自己实现的，参考上面的

```plain
MClass( MClass && h ): sz(h.sz ),pIntdata(h.pIntdata)  // move constructor 
{
    std::cout<< "move constructor " << this << std::endl;
    h.pIntdata = nullptr;             
}
```
