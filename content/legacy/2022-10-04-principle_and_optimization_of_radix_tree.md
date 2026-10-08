---
title: radix tree原理与优化
date: '2022-10-04T21:31:45+08:00'
url: /2022/10/04/principle_and_optimization_of_radix_tree/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### radix tree原理与优化 {#radix-tree原理与优化}

主要参考论文: <https://db.in.tum.de/~leis/papers/ART.pdf>

#### background {#background}

- 内存变得容量更大更便宜，整个数据库/存储引擎可以放到内存
- 传统的内存数据索引一般是平衡二叉树 在现代的硬件上效率不高
- 另一类索引结构，hash表 只支持单点查找，范围查找不好
- 索引效率是决定性能的关键因素

  ##### introduction {#introduction}
- T树和二叉树 硬件支持不好 on modern hardware architectures
- B+ tree 缓存敏感，但是update代价比较大
- The k-ary search tree and the Fast Architecture Sensitive Tree (FAST) 使用数据级并行性与单指令多数据 (SIMD) 指令同时执行多重比较。此外，FAST 使用的数据布局通过优化利用缓存行和 TLB 来避免缓存未命中。但两种数据结构都不能支持增量更新。
- 哈希表随机分散了键，只支持点查询；另一个问题是大多数哈希表不能优雅地处理增长
- GPU 可以实现比 CPU 更高的吞吐量。然而，使用 GPU 作为专用索引硬件还不实用，因为 GPU 的内存容量有限，与主存的通信成本很高，
- Trie

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663689380458-07a22c30-a954-4a7e-abe0-272c7f10e1b7.png#averageHue=%23f7f7f7&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=302&id=u60008ddd&margin=%5Bobject%20Object%5D&name=image.png&originHeight=604&originWidth=572&originalType=binary&ratio=1&rotation=0&showTitle=false&size=66255&status=done&style=none&taskId=ua0e3635b-83bc-494e-a469-33a567920c7&title=&width=286)  
Trie 大多数研究都集中在索引字符串上，但我们的目标是索引其他数据类型。因此，我们更喜欢术语基数树而不是 trie，因为它强调了与基数排序算法的相似性，并强调可以索引任意数据而不仅仅是字符串。  
在实际实现上，“基” 一般是 2^K中的 K，下面解释。

### radix tree {#radix-tree}

按字母顺序排列的字典书中的拇指索引。一个词的第一个字符可以直接用来跳到所有以该字符开始的词。在计算机中，这个过程可以用接下来的字符重复进行，直到找到一个匹配的字符。作为这个过程的结果，所有的操作都有O(k)的复杂性，其中k是key的长度。  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663687816774-c86f1750-3e28-4b36-822d-baf1bea1e5ff.png#averageHue=%23f5f4f4&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=202&id=ua899c8e7&margin=%5Bobject%20Object%5D&name=image.png&originHeight=404&originWidth=686&originalType=binary&ratio=1&rotation=0&showTitle=false&size=50854&status=done&style=none&taskId=ue60b8953-b076-4e79-95d9-b9bf0aa5506&title=&width=343)  
特点：

- 路径压缩和懒惰扩展，使ART能够通过折叠节点有效地索引长键，从而降低树的高度
- 根据子节点的数量动态地选择结构
- 树的高度与Node个数无关，与key长度有关
- all insertion orders result in the same tree
- key是隐藏的，可以从路径获得
- 不需要平衡
- Key按字典顺序存储

  #### Adaptive Nodes {#Adaptive-Nodes}

  ![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663724116777-a36c15fd-9164-400c-ad95-60ab32c32db8.png#averageHue=%23f7f7f7&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=176&id=u146062ae&margin=%5Bobject%20Object%5D&name=image.png&originHeight=352&originWidth=1284&originalType=binary&ratio=1&rotation=0&showTitle=false&size=39253&status=done&style=none&taskId=ub5ca45d6-4494-471b-97b7-e08e9e9ef94&title=&width=642)

  #### node span {#node-span}

  ![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663828940783-d2f377e3-6766-4e89-80b8-33d712836aca.png#averageHue=%23f8f8f8&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=498&id=u58f50794&margin=%5Bobject%20Object%5D&name=image.png&originHeight=996&originWidth=1360&originalType=binary&ratio=1&rotation=0&showTitle=false&size=194338&status=done&style=none&taskId=u2e2da33a-2650-4c1a-abda-3fdcd697644&title=&width=680)  
  ![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663724149979-beb2d3a7-0a4a-4226-a6b6-830529669e3e.png#averageHue=%23f1f1f1&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=242&id=ud2faeea9&margin=%5Bobject%20Object%5D&name=image.png&originHeight=484&originWidth=722&originalType=binary&ratio=1&rotation=0&showTitle=false&size=63409&status=done&style=none&taskId=uba56655f-df03-4e63-a7bf-315ed885e1a&title=&width=361)  
  影响因素：
- Span越⼤， ⾼度越低
- Span越⼤ ， 空间占⽤越⼤

问题: 选多大的span？ 1. 折中 2. 可变的node类型  
-> 可变基数树 node类型可变  
处于性能考虑，节点的类型比较少，如果是几十类，resize的代价会很高

##### Inner Nodes {#Inner-Nodes}

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663724833341-70415f4a-4659-41b6-aced-aaea0cc7e943.png#averageHue=%23f4f4f3&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=412&id=u7173bf28&margin=%5Bobject%20Object%5D&name=image.png&originHeight=824&originWidth=730&originalType=binary&ratio=1&rotation=0&showTitle=false&size=114706&status=done&style=none&taskId=uccf80716-5e10-433c-a98c-5cf893bd550&title=&width=365)

1. 插入时，当前node大小没有空间，扩展到大一级的node,同样，缩小时resize到小一级的node。
2. each inner node has at least two children.

   ##### Leaf Nodes {#Leaf-Nodes}

   ![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663726470513-d4309d9f-e2ef-42cd-8a5d-0576ce9d77ad.png#averageHue=%23f5f4f4&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=204&id=u21ec2d1b&margin=%5Bobject%20Object%5D&name=image.png&originHeight=408&originWidth=660&originalType=binary&ratio=1&rotation=0&showTitle=false&size=70757&status=done&style=none&taskId=u1640b387-118e-456f-8e43-0499f6809f3&title=&width=330)

   #### easy example {#easy-example}

   ```java
   put(0x3FAA01L,"0x3FAA01L");
   put(0x3FAA02L,"0x3FAA02L");
   put(0x4FAA03L,"0x4FAA03L");
   ```

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663726537319-701506d1-ace5-4692-8b0f-20ccc0d951fd.png#averageHue=%233c4042&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=283&id=u678ca7fc&margin=%5Bobject%20Object%5D&name=image.png&originHeight=566&originWidth=750&originalType=binary&ratio=1&rotation=0&showTitle=false&size=108712&status=done&style=none&taskId=ua47fd630-6771-4251-a4cf-b0235c59921&title=&width=375)

#### example 2 {#example-2}

```java
map.put(2,"2");
map.put(533, "533");
map.put(573, "573");
map.put(40001, "40001");
map.put(40021, "40021");
map.put(40023, "40023");
map.put(40044, "40044");
map.put(40086, "40086");

for(int i = 0;i<20;i++){
    long x = 0x3FCC01 + i;
    map.put(x , String.valueOf(x));
}
```

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663832495092-c2a594dd-9dd9-4494-95a2-c517fece45b3.png#averageHue=%23f8f4f1&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=407&id=u8849672e&margin=%5Bobject%20Object%5D&name=image.png&originHeight=814&originWidth=2618&originalType=binary&ratio=1&rotation=0&showTitle=false&size=142800&status=done&style=none&taskId=u47dd0e82-6842-4e0e-9c4f-1dec26c3955&title=&width=1309)

#### Algorithms {#Algorithms}

##### search {#search}

```java
get(key)
	cur = root
	while(cur){
		checkPrefix(key) != depth
			return null
		isLeaf(cur)
			return vale
		// 向下递归
		cur = cur.find(key)

	}
```

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663749650302-64af8248-a88c-479e-a1a1-73eb7144d131.png#averageHue=%23f2f2f2&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=316&id=uc8df0a34&margin=%5Bobject%20Object%5D&name=image.png&originHeight=632&originWidth=692&originalType=binary&ratio=1&rotation=0&showTitle=false&size=106315&status=done&style=none&taskId=ube09f1d1-ea46-473d-846f-08d4ffabfbc&title=&width=346)

##### Insert {#Insert}

```java
put(key , value){
	leaf = createLeaf(key,value)
	root == null ? root= leaf return
	cur = root
	while(true){
		if (checkPrefix(key) == true ){
			// 替换旧节点
			isLeaf() replace(leaf)

			//cur is inner
			child = cur.findChild(key)
			// 没有子树 branch 
			child == null {
				cur.needGrow() ? grow() 
				cur.put(leaf)
			} else {
            	// 递归向下
				cur = child
        	}
		} else {
			//分裂
			newNode = createNode4
			newNode.put(cur)
			newNode.put(leaf)
			curParent.adjust
		}

	}
}
```

##### grow & shrink {#grow-amp-shrink}

```java
grow(){
	node4 -> node16 {
		copy
	}

	node16 -> node48{
		// 规则转换
		key rules transform
		children index build 
	}

	node48 -> node256{
		// 没有keys数组了 children数组index转换
		children index by node48 key
	}
}
```

##### 分裂 {#分裂}

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663837692091-f7411f74-0c07-43ef-9c36-74dae3e266d2.png#averageHue=%23f9efea&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=436&id=uf9c1cbf9&margin=%5Bobject%20Object%5D&name=image.png&originHeight=872&originWidth=918&originalType=binary&ratio=1&rotation=0&showTitle=false&size=63671&status=done&style=none&taskId=u2b50b077-2ae1-45b1-b7ac-6c26d10403f&title=&width=459)  

```java
int checkPrefix(long key) {
        //从左到右 多少位是相同的
        //8位一个深度 depth 默认8 ，最深深度
        return Math.min(Long.numberOfLeadingZeros(key ^ nodeKey) / 8, depth);
}
// depth=6带一个Leaf, 再添加一个Leaf，创建一个node4 挂到 parent节点  如上例
```

##### evaluation {#evaluation}

| random | ART | TreeMap | diff |
| --- | --- | --- | --- |
| put (16M keys) | 12857 | 40562 | 215% |
| search (16M keys) | 7253 | 25413 | 250% |
| foreach (16M keys) | 695 | 2334 | 235% |
|  |  |  |  |
| put (64k keys) | 73 | 105 | 43% |
| search (64k keys) | 25 | 28 | 12% |
| foreach (64k keys) | 7 | 14 | 100% |
|  |  |  |  |
| put (1M keys) | 896 | 1420 | 58% |
| search (1M keys) | 373 | 684 | 83% |
| foreach (1M keys) | 52 | 73 | 40% |

| 连续 | ART | TreeMap | diff |
| --- | --- | --- | --- |
| put (16M keys) | 1999 | 4302 | 115% |
| foreach (16M keys) | 152 | 192 | 26% |
| put (650K keys) | 174 | 191 | 9% |
| foreach (650K keys) | 17 | 35 | 100% |

#### 分析性能产生的原因 {#分析性能产生的原因}

1. key越多，差距越大
2. 随机的比连续的差距大

随机的比连续的差距大 –> 考虑缓存的影响。  
对比二者的结构：  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663833868662-99bdf095-2794-45fa-81e8-68618b236b7d.png#averageHue=%23f8f8f8&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=153&id=u0c90c4bd&margin=%5Bobject%20Object%5D&name=image.png&originHeight=270&originWidth=534&originalType=binary&ratio=1&rotation=0&showTitle=false&size=14516&status=done&style=none&taskId=ue711e7ea-e52e-44bb-9747-7e225e148a6&title=&width=303) ![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663833886767-528b9ec9-e4c2-438d-91d5-5795e0c3ae91.png#averageHue=%23ece6e6&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=177&id=ud35e03d1&margin=%5Bobject%20Object%5D&name=image.png&originHeight=696&originWidth=1362&originalType=binary&ratio=1&rotation=0&showTitle=false&size=328182&status=done&style=none&taskId=ua3e69afe-74b4-4b49-a4b7-730a800c474&title=&width=347)

### 高速缓存存储器结构 {#高速缓存存储器结构}

撮合服务器 CPU型号是 Intel(R) Xeon(R) Platinum 8369B CPU @ 2.70GHz, intel官网没有（ L3 cache是49M，见附录），应该是订制的，找同类型的CPU参数:  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663813808533-27651b85-60ca-409f-b566-1d001af33fd4.png#averageHue=%23ededed&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=217&id=T6cLx&margin=%5Bobject%20Object%5D&name=image.png&originHeight=434&originWidth=1446&originalType=binary&ratio=1&rotation=0&showTitle=false&size=82853&status=done&style=none&taskId=u359bbce2-20d5-4f37-80c6-b830217ccee&title=&width=723)  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663814609517-b15acca5-8591-4184-b70e-81a8dcfe067e.png#averageHue=%23eeefec&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=444&id=DaHRY&margin=%5Bobject%20Object%5D&name=image.png&originHeight=888&originWidth=1308&originalType=binary&ratio=1&rotation=0&showTitle=false&size=660104&status=done&style=none&taskId=u19f3468c-22ff-4c89-9659-c4dbe7c6cd4&title=&width=654)

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663772602601-1738525a-d1e9-41c4-b986-61abbd0ba157.png#averageHue=%23e6eae3&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=276&id=AYIld&margin=%5Bobject%20Object%5D&name=image.png&originHeight=552&originWidth=975&originalType=binary&ratio=1&rotation=0&showTitle=false&size=59637&status=done&style=none&taskId=u9c89e5c4-614f-4154-a940-e27b8c67aef&title=&width=487.5)  
考虑一个计算机系统，其中每个存储器地址有 m 位，形成 M= 2^m个不同的地址。这样一个机器的高速缓存被组织成一个有 S=2^s个高速缓存组（cache set）的数组。每个组包含 E 个高速缓存行（cache line）。每个行是由一个 B=2^b 字节的数据块（block）组成的，一个有效位（valid bit）指明这个行是否包含有意义的信息（为了方便），还有 t = m-(b+s) 个标记位（tag bit）（是当前块的内存地址的位的一个子集），它们唯一地标识存储在这个高速缓存行中的块。  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663818149791-5bd6ba70-809b-4cdf-b145-74ae63ee0ebb.png#averageHue=%23f6f6f6&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=350&id=u636bd49c&margin=%5Bobject%20Object%5D&name=image.png&originHeight=700&originWidth=798&originalType=binary&ratio=1&rotation=0&showTitle=false&size=222852&status=done&style=none&taskId=u95986524-e513-4cc5-a696-7baf2591884&title=&width=399)  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663818319080-5b862ebb-2e46-47b7-ad3d-fa3dc33d5d85.png#averageHue=%23f6f5f5&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=858&id=u9518862e&margin=%5Bobject%20Object%5D&name=image.png&originHeight=1716&originWidth=1550&originalType=binary&ratio=1&rotation=0&showTitle=false&size=755131&status=done&style=none&taskId=u5bcd3030-8030-41a1-a674-d6f3cc66e69&title=&width=775)  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663834813587-8b89d9f5-efc5-4947-a0cc-357bbe03cda7.png#averageHue=%23f7f7f4&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=361&id=ektga&margin=%5Bobject%20Object%5D&name=image.png&originHeight=722&originWidth=1036&originalType=binary&ratio=1&rotation=0&showTitle=false&size=364670&status=done&style=none&taskId=u0c8ac52e-946a-42cc-8563-bd32d7f9744&title=&width=518)

##### 不命中时的行替换 {#不命中时的行替换}

替换组中的哪一行（没有空行）？替换策略：

- 最不常使用（Least-Frequently-Used，LFU）
- 最近最少使用（Least-Recently-Used，LRU）

所有这些策略都需要额外的时间和硬件。

### 另外的优化途径-SIMD {#另外的优化途径-SIMD}

#### 硬件支持 {#硬件支持}

![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663838518639-55937322-dbf2-4383-9b0c-dec58b1b0de8.png#averageHue=%23efefef&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=304&id=u43b0bd80&margin=%5Bobject%20Object%5D&name=image.png&originHeight=608&originWidth=1224&originalType=binary&ratio=1&rotation=0&showTitle=false&size=153643&status=done&style=none&taskId=u0c7e4e4d-38f1-4207-970e-b3f41e8a5fa&title=&width=612)  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663838548131-1c78a5e0-7a30-40a5-9bdf-04ea316faf6e.png#averageHue=%23f0f0f0&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=109&id=u9dea437a&margin=%5Bobject%20Object%5D&name=image.png&originHeight=218&originWidth=1452&originalType=binary&ratio=1&rotation=0&showTitle=false&size=134838&status=done&style=none&taskId=ufb657c3c-d55b-4739-9e79-f3571bcea13&title=&width=726)

#### 语言支持 {#语言支持}

Java JEP 417: Vector API (Third Incubator) on jdk17/18

- <https://openjdk.org/jeps/414>
- <https://openjdk.org/jeps/417>

  #### 例子 {#例子}

  ```java
  static final VectorSpecies<Integer> SPECIES = IntVector.SPECIES_256;

  void multiply(int[] array, int by) {
      int i = 0;
      int bound = SPECIES.loopBound(array.length);
      IntVector byVector = IntVector.broadcast(SPECIES, by);
      for (; i < bound; i += SPECIES.length()) {
          IntVector vec = IntVector.fromArray(SPECIES, array, i);
          IntVector multiplied = vec.mul(byVector);
          multiplied.intoArray(array, i);
      }
      for (; i < array.length; i++) {
          array[i] *= by;
      }
  }
  ```

#### Vector Api {#Vector-Api}

| 组件 | items |
| --- | --- |
| VectorSpecies | SPECIES\_512, 256, 128… |
| Vector | ShortVector, LongVector |
| VectorMask | firstTrue() , trueCount() , anyTrue()… |
| VectorOperators | EQ , LT, GT… |

#### 用SIMD 优化 radix tree的查询 {#用SIMD-优化-radix-tree的查询}

```java
   short[] sss = new short[16];
   for(short i = 0; i<16; i++){
       sss[i] = (short) (i*10 + 1);
   }
   // 1 11 21 31 41 51 61 ...
   VectorSpecies<Short> species = ShortVector.SPECIES_256;
   Vector<Short> vecArrays = species.fromArray(sss,0);

// compare 16 keys with one instruction in parallel. 
   int idx = vecArrays.compare(VectorOperators.EQ,51).firstTrue();
```

##### node 48、256 search {#node-48、256-search}

由于VectorApi不支持指针和bool， 对于children多的node，维护一个与children对应的bool flags[childrens]，通过 SIMD批量获取 index后，children.get(index)。

### summary {#summary}

- 适合Key⻓度固定或较短的场景
- 较大的Span和垂直压缩显著提高性能
- 可变的Node大小，显著提高空间利用率，并充分利用SIMD指令
- 使用数组存储数据对缓存更加友好
- 只支持字典序，不适合BigDecimal
- 可变Node与node256 性能近似，与TreeMap比显著提升

### 引用 {#引用}

<https://www.intel.com/content/www/us/en/products/sku/120504/intel-xeon-platinum-8168-processor-33m-cache-2-70-ghz/specifications.html>  
志强铂金8280: <https://en.wikichip.org/wiki/intel/xeon_platinum/8280>  
其他处理器: <https://ark.intel.com/content/www/cn/zh/ark.html#@Processors>  
硬件结构: <https://zh.wikipedia.org/wiki/File:Hwloc.png>  
![image.png](https://cdn.nlark.com/yuque/0/2022/png/2691051/1663835511810-4013030d-7ae3-4792-adac-c1052a882dc6.png#averageHue=%23dcdbdb&clientId=u5c92151a-cd4b-4&crop=0&crop=0&crop=1&crop=1&from=paste&height=335&id=uc2051b28&margin=%5Bobject%20Object%5D&name=image.png&originHeight=670&originWidth=1450&originalType=binary&ratio=1&rotation=0&showTitle=false&size=51770&status=done&style=none&taskId=ud3f2978b-a854-46fb-8945-a22678bdd71&title=&width=725)

#### 论文 {#论文}

- <https://db.in.tum.de/~leis/papers/ART.pdf>

  #### 参考实现 {#参考实现}
- <https://github.com/exchange-core/collections/tree/master/src/main/java/exchange/core2/collections/art>
- <https://github.com/rohansuri/adaptive-radix-tree>
