---
title: LeetCode_20 Valid Parentheses
date: '2019-03-11T13:08:45+08:00'
url: /2019/03/11/LeetCode20_Valid_Parentheses/
draft: false
categories: []
tags: []
build:
  list: never
  render: always
---

### Valid Parentheses {#Valid-Parentheses}

Given a string containing just the characters ‘(‘, ‘)’, ‘{‘, ‘}’, ‘[‘ and ‘]’, determine if the input string is valid.

An input string is valid if:

1. Open brackets must be closed by the same type of brackets.
2. Open brackets must be closed in the correct order.

Note that an empty string is also considered valid.

```plain
class Solution {
public:
    bool isMatch(char c1, char c2)
{
    if(c2 == '}' && c1 == '{'){ return true;}
    if(c2 == ']' && c1 == '['){ return true;}
    if(c2 == ')' && c1 == '('){ return true;}
    return false;
}

bool isValid(string s) {
    // first blood
    if(s.size() == 0) return true;
    if(s.size() %2  == 1) return false;
    // })]...
    if(s[0] == '}' || s[0] == ']' || s[0] == ')')
    {
        return false;
    }
    stack<char> ss;

    for (int i = 0; i < s.size(); ++i) {
        if( s[i] == '{' || s[i] == '[' || s[i] == '(')
        {
            ss.push(s[i]);
        }
        if( s[i] == '}' || s[i] == ']' || s[i] == ')')
        {
            if(isMatch(ss.top(),s[i]))
            {
                ss.pop();
            } else{
                return false;
            }
        }
    }
    if (ss.size() == 0)
        return true;
    else
        return false;
}
};
```
