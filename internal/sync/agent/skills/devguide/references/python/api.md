# Python API 规范

## DTO 与校验

请求和响应对象必须是 DTO，必须和业务领域实体分离。不要把 domain model 直接复用为 API request 或 response。

请求校验优先使用声明式 schema，例如 Pydantic、FastAPI schema 或项目既有 serializer，避免手写分散的 if-else 校验逻辑。

Request 和 response 的 JSON 字段必须使用 `snake_case`。

## 客户端可见错误

客户端可见错误必须通过项目统一错误响应模型返回；如果 `uqpysdk` 已提供对应能力，直接使用 `uqpysdk`。

内部错误仍按错误封装规范传递；只有在需要返回给客户端的 API 边界，才转换为客户端可见错误响应。
