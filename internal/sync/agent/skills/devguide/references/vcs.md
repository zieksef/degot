# 版本控制规范

## 分支创建

新功能或修复开始前，默认基于最新的 `release` 分支创建开发分支。

分支命名格式：

```text
{name}/{type}-{description}
```

- `name`：开发者标识，使用小写。
- `type`：分支类型，见下表；表中没有合适类型时，参照 commit type 命名。
- `description`：功能或修复描述，使用小写，单词之间用 `-` 连接。

| type | 用途 |
| --- | --- |
| `feature` | 新功能 |
| `fix` | 缺陷修复 |
| `hotfix` | 线上紧急修复 |
| `refactor` | 重构 |
| `chore` | 维护性变更 |

示例：

```text
zhangsan/feature-add-rate-limit
lisi/hotfix-fix-callback-timeout
wangwu/refactor-split-config-parser
```

## Commit 信息

Commit 信息参考 Conventional Commits 1.0.0：

https://www.conventionalcommits.org/en/v1.0.0/

基本格式：

```text
<type>[optional scope][!]: <description>

[optional body]

[optional footer(s)]
```

## 类型与范围

- `feat`：新增能力或用户可感知的功能。
- `fix`：修复缺陷。
- `docs`：仅文档变更。
- `refactor`：不改变外部行为的代码重构。
- `test`：测试相关变更。
- `chore`：维护性变更。
- `build`：构建系统或依赖变更。
- `ci`：CI/CD 配置或流程变更。
- `perf`：性能优化。
- `style`：格式、排版等不影响行为的变更。

Scope 用于标识影响范围，优先使用模块、包、服务、命令或配置域名称。Scope 应简短稳定，不要写成一句描述。

## 详细程度

Commit 信息必须能让读者不看 diff 也理解主要变更。

- Subject 描述“做了什么”，避免只写 `update`、`fix bug`、`change code` 这类泛化内容。
- 变更原因、行为影响、兼容性影响、迁移说明、配置变化或风险点不明显时，必须写 body。
- 一个 commit 覆盖多个不相关意图时，优先拆分 commit。
- Footer 用于记录 issue、review、breaking change 等结构化补充信息。

## 语言

- `type`、`scope`、`BREAKING CHANGE`、issue key、模块名、API 名、配置 key、错误码等术语可以保留英文。
- `<description>`、body、footer 中的说明性信息使用中文。
- 中文描述保持具体、客观，避免口语化和无信息量表述。

## Commit 示例

```text
feat(card): 支持虚拟卡创建流程

新增虚拟卡创建参数校验和下游请求组装。
调用方现在可以通过 card_form=virtual 创建虚拟卡。
```

```text
fix(login): 修复验证码过期时间计算错误

原逻辑使用本地时区计算过期时间，跨时区部署时会提前失效。
改为使用 UTC 时间进行计算。
```

```text
refactor(config): 拆分网关插件配置解析逻辑

提取插件组解析流程，保持现有 YAML 字段和运行时行为不变。
```

```text
feat(api)!: 调整账户查询响应结构

账户余额字段从 balance 改为 available_balance。

BREAKING CHANGE: 旧字段 balance 不再返回，调用方需要改用 available_balance。
```

## 合并策略

默认使用 **Merge Commit**，保留完整提交历史。

当开发分支包含多个只服务于开发过程、单独保留没有审计或回滚价值的 commit 时，必须使用 **Squash Merge**，合并为一个清晰、完整的提交。

例如：

- 临时修复、反复调整、格式修正。
- `fix review comments`、`update`、`wip` 这类无法独立说明业务或技术意图的提交。
- 多个提交共同完成同一个不可拆分的功能或修复。
