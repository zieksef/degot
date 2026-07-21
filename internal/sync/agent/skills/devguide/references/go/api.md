# Go API 规范

## DTO 与校验

请求结构体必须通过 struct tag 声明校验规则，避免手写 if-else 校验逻辑。

```go
type CreateCardReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Schema    string `json:"schema" validate:"required,oneof=visa mastercard"`
}
```

Request 和 response 的 JSON tag 都必须使用 `snake_case`。

金额字段使用 `decimal.Decimal`（序列化为 JSON 字符串），不使用 `float`，避免精度丢失。

## 客户端可见错误

所有客户端可见错误响应必须使用 `git.uqpaytech.com/sdk-group/uqpay-core-sdk/pkg/apierror`。

内部错误仍按错误封装规范传递；只有在需要返回给客户端的 API 边界，才转换为 `apierror` 响应。

```json
{
  "type": "invalid_request_error",
  "code": "invalid_parameter",
  "message": "Your request contains a parameter that doesn't look right."
}
```
