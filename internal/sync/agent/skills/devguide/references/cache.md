# 缓存规范

## Key 格式

```text
{domain}:{entity}:{identifier}[:{sub_identifier}]
```

最多 4 层，第四层可选。

- 使用 `:` 分隔层级。
- 层级内部使用 `_` 连接单词，业务自定义命名采用 `snake_case`。

## Key 层级

| 层级 | 必填 | 说明 | 示例 |
| --- | --- | --- | --- |
| `domain` | 是 | 业务域 | `payment`, `merchant`, `risk`, `fx`, `notification` |
| `entity` | 是 | 业务实体或用途 | `session`, `exchange_rate`, `config`, `retry_count` |
| `identifier` | 是 | 唯一标识 | 具体业务 ID |
| `sub_identifier` | 否 | 二级定位标识 | 子层级限定 |

## Key 示例

```text
payment:session:abc123
merchant:config:merchant_001
risk:score:txn_xyz
fx:exchange_rate:USD_HKD
notification:template:sms_otp
merchant:rate_limit:merchant_001:create_payment
```

## Key 规则

1. 业务自定义命名使用小写 `snake_case`，不使用 camelCase；系统级或标准化标识可以保留大写。
2. 层级之间使用 `:`，层级内部使用 `_`，不使用 `-` 或 `.`。
3. 必须包含 `domain`，用于按业务域隔离命名空间。
4. 最多 4 层，格式为 `domain:entity:identifier[:sub_identifier]`。
5. Key 只用于定位，不在 key 中嵌入数据；数据应放在 value 中。
6. Key 长度不得超过 128 bytes。

## 敏感数据

卡号（PAN）、CVV、卡有效期、密码、密钥等敏感数据不得明文写入缓存；确需缓存时必须加密或脱敏。

- 令牌、会话等鉴权数据可以按设计缓存，但必须设置 TTL，且不得出现在 key 中或写入日志。
- 不确定某字段是否敏感时，先询问用户，不静默决定。

## 过期时间

所有缓存默认必须设置过期时间，除非有明确记录的特殊原因。

- 默认不使用永久缓存，每个 key 都必须设置 TTL。
- 如果某个 key 确实需要永久保存，必须添加代码注释说明原因。

TTL 根据数据的更新频率和可容忍的陈旧程度选择。下表是参考起点，应按实际数据调整，不是固定值：

| 数据类型 | 参考 TTL |
| --- | --- |
| 热点配置 / 字典（merchant config） | 5–30 分钟 |
| 会话 / 令牌（session） | 按业务生命周期，分钟至小时 |
| 汇率等按周期刷新的数据（exchange_rate） | 按刷新周期，秒至分钟 |
| 幂等标记 / 重试计数 / 限流窗口 | 等于对应业务窗口时长 |
| 热查询结果 | 秒至分钟 |

## 缓存防护

- 缓存穿透：对查询结果为空的 key 也写入缓存（使用较短 TTL），避免反复穿透到数据库；命中量大时可用布隆过滤器前置过滤。
- 缓存击穿：热点 key 回源时用 singleflight、分布式锁或逻辑过期，保证同一时刻只有一个请求回源重建，其余等待结果。
