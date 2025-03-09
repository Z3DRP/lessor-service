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

	Id                int64           `bun:"column:id,pk,autoincrement" json:"-"`
	Pid               uuid.UUID       `bun:"type:uuid,notnull,unique" json:"pid"`
	Property          *Property       `bun:"rel:belongs-to,join:pid=pid" json:"property"`
	RentalPrice       decimal.Decimal `bun:"type:money,notnull" json:"rentalPrice"`
	RentDueDate       time.Time       `bun:"type:timestamptz,nullzero" json:"rentDueDate"`
	LeaseSigned       bool            `bun:"type:boolean,nullzero,notnull,default:false" json:"leaseSigned"`
	LeaseDuration     int             `bun:",nullzero" json:"leaseDuration"`
	LeaseRenewDate    time.Time       `bun:"type:timestamptz,nullzero" json:"leaseRenewDate"`
	IsVacant          bool            `bun:"type:boolean,notnull,nullzero,default:false" json:"isVacant"`
	PetFriendly       bool            `bun:"type:boolean,nullzero,notnull,default:false" json:"petFriendly"`
	NeedsEviction     bool            `bun:"type:boolean,nullzero,notnull,default:false" json:"needsEviction"`
	EvictionStartDate time.Time       `bun:"type:timestamptz,nullzero,notnull,default:false" json:"evictionStartDate"`
}

func (r RentalProperty) Info() string {
	return fmt.Sprintf("%#v\n", r)
}
