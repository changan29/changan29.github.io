---
title: LeetCode_26 Remove Duplicates from Sorted Array
date: '2019-03-12T19:08:45+08:00'
url: /2019/03/12/LeetCode26_remove-duplicates-from-sorted-array/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### Remove Duplicates from Sorted Array {#Remove-Duplicates-from-Sorted-Array}

Given a sorted array nums, remove the duplicates in-place such that each element appear only once and return the new length.

Do not allocate extra space for another array, you must do this by modifying the input array in-place with O(1) extra memory.

**Example 1:**

```plain
Given nums = [1,1,2],

Your function should return length = 2, 
with the first two elements of nums being 1 and 2 respectively.

It doesn't matter what you leave beyond the returned length.
```

Example 2:  

```plain
Given nums = [0,0,1,1,1,2,2,3,3,4],

Your function should return length = 5, 
with the first five elements of nums being modified to
0, 1, 2, 3, and 4 respectively.

It doesn't matter what values are set beyond the returned length.
```

```plain
int removeDuplicates(vector<int>& nums)
{
    if(nums.size() == 0 || nums.size() == 1) return nums.size();
    int i =0;
    for (int j = 1; j < nums.size(); ++j) {
        if(nums[j] != nums[i])
        {
            i++;
            nums[i] = nums[j];
        }
    }
    return i+1;
}
```
