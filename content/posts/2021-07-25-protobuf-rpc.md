---
title: 基于protobuf 的 rpc
date: '2021-07-25T21:40:39+08:00'
url: /2021/07/25/protobuf-rpc/
draft: false
categories:
- protobuf
- rpc
tags:
- protobuf
- rpc
---

### 基于protobuf 的 rpc {#基于protobuf-的-rpc}

protobuf 实现了序列化部分，预留了 RPC 接口，但是没有实现网络交互的部分。  
基于pb里面的 service 接口，自己实现实际的通信过程，实现一个简易的 rpc是比较容易的， 对我们阅读 brpc、muduo、grpc等著名开源的rpc有很大帮助。

google的 文档这里描述的也比较清楚， 在 google/protobuf/service.h 里，通过一个简单的例子，描述了实现 一个基于pb的RPC的过程。

#### protoc 自动生成的代码 {#protoc-自动生成的代码}

> When you use the protocol compiler to compile a service definition, it generates two classes: An abstract interface for the service (with methods matching the service definition) and a “stub” implementation.  
> A stub is just a type-safe wrapper around an RpcChannel which emulates a local implementation of the service.

##### 例子 {#例子}

```plain
service MyService {
  rpc Foo(MyRequest) returns(MyResponse);
}
```

会生成两个接口: “MyService” and class “MyService\_Stub”,分别对应服务器和客户端接口。

##### 服务器 {#服务器}

对于每个service生成的类，你需要做的是 继承他并重写自己的方法，比如上面的 MyService，你可以这样实现你自己的处理类：

```plain
// server 的处理
// 定义自己的处理类（函数），继承生成的类
class MyServiceImpl : public MyService {
 public:
  MyServiceImpl() {}
  ~MyServiceImpl() {}
  // 实现 MyService::Foo
  void Foo(google::protobuf::RpcController* controller,
           const MyRequest* request,
           MyResponse* response,
           Closure* done) {
    // ... read request and fill in response ...
    done->Run();
  }
};
```

##### 客户端 {#客户端}

要调用远程 MyServiceImpl，首先需要一个 RpcChannel 连接。如何构造通道同样取决于您的 RPC 实现。  
这里我们以一个假设的“MyRpcChannel”为例：

```plain
MyRpcChannel channel("rpc:hostname:1234/myservice");
MyRpcController controller;
MyServiceImpl::Stub stub(&channel);
FooRequest request;
FooRespnose response;
// ... fill in request ...
stub.Foo(&controller, request, &response, NewCallback(HandleResponse));
```

### 几个基类 {#几个基类}

- Service
- RpcController
- RpcChannel

RpcController是处理失败连接以及错误信息用的，对于具体通信，意义不是特别大，我们来看 Service 和 RpcChannel

```plain
class  Service {
 public:
  inline Service() {}
  virtual ~Service();

  virtual const ServiceDescriptor* GetDescriptor() = 0;

  virtual void CallMethod(const MethodDescriptor* method,
                          RpcController* controller,
                          const Message* request,
                          Message* response,
                          Closure* done) = 0;

  virtual const Message& GetRequestPrototype(
    const MethodDescriptor* method) const = 0;
  virtual const Message& GetResponsePrototype(
    const MethodDescriptor* method) const = 0;

};
```

```plain
class  RpcChannel {
 public:
  inline RpcChannel() {}
  virtual ~RpcChannel();

  virtual void CallMethod(const MethodDescriptor* method,
                          RpcController* controller,
                          const Message* request,
                          Message* response,
                          Closure* done) = 0;

};
```

我们看到 Service类和 RpcChannel类都有一个函数叫 CallMethod， 参数列表也一样，可以这样考虑，RpcChannel是一个通道，可以是socket或者其他方式，你去继承这个类实现 RpcChannel， 这样就可以通过 RpcChannel 发送数据到对端， 这里你是发送方（客户端）； 当你收到消息，解包，按照客户端传过来的 method 调用指定方法， 这里就是Service::CallMethod 做的。

所以除了 Message 包以外，还需要协议包头去描述 method， 也就是常用的 cmd\_id一类的。

这样，服务器处理请求包的时候，就可以按照 method 去路由，比如：

```plain
const MethodDescriptor* method = service->GetDescriptor()->FindMethodByName("Foo");
Message* request  = stub->GetRequestPrototype (method)->New();
Message* response = stub->GetResponsePrototype(method)->New();
request->ParseFromString(input);
service->CallMethod(method, *request, response, callback);
```

[protobuf-rpc-demo](https://github.com/changan29/playcpp/tree/master/protobuf-rpc-demo)

### 总结 {#总结}

基于 protobuf 实现一个 rpc,需要关注的点：

- 客户端要实现RpcChannel::CallMethod接口，相当于客户端实现 具体通信过程
- 服务端要实现MyService 中定义的Foo， 实现服务器处理逻辑
- 框架实现 Service注册和Method区分

其实除了序列化和网络通信外，RPC框架基本会包括服务发现、高并发线程库、运维工具等，从一个简易的rpc可以先理解rpc的数据流全貌，有助于学习其他rpc框架。
