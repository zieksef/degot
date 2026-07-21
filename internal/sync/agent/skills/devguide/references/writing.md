# 注释与文档规范

Code tells you how, comments tell you why：只写代码/命令本身表达不了的信息——不明显的约束、边界情况、权衡、外部原因。能自解释的不写。

同一原则适用于代码注释、文档、README、commit 正文：简洁但明了，不含废话、复述、流水账、背景铺垫。

禁止：

- 复述代码的冗余注释（`i++ // i 加一`）。
- 过程流水账（`// 现在创建 payment`、`// 遍历订单列表`）。
- 指向本次改动或对话的注释（`// 改用 UTC`、`// 新增字段`、`// 按需求调整`）——这类信息属于 commit message。

```go
// Bad —— 复述代码做了什么
// 把金额转成分
amountInCents := amount.Mul(decimal.NewFromInt(100))

// Good —— 解释代码没法自己说明的原因
// 下游网关只接受整数分，不接受小数
amountInCents := amount.Mul(decimal.NewFromInt(100))
```
