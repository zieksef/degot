# Go ORM 与数据库规范

Go 数据库模型默认使用 GORM。设计表结构、模型字段、时间字段和软删除时，优先使用 GORM 的约定能力，不手写重复逻辑。

## 命名

| 对象 | 规则 | 示例 |
| --- | --- | --- |
| Table | `snake_case`，单数名词 | `payment`, `merchant`, `card` |
| Column | `snake_case` | `account_id`, `created_at` |
| Index | `idx_{table}_{columns}` | `idx_payment_merchant_id` |
| Unique index | `uk_{table}_{columns}` | `uk_card_card_number` |

## 必需字段

每张表必须包含内部主键和时间字段：

```sql
id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY
created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
updated_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
```

`updated_at` 由数据库 `ON UPDATE` 与 GORM `autoUpdateTime` 共同维护，两者不冲突：GORM 写入时会显式赋值并以其为准，`ON UPDATE` 作为非 GORM 写入（原生 SQL、其他服务、手工改库）的兜底。

## 软删除

需要软删除的表必须添加 `deleted_at`：

```sql
deleted_at  DATETIME(3) NULL DEFAULT NULL
```

业务字段默认 `NOT NULL` 并给默认值，避免 `NULL` 带来的三值逻辑问题。允许为 `NULL` 的情况有两类：`deleted_at`（`NULL` 表示未删除），以及事件类时间戳（如 `paid_at`、`refunded_at`、`cancelled_at`，`NULL` 表示该事件尚未发生）。

GORM 模型使用 `gorm.DeletedAt`：

```go
DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
```

## 主键

- 内部主键使用 `BIGINT UNSIGNED AUTO_INCREMENT`。
- 对外暴露必须使用独立业务 ID 字段，例如 `payment_id VARCHAR(36)`。

## 枚举约定

数值枚举建议从 3～4 位数值开始（如 `100`、`1000`），便于按状态段分组；避免用 0 作为有效值，防止与 Go 零值混淆，也不使用负数。

```go
// Bad
const (
	StatusPending = 0
	StatusSuccess = 1
)

// Good
const (
	StatusPending    = 100
	StatusProcessing = 101
	StatusSuccess    = 200
	StatusFailed     = 300
)
```

数值枚举列的类型必须能容纳取值范围（按 3～4 位取值使用 `SMALLINT UNSIGNED` 或更宽），列默认值必须是合法枚举值。

字符串枚举全部小写，不使用混合大小写。

```go
// Bad
const (
	CardSchemaVISA = "VISA"
)

// Good
const (
	CardSchemaVisa = "visa"
)
```

## 金额

金额使用 `DECIMAL`（GORM 模型用 `decimal.Decimal`），不使用浮点类型。

小数位按币种类型确定：法币保留 4 位，加密货币保留 8 位。

## GORM 约定

GORM 模型每个持久化字段都必须在 tag 中显式声明 `column`，不依赖默认命名推断。

```go
type Payment struct {
	ID        uint64          `gorm:"column:id;primaryKey;autoIncrement" json:"-"`
	PaymentID string          `gorm:"column:payment_id;type:varchar(36);uniqueIndex" json:"payment_id"`
	Amount    decimal.Decimal `gorm:"column:amount;type:decimal(20,8);not null;default:0" json:"amount"`
	CreatedAt time.Time       `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"column:deleted_at" json:"-"`
}
```

## SQL 查询规范

查询默认使用 GORM 方法；只有查询过于复杂、无法清晰表达为 GORM 链式调用时，才允许使用原生 SQL。

- 单次查询最多关联 3 张表；超过时重新设计查询或拆分成多次查询。
- 字段数超过 20 个的表或存在 text 类型的字段，才考虑显式 `Select` 必要字段；实施前必须询问用户是否需要。确需 `Select` 子集时，扫描进专门的投影结构体，不要用完整领域模型接收，避免未选中字段变成零值被下游误用。
- 简单等值条件优先使用 struct 字段，减少裸字段名字符串。
- 写入操作使用 GORM 的 `Create`、`Save` 或明确的 update 方法。
- 不在循环中查询数据库，改用批量查询。
- 不在循环中逐条插入或更新，改用批量操作。
- 查询参数不超过 3 个时直接传参；超过 3 个时使用 query struct。

## Repository 与事务

Repository 持有 `*gorm.DB` 并实现自己的接口；接口提供必需 `WithTx` 方法，返回绑定到事务句柄的同类型 Repository。

- Repository 方法第一个参数是 `ctx context.Context`，其后是业务参数；查询时用 `WithContext(ctx)` 传给 GORM，不在方法签名上逐个传 `tx`。
- Service 在初始化时注入 `*gorm.DB`（仅用于开启事务，查询仍走 Repository）；用 `db.Transaction` 驱动事务，用 `WithTx(tx)` 把各 Repository 绑到同一事务。
- 事务内必须使用 `WithTx(tx)` 绑定后的 Repository；直接用原 Repository 会走事务外的连接，提交回滚都管不到它。

```go
type PaymentRepository interface {
	GetByID(ctx context.Context, id string) (*Payment, error)
	Create(ctx context.Context, p *Payment) error
	WithTx(tx *gorm.DB) PaymentRepository
}

type paymentRepo struct{ db *gorm.DB }

func NewPaymentRepo(db *gorm.DB) PaymentRepository { return &paymentRepo{db: db} }

func (r *paymentRepo) WithTx(tx *gorm.DB) PaymentRepository { return &paymentRepo{db: tx} }
```

```go
// Service 初始化时注入 *gorm.DB，仅用于开启事务。
type PaymentService struct {
	db          *gorm.DB
	paymentRepo PaymentRepository
	ledgerRepo  LedgerRepository
}

func (s *PaymentService) Settle(ctx context.Context, p *Payment, e *LedgerEntry) error {
      tx := s.db.WithContext(ctx).Begin()
      if tx.Error != nil {
              return tx.Error
      }
      defer tx.Rollback() // 关键：任何提前 return 或 panic 都会回滚；commit 成功后它是安全空操作

      if err := s.paymentRepo.WithTx(tx).Create(ctx, p); err != nil {
              return err // defer 的 rollback 生效
      }
      if err := s.ledgerRepo.WithTx(tx).Append(ctx, e); err != nil {
              return err
      }

      return tx.Commit().Error
}
```

## DDL 示例

```sql
CREATE TABLE payment (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'Auto-increment PK',
    payment_id  VARCHAR(36)  NOT NULL DEFAULT '' COMMENT 'External payment ID',
    merchant_id VARCHAR(36)  NOT NULL DEFAULT '' COMMENT 'Merchant ID',
    amount      DECIMAL(20,8) NOT NULL DEFAULT 0 COMMENT 'Amount (scale by currency)',
    currency    VARCHAR(3)   NOT NULL DEFAULT '' COMMENT 'Currency ISO 4217',
    status      SMALLINT UNSIGNED NOT NULL DEFAULT 100 COMMENT 'Status 100:pending 200:success 300:failed',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Created time',
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Updated time',
    UNIQUE KEY uk_payment_payment_id (payment_id),
    INDEX idx_payment_merchant_id (merchant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Payment order table';
```
