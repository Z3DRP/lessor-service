package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type SaleProperty struct {
	bun.BaseModel `bun:"table:sale_properties,alias:sp"`

	Id                int64           `bun:"column:id,pk,autoincrement"`
	Pid               uuid.UUID       `bun:"type:uuid,notnull,unique"`
	Property          *Property       `bun:"rel:belongs-to,join:pid=pid"`
	ListingPrice      decimal.Decimal `bun:"type:money,nullzero"`
	AppraisedValue    decimal.Decimal `bun:"type:money,nullzero"`
	OfferPrice        decimal.Decimal `bun:"type:money,nullzero"`
	FinalPrice        decimal.Decimal `bun:"type:money,nullzero"`
	SoldOn            time.Time       `bun:"type:timestamptz,nullzero"`
	NeedsEviction     bool            `bun:"type:boolean,nullzero,notnull,default:false"`
	EvictionStartDate time.Time       `bun:"type:timestamptz,nullzero"`
}

func (s SaleProperty) Info() string {
	return fmt.Sprintf("%#v\n", s)
}
