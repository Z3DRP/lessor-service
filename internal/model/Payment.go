package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type PaymentStatus string

const (
	Rejected PaymentStatus = "rejected"
	Accepted PaymentStatus = "accepted"
)

type Payment struct {
	bun.BaseModel `bun:"table:payments,alias:pmts"`

	Id                int64           `bun:"column:id,pk,autoincrement"`
	TxId              uuid.UUID       `bun:"column_name:txid,type:uuid,notnull,unique"`
	SqPid             string          `bun:"column_name:sqpid,type:varchar(50),nullzero"`
	Amount            decimal.Decimal `bun:"type:money,notnull,nullzero"`
	CurrenyCode       string          `bun:"type:char(3),nullzero"`
	TransactionStatus PaymentStatus   `bun:"type:payment_status,notnull"`
	Note              string          `bun:"type:varchar(50)"`
	ReceitNumber      string          `bun:"type:varchar(100)"`
	TenantId          uuid.UUID       `bun:"type:uuid,notnull"`
	Tenant            *Tenant         `bun:"rel:belongs-to,join:tenant_id=uid"`
	CreatedAt         time.Time       `bun:"type:timestamptz,notnull,nullzero,default:current_timestamp"`
}

func (p Payment) Info() string {
	return fmt.Sprintf("%#v\n", p)
}
