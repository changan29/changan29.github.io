---
title: unity3d 手游从0到0.01
date: '2018-05-26T22:21:39+08:00'
url: /2018/05/26/unity3d_from_0_to_0.01/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### why？ {#why？}

为什么要写这个东西，本身对u3d没有特别大的兴趣，但是毕竟在游戏行业，对前端一无所知好尴尬，而且客户端除了图形部分，还有业务逻辑，想要拿到一个客户端的项目能做到下断点调试。

两天时间不可能对整个引擎了解的特别深入，写了部分个人看法，属于个人理解，过一段时间可能就不这样认为了，涉及如下:

- 王者荣耀里面的安其拉原型是怎么做出来的？
- 怎么通过 Unity 编写游戏客户端逻辑
- 游戏里面的人物 怎么控制其射击，放技能
- 一款游戏的代码断点调试

### 王者荣耀里面的安其拉原型是怎么做出来的 {#王者荣耀里面的安其拉原型是怎么做出来的}

首先，一款游戏是很多人合作完成的结果，其中有游戏策划，美术和程序，安琪拉这个英雄最开始肯定是策划想要做这个英雄，设计这个英雄的各个技能是怎么样的，设计的方式多种多样，最有效的策划方式是 抄一个 哈哈

安琪拉的人物确定下来，是这样的  
![image](http://otc9g6h5k.bkt.clouddn.com/142-mobileskin-1.jpg)

接下来就是美术制作模型，美术常用的工具是 3dmax maya等，u3d对绝大多数建模软件都支持，在建模软件中做好模型 这种建模工具，制作模型的过程大概是这样的:

[暴走萝莉安琪拉](https://www.bilibili.com/video/av14785189/)

模型这部分主要是美术的工作，模型导出到 unity 然后由程序去驱动

### 怎么通过 Unity 编写游戏客户端 {#怎么通过-Unity-编写游戏客户端}

unity 一般是由不同的场景构成的，一般拿到一个 unity工程可以去 scenes 文件下 打开以 .unity 结尾的文件，我们在 一个场景中加载 多个 game object, 我们把上文中 美术做好的模型导入到场景中，设置好 光源，摄像机位置和模型比例，看到的是这样的  
(没有王者荣耀的资源，这是一个射击游戏的，<https://github.com/changan29/SimpleSwat-Unity>)  
对于场景中的物体，可以将 c# 脚本挂载到物体上，选定操作的对象，可以看到绑定到其上的脚本  
![image](http://otc9g6h5k.bkt.clouddn.com/obj_code.png)  
unity 中挂载到游戏对象中的脚本都是以下这种形式，全都继承 MonoBehaviour  
unity和mono的关联，分享很多，类似: [unity3d中脚本生命周期](https://blog.csdn.net/qitian67/article/details/18516503)

```plain
public class BasicController : MonoBehaviour {
	void Start ()
	{
		// 一些初始化操作
	}
	void Update ()
	{
		/*
		Update是在每次渲染新的一帧的时候才会调用，
		也就是说，这个函数的更新频率和设备的性能有关以及被渲染的物体（可以认为是三角形的数量）。
		在性能好的机器上可能fps 30，差的可能小些。
		这会导致同一个游戏在不同的机器上效果不一致，有的快有的慢。因为Update的执行间隔不一样了。
		*/
	}
	void FixedUpdate()
	{
		/*
		FixedUpdate是在固定的时间间隔执行，不受游戏帧率的影响。
		*/
	}
}
```

这样 我们就能通过引擎注册的这几个特殊函数 Update FixedUpdate 写自己的逻辑

引擎能帮你做的事情很多，构建场景，各种控件，物理引擎，粒子系统都是具备的，继承他，拖几个控件，在继承的代码里多写几行，显得你也很牛逼。。

### 游戏里面的人物 怎么控制其射击，放技能 {#游戏里面的人物-怎么控制其射击，放技能}

首先，我们看这个游戏怎么玩：

提到美术同学做模型，其实模型里面是可以加入动画的，以前的RPG游戏，很多动画是由程序切 一系列的图，达到操作人物的某些动作。

在实际的游戏中，用的最多的是骨骼动画、关节动画、关键帧动画三种基本的动画。  
在关键帧动画中，模型在每个关键帧中都是一个固定的姿势，相当于一个“快照”，通过在不同的关键帧中进行插值平滑计算，可以得到一个较为流畅的动画表现。关键帧动画的一个优势是只需要做插值计算，相对于其他的动画计算量很小，但是劣势也比较明显，基于固定的“快照”进行插值计算，表现大大被限制，同时插值如果不够平滑容易出现尖刺等现象。

　　关节动画是早期出现的一种动画，在这种动画中，模型整体不是一个Mesh, 而是分为多个Mesh，通过父子的关系进行组织，这样父节点的Mesh就会带动子节点的Mesh进行变换，这样层层的变换关系，就可以得到各个子Mesh在不同关键帧中的位置。关节动画相比于关键帧动画，依赖于各个关键帧的动画数据，可以实时的计算出各个Mesh的位置，不再受限于固定的位置，但是由于是分散的各个Mesh，这样在不同Mesh的结合处容易出现裂缝。

　　骨骼动画是进一步的动画类型，原理构成极其简单，但是解决问题极其有优势。将模型分为骨骼Bone和蒙皮Mesh两个部分，其基本的原理可以阐述为：模型的骨骼可分为基本多层父子骨骼，在动画关键帧数据的驱动下，计算出各个父子骨骼的位置，基于骨骼的控制通过顶点混合动态计算出蒙皮网格的顶点。在骨骼动画中，通常包含的是骨骼层次数据，网格Mesh数据， 网格蒙皮数据Skin Info和骨骼的动画关键帧数据。  
　　  
　　  
　　骨骼的本质，其实就是一个坐标空间，我们在做骨骼动画的时候，关键帧中包含的对骨骼的变换主要为旋转矩阵，所以对骨骼的变换就是对骨骼空间的旋转变换。说简单点，一个骨骼动画，带来的变换，首先作用在根骨骼上，影响根骨骼的坐标空间，然后递归的影响根骨的子骨骼，这样层层的递归影响，最后带来的就是整体骨骼变换。

我们在一个实际游戏中再看一下，还是上面那个射击游戏，定位到要射击的人物上，发现其绑定了Animator ，  
![image](http://otc9g6h5k.bkt.clouddn.com/animator.png)  
Animator Controller在Unity中是作为一种单独的配置文件存在的文件类型，其后缀为controller，Animator Controller包含了以下几种功能：

```
- 可以对多个动画进行整合
- 使用状态机来实现动画的播放和切换
- 可以实现动画融合和分层播放
- 可以通过脚本来对动画播放进行深度控制
```

在引擎中我们怎么使用这个组件呢？在Mecanim中，动画之间的播放不再是通过调用诸如“Play”之类的方法进行切换了，而是通过判断参数的变换来进行状态即动画的切换。  
这些状态的变化，unity 通过一个 状态图表示：  
![image](http://otc9g6h5k.bkt.clouddn.com/animator_img.png)

当我们在代码中切换状态时，会控制不同的动画变化

```plain
/*
部分代码，animator是绑定到人物对象上的 Animator 
*/
if(Input.GetKey(KeyCode.Q))
        {
			animator.SetBool("TurnLeft", true);
			transform.Rotate(Vector3.up * (Time.deltaTime * -45.0f), Space.World);
		}  else
        {
		    animator.SetBool("TurnLeft", false);	
		}
		if(Input.GetKey(KeyCode.E))
		{
			animator.SetBool("TurnRight", true);
			transform.Rotate(Vector3.up * (Time.deltaTime * 45.0f), Space.World);
				
		}  else
        {
		    animator.SetBool("TurnRight", false);	
		}
		if(Input.GetKeyDown(KeyCode.F) && animator.layerCount >= 2)
        {
			animator.SetBool("Grenade", true);
		} else
        {
			animator.SetBool("Grenade", false);
		}
		if(Input.GetButtonDown("Fire1") && animator.layerCount >= 2 && Time.time>nextFire)
        {
			animator.SetBool("Fire", true);
            nextFire = Time.time + fireRate;
            Instantiate(shot, Lo_Muzzle.position, Lo_Muzzle.rotation);
            GetComponent<AudioSource>().Play();
		}
```

### 一款游戏的代码断点调试 {#一款游戏的代码断点调试}

一般写 C# 在 VS中比较方便，unity和VS可以一起调试，在VS中看代码下断点,只要在安装VS的时候选择unity 选项就好了  
![image](http://otc9g6h5k.bkt.clouddn.com/u3d_debug.png)

然后去 unity 运行工程，达到断点条件，可见执行部分的堆栈  
![image](http://otc9g6h5k.bkt.clouddn.com/u3d_debug2.png)

### 参考 {#参考}

- <https://www.cnblogs.com/SHOR/p/5735109.html>
- <http://www.cnblogs.com/zblade/p/6986173.html>
- <http://www.cnblogs.com/zhanglitong/p/3196752.html>

@Sat May 26 22:16:52 CST 2018
