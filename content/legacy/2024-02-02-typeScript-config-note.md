---
title: TypeScript配置笔记
date: '2024-02-02T18:20:39+08:00'
url: /2024/02/02/typeScript-config-note/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 创建 {#创建}

```typescript
npx create-react-app my-app-ts --template typescript
```

### 项目文件 {#项目文件}

从项目整体上来说，TS 与原生 JS 的项目架构区别不大，同样也是有`public`文件夹用来存放打包后的文件，同样也有一个`package.json`文件用来说明项目的基本信息以及项目依赖，同样也有一个`src`文件夹用来存放所有的源代码，只不过在 TS 项目中，文件的后缀需要使用`ts`来对应`js`，或者是使用`tsx`来对应`jsx`。

### 基础知识 {#基础知识}

首先，TypeScript 是JavaScript 的超集，可以给原生是JavaScript添加静态类型检查，但是ts和es6所面临的问题一样，就是目前还无法被主流浏览器直接读取。所以如果我们的网站使用typescript或es6编写代码，那么网站上线的第一步就是要把所有的代码转换为浏览器可以识别的代码，也就是基于es5标准下的原生JavaScript。而代码转换这个步骤就叫做编译，compile。

编译typescript代码必须得使用相应的编译器，就好像c++的编译器是gcc，java的编译器是javac一个道理。typescript也有很多不同的编译器可以选择，最常用的是ts-loader、awesome-typescript-loader以及babel-loader。但是，在我们的react项目中，因为使用了create-react-app脚手架工具，所以默认的typescript编译器是babel-loader。而这个babel-loade编译器的配置文件则是这个tsconfig.json文件。

### tsconfig.json {#tsconfig-json}

从文件结构上来说，tsconfig就是一个普通的json格式的数据，最重要的属性是compilerOptions，顾名思义，这个属性定义了编译器的工作方式。

首先，我们先给compilerOptions的最后，加上一个属性，”noImplicitAny”: false。加上这句话以后，我们在写ts代码的时候就不需要显式地声明变量的类型了。  
比如说，我们打开app.js， 我们给app函数加一个参数，a，在没有声明类型的情况下，这个变量a的类型被自动定义为any了。但是，如果我们回到tsconfig中，把noImplicitAny删掉，那么这时候函数app报错了，编译器强制要求我们必须得给参数a加上类型，类型为any。

其实这是一个小问题，但是加上noImplicitAny的好处是什么呢？最大的好处就是在开启noImplicitAny以后，ts代码在不做类型要求的时候与js代码几乎一致，这样我们就可以实现ts与js的混合编写了

#### tsconfig中各种不同的配置属性。 {#tsconfig中各种不同的配置属性。}

##### 第一个，”target”: “es5”, {#第一个，”target”-“es5”}

这个属性定义了你编译后的目标javascript版本，一般来说，我们需要让他编译为es5，这样就可以被主流浏览器解读了。当然，你说我的react代码不是给浏览器看的，比如说，你使用的是react-native做作手机app，那么这里的选项可以选择es6。除了es5和es6，我们还有很多其他常见的选项，ES5, ES6/ES2015, ES2016, ES2017, ES2018, ES2019, ES2020, ESNext，等等等等

##### 第二个， lib 属性， {#第二个，-lib-属性，}

这个属性列出了编译期间需要被包括进来的库文件，通过这些库文件，告诉typescript编译器可以使用哪些功能。比如说，我们这里有一个dom的库文件，这个文件会告诉编译器dom api的接口，所以当我们在ts代码中使用dom的时候，比如说执行“document.getElementById(“root”)”这句话的时候，编译器就会知道该如何进行检查。如果我们不设置这个选项，那么编译器也有自己默认的库文件列表，一般来说是[“dom”, “es6”,“DOM.Iterable”]等等

##### 第三个：”allowJs”: true {#第三个：”allowJs”-true}

允许混合编译JavaScript文件

##### 第四个：”esModuleInterop”: true {#第四个：”esModuleInterop”-true}

这个选项允许我们使用commonjs的方式import默认文件。比如说，在没有开启这个选项时，我们需要这样写才能引用react，

```javascript
import * as React from 'react'
```

但是当我们开启了这个选项以后，import方式就与普通的JavaScript没有区别了，可以写为

```javascript
import React from 'react'
```

##### 第五个，”module”: “esnext”, {#第五个，”module”-“esnext”}

这里配置的是我们代码的模块系统，比较常见的有Node.js的CommonJS系统，ES6标准的esnext系统，以及requirejs的AMD系统。我们这里使用的是ES6标准的esnext系统，不过如果把这里换成CommonJS也是可以的。

##### 第六个，”moduleResolution”: “node”, {#第六个，”moduleResolution”-“node”}

这个选项决定了我们编译器的工作方式，也决定了我们各个文件之间调用、import的工作流程。这里曾经有两个选项，node 和 classic，但是classic这个选项在2019年12月就已经废弃了

##### 第七个：”isolatedModules”: true, {#第七个：”isolatedModules”-true}

开启这个选项以后，编译器会将每个文件作为单独的模块来使用。

##### 8. “noEmit”: true, {#8-“noEmit”-true}

开启这个选项表示当发生错误的时候，编译器不会生成 JavaScript 代码。

##### 9. “jsx”: “react”, {#9-“jsx”-“react”}

显而易见，这个选项允许编译器支持编译react代码

##### 10. “include”: [“src/\*\*/\*”] {#10-“include”-“src-”}

使用此选项列出我们需要编译的文件， “文件路径”选项需要文件的相对或绝对路径，例如：

“\*\*” - 任意子目录

“\*” - 任意文件名

“?” - 只要字符跟随“?”，这个字符就会被视为可忽略字符 (e.g., “src/.tsx?”则同时指代”src/.tsx”与”src/\*.ts”)

##### 11. “files”: [“./file1.ts”, “./file2.d.ts”, …] {#11-“files”-“-file1-ts”-“-file2-d-ts”-…}

使用此选项列出编译器应始终包含在编译中的文件。无论是否使用“exclude”选项，都将会编译使用此选项内包括的所有文件。

##### 12. “exclude”: [“node\_modules”, “\*\*//.test.ts”] {#12-“exclude”-“node-modules”-“-test-ts”}

此选项将会列出从编译中排除的文件。它与“include”选项采取相同的模式，我们以使用此选项来过滤使用“include”选项指定的文件。 但是, “exclude”选项不会影响“files”选项。

通常，我们会排除node\_modules、测试文件、和编译输出目录。

如果省略此选项，编译器将使用“outDir”选项指定的文件夹。

如果没有同时指定“files”和“include”这两个选项，则编译器将编译根目录和任何子目录中的所有TS文件，但不包括使用“exclude”选项指定的文件。
