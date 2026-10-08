---
title: 排序与突破O(n2)
date: '2019-05-25T22:11:45+08:00'
url: /2019/05/25/algorithm_sort/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### 复杂度 {#复杂度}

![image](http://201905-1251969284.cossh.myqcloud.com/sort_quality.png)

### 常用排序 {#常用排序}

#### Bubble Sort {#Bubble-Sort}

常识，不较介绍了

![image](http://201905-1251969284.cossh.myqcloud.com/Bubble_Sort.gif)

#### Selection Sort {#Selection-Sort}

选择最小的一个交换位置，交换次数比较少

![image](http://201905-1251969284.cossh.myqcloud.com/Selection_Sort.gif)

#### Insertion Sort {#Insertion-Sort}

不太喜欢这种思路

![image](http://201905-1251969284.cossh.myqcloud.com/Insertion_Sort.gif)

#### Shell Sort {#Shell-Sort}

是插入排序的一种更高效的改进版本，跟快排比起来有点尴尬

假设有这样一组数[ 13 14 94 33 82 25 59 94 65 23 45 27 73 25 39 10 ]，如果我们以步长为5开始进行排序，我们可以通过将这列表放在有5列的表中来更好地描述算法，这样他们就应该看起来是这样：  

```plain
13 14 94 33 82
25 59 94 65 23
45 27 73 25 39
10
```

然后我们对每列进行排序：  

```plain
10 14 73 25 23
13 27 94 33 39
25 59 94 65 82
45
```

将上述四行数字，依序接在一起时我们得到：[ 10 14 73 25 23 13 27 94 33 39 25 59 94 65 82 45 ].这时10已经移至正确位置了，然后再以3为步长进行排序：  

```plain
10 14 73
25 23 13
27 94 33
39 25 59
94 65 82
45
```

最后以1步长进行排序。

#### Merge Sort {#Merge-Sort}

过程可视化看这个:

[Merge Sort bilibili](https://www.bilibili.com/video/av18980253)

```plain
void merge_sort_recursive(int arr[], int reg[], int start, int end) {
    if (start >= end)
        return;
    int len = end - start, mid = (len >> 1) + start;
    int start1 = start, end1 = mid;
    int start2 = mid + 1, end2 = end;
    merge_sort_recursive(arr, reg, start1, end1);
    merge_sort_recursive(arr, reg, start2, end2);
    int k = start;
    while (start1 <= end1 && start2 <= end2)
        reg[k++] = arr[start1] < arr[start2] ? arr[start1++] : arr[start2++];
    while (start1 <= end1)
        reg[k++] = arr[start1++];
    while (start2 <= end2)
        reg[k++] = arr[start2++];
    for (k = start; k <= end; k++)
        arr[k] = reg[k];
}

void merge_sort(int arr[], const int len) {
    int reg[len];
    merge_sort_recursive(arr, reg, 0, len - 1);
}
```

#### Quick Sort {#Quick-Sort}

过程可视化看这个:

[快排 bilibili](https://www.bilibili.com/video/av39093184)  

```plain
void swap(int *x, int *y) {
    int t = *x;
    *x = *y;
    *y = t;
}

void quick_sort_recursive(int arr[], int start, int end) {
    if (start >= end)
        return;
    int mid = arr[end];
    int left = start, right = end - 1;
    while (left < right) {
        while (arr[left] < mid && left < right)
            left++;
        while (arr[right] >= mid && left < right)
            right--;
        swap(&arr[left], &arr[right]);
    }
    if (arr[left] >= arr[end])
        swap(&arr[left], &arr[end]);
    else
        left++;
    if (left)
        quick_sort_recursive(arr, start, left - 1);
    quick_sort_recursive(arr, left + 1, end);
}
```

#### Heap Sort {#Heap-Sort}

看这个，很容易懂，利用二叉树这种结构:

[Heapsort visualization (墙)](https://www.youtube.com/watch?v=MtQL_ll5KhQ)

然后再看这个，很清晰:

![image](http://201905-1251969284.cossh.myqcloud.com/Sorting_heapsort_anim.gif)

### 突破 O(n2) {#突破-O-n2}

排序能突破O(N^2)的界，可以用逆序数来理解，假设我们要从小到大排序，一个数组中取两个元素如果前面比后面大，则为一个逆序，容易看出排序的本质就是消除逆序数，可以证明对于随机数组，逆序数是O(N^2)的，而如果采用“交换相邻元素”的办法来消除逆序，每次正好只消除一个，因此必须执行O(N^2)的交换次数，这就是为啥冒泡、插入等算法只能到平方级别的原因。

反过来，基于交换元素的排序要想突破这个下界，必须执行一些比较，交换相隔比较远的元素，使得一次交换能消除一个以上的逆序，归并、快排、堆排等等算法都是交换比较远的元素，只不过规则各不同罢了
