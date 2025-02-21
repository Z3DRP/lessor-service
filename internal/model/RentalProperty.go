package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type RentalProperty struct {
	bun.BaseModel `bun:"table:rental_properties,alias:rp"`

	Id                int64           `bun:"column:id,pk,autoincrement"`
	Pid               uuid.UUID       `bun:"type:uuid,notnull,unique"`
	Property          *Property       `bun:"rel:belongs-to,join:pid=pid"`
	RentalPrice       decimal.Decimal `bun:"type:money,notnull"`
	RentDueDate       time.Time       `bun:"type:timestamptz,nullzero"`
	LeaseSigned       bool            `bun:"type:boolean,nullzero,notnull,default:false"`
	LeaseDuration     int             `bun:",nullzero"`
	LeaseRenewDate    time.Time       `bun:"type:timestamptz,nullzero"`
	IsVacant          bool            `bun:"type:boolean,notnull,nullzero,default:false"`
	PetFriendly       bool            `bun:"type:boolean,nullzero,notnull,default:false"`
	NeedsEviction     bool            `bun:"type:boolean,nullzero,notnull,default:false"`
	EvictionStartDate bool            `bun:"type:boolean,nullzero,notnull,default:false"`
}

func (r RentalProperty) Info() string {
	return fmt.Sprintf("%#v\n", r)
}
