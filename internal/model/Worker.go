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

	Id            int64           `bun:"column:id,pk,autoincrement" json:"-"`
	Uid           uuid.UUID       `bun:"type:uuid,notnull,unique" json:"uid"`
	User          *User           `bun:"rel:belongs-to,join:uid=uid" json:"user"`
	StartDate     time.Time       `bun:"type:timestamptz,nullzero" json:"startDate"`
	EndDate       time.Time       `bun:"type:timestamptz,nullzero" json:"endDate"`
	Title         string          `bun:"type:varchar(100),nullzero" json:"title"`
	Specilization string          `bun:"type:varchar(255),nullzero" json:"specilization"`
	PayRate       decimal.Decimal `bun:"type:money,nullzero" json:"payRate"`
	LessorId      uuid.UUID       `bun:"type:uuid,notnull,unique" json:"lessorId"`
	Alessor       *Alessor        `bun:"rel:belongs-to,join:lessor_id=uid" json:"alessor"`
	PaymentMethod MethodOfPayment `bun:"type:method_of_payment" json:"-"`
	Image         string          `bun:"type:text" json:"iamge"`
}

func (w Worker) Info() string {
	return fmt.Sprintf("%#v\n", w)
}
