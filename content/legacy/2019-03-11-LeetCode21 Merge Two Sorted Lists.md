---
title: LeetCode_21 Merge Two Sorted Lists
date: '2019-03-11T23:08:45+08:00'
url: /2019/03/11/LeetCode21 Merge Two Sorted Lists/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### Valid Parentheses {#Valid-Parentheses}

Merge two sorted linked lists and return it as a new list. The new list should be made by splicing together the nodes of the first two lists.

Example:

1. Input: 1->2->4, 1->3->4
2. Output: 1->1->2->3->4->4

```plain
ListNode* mergeTwoLists(ListNode* l1, ListNode* l2) {
    if(l1 == NULL ){ return  l2;}
    if(l2 == NULL ){ return  l1;}
    ListNode* Header = l1;
    ListNode* flag1 = l1;
    ListNode* flag2 = l2;

    if(flag1->val <= flag2->val)
    {
        Header = flag1;
    }else
    {
        ListNode* swp= NULL;
        swp = flag1;
        flag1 = flag2;
        flag2 = swp;
        Header = flag1;
    }

    while (flag1 && flag2)
    {
        if(flag1->val <= flag2->val)
        {
            if(flag1->next == NULL)
            {
                flag1->next = flag2;
                return Header;
            }
            else{
                ListNode* tp = flag1->next;
                while (tp)
                {
                    if(tp->val > flag2->val)
                    {
                        ListNode* tmp = flag1->next;
                        flag1->next = flag2;
                        flag2 = flag2->next;
                        flag1->next->next = tmp;
                        flag1= flag1->next;
                        break;
                    }
                    else
                     {
                        if(tp->next == NULL)
                        {
                            tp->next = flag2;
                            return Header;
                        }
                        tp = tp->next;
                        flag1 = flag1->next;
                    }
                }
            }
        }
    }

    return Header;
}
```
