# Python 代码规范

Python 代码必须使用 Ruff 格式化和静态检查。

每次开发完成或构建前，必须运行 `ruff check .` 和 `ruff format --check .`。

必须使用项目已有 Ruff 配置，不临时绕过规则。

## 注释

注释规范见 `references/writing.md`。

## 包管理与虚拟环境

Python 包管理、依赖安装、虚拟环境管理和命令执行必须使用 `uv`。

不要新增 Poetry、Pipenv、conda、virtualenv、venv 或 pip-tools 管理路径。

## 文件组织

不要默认创建 `types.py`、`utils.py`、`common.py`。

代码应优先放在表达领域或职责的文件中。只有当一组内容确实属于同一个明确主题，且没有更清晰的领域文件名时，才允许使用聚合文件名。

## 命名

在使用泛化命名前，先确认是否有更好的领域命名；没有更好名字时，才使用泛化命名。

类型别名按类型命名，使用 `PascalCase`；有明确归属时优先带上领域前缀，例如 `ServerOption`。

## 函数与方法

函数和方法必须声明参数类型和返回类型；`__init__` 的返回类型写 `None`。

## 初始化

复杂对象初始化优先使用 options 模式：必要参数放在 `__init__` 的显式参数中，可选配置通过 `*options` 传入，并在 `__init__` 内应用。

不要把必要参数塞进 config 或 options 对象里。

```python
from collections.abc import Callable

ServerOption = Callable[["Server"], None]


class Server:
    def __init__(self, name: str, base_url: str, *options: ServerOption) -> None:
        self.name = name
        self.base_url = base_url
        self.timeout = 30
        self.max_retries = 3

        for option in options:
            option(self)


def with_timeout(timeout: int) -> ServerOption:
    def option(server: Server) -> None:
        server.timeout = timeout

    return option


server = Server("payment", "https://api.example.com", with_timeout(60))
```

## 库使用

本规范适用于基础设施级能力，例如日志、配置、链路追踪、错误处理、AWS 集成、worker pool、参数校验、分页和 JWT 等。

引入基础设施能力前，必须先检查 `uqpysdk`。

- `uqpysdk` 已提供能力时，直接使用 SDK 包，不再封装一层。
- `uqpysdk` 未提供能力时，直接使用社区标准库或事实标准库，不在业务项目里新增本地 utility wrapper。
- 如果缺失能力属于可复用基础设施能力，应补到 `uqpysdk`，而不是沉淀在单个业务项目中。
