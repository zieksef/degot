---
name: devguide
description: "Team development standards. Use when writing, modifying, refactoring, or reviewing code, designing DB schemas or APIs, writing logs, using caches, or committing. Also invoked as /devguide or $devguide."
---

# Develop Guide

## 执行原则

- 编码前先明确假设、歧义和取舍；不确定时先问，不静默猜测。
- 控制最小公开面：默认内部可见，只有外部调用方明确需要时才公开或导出。

## 使用流程

1. 识别当前任务涉及的语言和领域。
2. 按下方路由表读取对应规范文件；一个任务可能命中多行，命中的都要读。
3. 涉及代码修改时，必须按对应语言 `style.md` 的要求执行格式化和静态检查。
4. 路由表中没有对应规范的领域，说明当前尚未提供具体规范；不要自行补写或编造规范内容。

## 规范路由

Go 规范覆盖所有 Go 项目，默认语境以后端服务为主；Python 规范覆盖 Python 项目，默认语境以后端服务和 SDK 为主。

| 任务信号 | 必读文件 |
| --- | --- |
| 任何 Go 代码修改 / 引入依赖 | `references/go/style.md` |
| GORM / 建表 / DDL / 数据库 schema | `references/go/orm.md` |
| Go API req/resp / DTO / 校验 / 客户端错误 | `references/go/api.md` |
| 任何 Python 代码修改 / 引入依赖 | `references/python/style.md` |
| Python ORM / 数据库访问 / SQL | `references/python/orm.md` |
| Python API req/resp / DTO / 校验 | `references/python/api.md` |
| 新增或修改测试 / 修复缺陷 / 改变既有行为 | `references/testing.md` |
| 日志 / 错误封装 | `references/log.md` |
| 缓存 key / TTL | `references/cache.md` |
| 分支 / commit / 合并策略 | `references/vcs.md` |
| 写注释 / 文档 / README | `references/writing.md` |

## 完成前检查

任务收尾时重新对照本节，逐项确认后再交付：

- [ ] 改过 Go：`goimports` 已格式化，`golangci-lint run` 通过。
- [ ] 改过 Python：`ruff check .` 和 `ruff format --check .` 通过。
- [ ] 修复缺陷 / 改变既有行为 / 修改测试：已按 `references/testing.md` 优先复用现有测试；新增测试函数有独立行为边界或不同测试准备；未保留重复、过时或按本次修改命名的测试。
- [ ] 注释/文档/README：符合 `references/writing.md`，简洁但明了，不含废话、复述、流水账、背景铺垫。
- [ ] 新增表或字段：命名与必需字段符合 `references/go/orm.md` 或 `references/python/orm.md`。
- [ ] 新增日志：等级和上下文符合 `references/log.md`。
- [ ] 新增缓存 key：格式和 TTL 符合 `references/cache.md`。
- [ ] 准备提交：分支名和 commit message 符合 `references/vcs.md`。
