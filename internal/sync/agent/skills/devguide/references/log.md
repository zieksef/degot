# 日志与错误封装规范

## 核心原则

- 日志必须能帮助定位问题，而不是只说明发生了问题。
- 错误封装按“动作轨迹”逐层传递：外层到内层，每层只描述自己正在做的客观动作。
- `failed to`、`error`、`unable to` 这类结果性描述只出现在最上层日志输出，不出现在中间层错误包装里。
- 普通情况下，在调用链顶层统一记录日志；中间层返回带上下文的错误。
- 日志使用项目的结构化 logger，以 key-value 字段输出；字段 key 使用 `snake_case`。

## 敏感数据

日志和错误封装中禁止出现完整敏感数据：卡号（PAN）、CVV、卡有效期、密码、令牌、密钥、身份证件号等。

- 需要定位时使用脱敏形式：卡号只保留后四位（`card_last4: 1234`），令牌只记录标识或哈希前缀。
- 不确定某字段是否敏感时，先询问用户，不静默决定。

## 错误封装

错误封装的目标是保留定位问题所需的上下文。每一层只补充当前层知道的事实，不重复表达“失败”。

错误封装应包含：

- 当前层正在执行的业务动作。
- 当前层直接持有的关键参数。
- 原始错误。

错误封装不应包含：

- `failed to`、`could not`、`unable to`、`error` 等结果性描述。
- 与当前层无关的上游或下游动作。
- 无助于定位问题的泛化描述。

## 链式错误

链式错误应形成从外到内的动作轨迹。外层描述业务入口，内层描述更具体的执行步骤，末尾保留原始错误。

```text
create payment: call gateway (merchant_id m_123): make request: do post: https://api.example.com/payments: context canceled
```

链路中的每一段都应能回答一个问题：当前层正在做什么，以及它掌握了哪些关键参数。

段内参数使用 `(key value)` 形式，不使用 `:`；`:` 只作为链路层级分隔符，避免参数与层级混淆。

## 日志内容

每条日志都必须带上下文。至少包含关键输入参数；必要时也包含输出结果、外部请求目标、状态或数量。

下面的示例说明“该带哪些信息”，不是字面输出格式；实际日志用项目的结构化 logger 以 key-value 字段输出。

```text
Bad:
user login
create card error
index error

Good:
user login, user_id: 123, login_time: 12:04, ip: 127.0.0.1
create card error, account_id: 123, card_schema: visa, card_form: virtual
index 15 out of range 10
```

日志输出链式错误时，日志消息可以表达结果，错误内容负责提供完整动作轨迹。

```text
Bad:
unable to create payment: failed to call gateway: failed to do request: post https://api.example.com/payments: context canceled

Good:
message: failed to create payment
error: create payment: call gateway (merchant_id m_123): make request: do post: https://api.example.com/payments: context canceled
```

## 日志等级

日志等级用于表达处理结果的严重程度和是否需要人工关注，不等同于当前处理结果是否成功。线上告警通常绑定 `WARN` 及以上，因此已知、可预期、按设计处理且无需人工关注的结果不得使用 `WARN` 或 `ERROR`。

- `DEBUG`：只在主动排查时需要的细节，线上默认不关注。
- `INFO`：系统按设计完成一次处理，结果已知且可预期；该结果可以是成功，也可以是按规则拒绝、跳过、取消或无结果。
- `WARN`：处理过程偏离正常路径，问题已被识别、控制或恢复；当前职责未必失败，但仍需要人工关注趋势、补偿或依赖稳定性。
- `ERROR`：系统没有按设计完成当前职责，原因来自内部逻辑、外部依赖、数据状态或未知异常，需要工程或运维排查。
- `FATAL` / `PANIC`：进程无法继续运行，只用于启动失败或不可恢复的运行时状态。

判断 `WARN` 及以上的核心标准：这条日志是否应该进入人工关注或告警链路。判断 `ERROR` 的核心标准：当前职责是否已经失败，而不是当前处理结果是否符合调用方期望。

如果存在协议、任务状态或错误码，可以作为辅助判断，但责任归属优先于外部状态码：已知且按设计处理的结果不应升级为 `WARN` 或 `ERROR`；由系统实现、依赖、数据或未知异常导致职责无法完成时，应记录为 `ERROR`。

边界示例（最容易判错的场景）：

- 支付被风控规则按设计拒绝 → `INFO`（已知、可预期的处理结果，不是失败）。
- 幂等请求命中重复、按设计跳过 → `INFO`。
- 调用下游超时后重试成功 → `WARN`（偏离正常路径但已恢复，需关注依赖稳定性）。
- 下游持续不可用触发熔断降级 → `WARN`。
- 数据库写入失败导致下单未完成 → `ERROR`（当前职责未完成，需排查）。
- 反序列化外部响应时发生 panic → `ERROR`。
