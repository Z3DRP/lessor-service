package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type MethodOfPayment string

const (
	Check MethodOfPayment = "check"
	Cash  MethodOfPayment = "cash"
)

type Worker struct {
	bun.BaseModel `bun:"table:workers,alias:w"`

	Id            int64           `bun:"column:id,pk,autoincrement"`
	Uid           uuid.UUID       `bun:"type:uuid,notnull,unique"`
	User          *User           `bun:"rel:belongs-to,join:uid=uid"`
	StartDate     time.Time       `bun:"type:timestamptz,nullzero"`
	EndDate       time.Time       `bun:"type:timestamptz,nullzero"`
	Title         string          `bun:"type:varchar(100),nullzero"`
	Specilization string          `bun:"type:varchar(255),nullzero"`
	PayRate       decimal.Decimal `bun:"type:money,nullzero"`
	LessorId      uuid.UUID       `bun:"type:uuid,notnull,unique"`
	Alessor       *Alessor        `bun:"rel:belongs-to,join:lessor_id=uid"`
	PaymentMethod MethodOfPayment `bun:"type:method_of_payment"`
}

func (w Worker) Info() string {
	return fmt.Sprintf("%#v\n", w)
}
