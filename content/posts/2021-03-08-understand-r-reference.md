---
title: 理解右值引用
date: '2021-03-08T21:40:39+08:00'
url: /2021/03/08/understand-r-reference/
draft: false
categories:
- C++
tags:
- C++
- C++11
---

#### 一句话概述

std::move本身只做类型转换，对性能无影响。  
我们可以在自己的类中实现移动语义，避免深拷贝，充分利用右值引用和std::move的语言特性。

移动语义目的就是用浅拷贝代替深拷贝，右值引用跟深拷贝放到同一场景才是有意义的。

#### 实现移动语义

在没有右值引用之前，一个简单的数组类通常实现如下，有构造函数、拷贝构造函数、赋值运算符重载、析构函数等。深拷贝/浅拷贝在此不做讲解

```plain
class Array {
public:
    Array(int size) : size_(size) {
        data = new int[size_];
    }
     
    // 深拷贝构造
    Array(const Array& temp_array) {
        size_ = temp_array.size_;
        data_ = new int[size_];
        for (int i = 0; i < size_; i ++) {
            data_[i] = temp_array.data_[i];
        }
    }
     
    // 深拷贝赋值
    Array& operator=(const Array& temp_array) {
        delete[] data_;
 
        size_ = temp_array.size_;
        data_ = new int[size_];
        for (int i = 0; i < size_; i ++) {
            data_[i] = temp_array.data_[i];
        }
    }
 
    ~Array() {
        delete[] data_;
    }
 
public:
    int *data_;
    int size_;
};
```

该类的拷贝构造函数、赋值运算符重载函数已经通过使用左值引用传参来避免一次多余拷贝了，但是内部实现要深拷贝，无法避免。

这时，有人提出一个想法：是不是可以提供一个移动构造函数，把被拷贝者的数据移动过来，被拷贝者后边就不要了，这样就可以避免深拷贝了，如：

```plain
class Array {
public:
    Array(int size) : size_(size) {
        data = new int[size_];
    }
     
    // 深拷贝构造
    Array(const Array& temp_array) {
        ...
    }
     
    // 深拷贝赋值
    Array& operator=(const Array& temp_array) {
        ...
    }
 
    // 移动构造函数，可以浅拷贝
    Array(const Array& temp_array, bool move) {
        data_ = temp_array.data_;
        size_ = temp_array.size_;
        // 为防止temp_array析构时delete data，提前置空其data_      
        temp_array.data_ = nullptr;
    }
     
 
    ~Array() {
        delete [] data_;
    }
 
public:
    int *data_;
    int size_;
};
```

这么做有2个问题：

- 不优雅，表示移动语义还需要一个额外的参数(或者其他方式)。
- 无法实现！temp\_array是个const左值引用，无法被修改，所以temp\_array.data\_ = nullptr;这行会编译不过。当然函数参数可以改成非const：Array(Array& temp\_array, bool move){…}，这样也有问题，由于左值引用不能接右值，Array a = Array(Array(), true);这种调用方式就没法用了。

可以发现左值引用真是用的很不爽，右值引用的出现解决了这个问题，在STL的很多容器中，都实现了以右值引用为参数的移动构造函数和移动赋值重载函数，或者其他函数，最常见的如std::vector的push\_back和emplace\_back。参数为左值引用意味着拷贝，为右值引用意味着移动。

```plain
class Array {
public:
    ......
 
    // 优雅
    Array(Array&& temp_array) {
        data_ = temp_array.data_;
        size_ = temp_array.size_;
        // 为防止temp_array析构时delete data，提前置空其data_      
        temp_array.data_ = nullptr;
    }
     
 
public:
    int *data_;
    int size_;
};
```

#### 实例：vector::push\_back使用std::move提高性能

```plain
// 例2：std::vector和std::string的实际例子
int main() {
    std::string str1 = "aacasxs";
    std::vector<std::string> vec;
     
    vec.push_back(str1); // 传统方法，copy
    vec.push_back(std::move(str1)); // 调用移动语义的push_back方法，避免拷贝，str1会失去原有值，变成空字符串
    vec.emplace_back(std::move(str1)); // emplace_back效果相同，str1会失去原有值
    vec.emplace_back("axcsddcas"); // 当然可以直接接右值
}
 
// std::vector方法定义
void push_back (const value_type& val);
void push_back (value_type&& val);
 
void emplace_back (Args&&... args);
```

在vector和string这个场景，加个std::move会调用到移动语义函数，避免了深拷贝。

#### ref

- KM
- <https://zh.cppreference.com/w/cpp/language/reference>
