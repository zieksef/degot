# Python ORM 与数据库规范

Python 数据库访问默认使用 SQLAlchemy。Django 项目使用 Django ORM；已有明确 ORM 选型的项目沿用项目约定。

只有查询过于复杂、无法清晰表达为 ORM 调用时，才允许使用原生 SQL。

## Schema

| 对象 | 规则 | 示例 |
| --- | --- | --- |
| Table | `snake_case`，单数名词 | `payment`, `merchant`, `card` |
| Column | `snake_case` | `account_id`, `created_at` |
| Index | `idx_{table}_{columns}` | `idx_payment_merchant_id` |
| Unique index | `uk_{table}_{columns}` | `uk_card_card_number` |

每张表必须包含内部主键和时间字段：`id`、`created_at`、`updated_at`。

需要软删除的表必须添加 `deleted_at`；`deleted_at` 是唯一允许为 `NULL` 的字段，`NULL` 表示未删除。

内部主键使用自增 ID；对外暴露必须使用独立业务 ID 字段，例如 `payment_id`。

## 枚举

数值枚举从 3 到 4 位数值开始，例如 `100` 或 `1000`，不使用负数。

字符串枚举全部小写，不使用混合大小写。

## SQL 查询

- 单次查询最多关联 3 张表；超过时重新设计查询或拆分成多次查询。
- 字段数超过 20 个的表，才考虑显式选择必要字段；实施前必须询问用户是否需要。
- 简单等值条件优先使用 ORM 字段表达，减少裸字段名字符串。
- 写入操作使用 ORM 的 create、save 或明确的 update 方法。
- 不在循环中查询数据库，改用批量查询。
- 不在循环中逐条插入或更新，改用批量操作。
- Repository 方法接收 caller 传入的 session 或 transaction，不在方法内部私自创建连接。
- 查询参数不超过 3 个时直接传参；超过 3 个时使用 query object。
