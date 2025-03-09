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

	Id                int64           `bun:"column:id,pk,autoincrement" json:"-"`
	Pid               uuid.UUID       `bun:"type:uuid,notnull,unique" json:"pid"`
	Property          *Property       `bun:"rel:belongs-to,join:pid=pid" json:"property"`
	ListingPrice      decimal.Decimal `bun:"type:money,nullzero" json:"listingPrice"`
	AppraisedValue    decimal.Decimal `bun:"type:money,nullzero" json:"appraisedValue"`
	OfferPrice        decimal.Decimal `bun:"type:money,nullzero" json:"offerPrice"`
	FinalPrice        decimal.Decimal `bun:"type:money,nullzero" json:"finalPrice"`
	SoldOn            time.Time       `bun:"type:timestamptz,nullzero" json:"soldOn"`
	NeedsEviction     bool            `bun:"type:boolean,nullzero,notnull,default:false" json:"needsEviction"`
	EvictionStartDate time.Time       `bun:"type:timestamptz,nullzero" json:"evictionStarted"`
}

func (s SaleProperty) Info() string {
	return fmt.Sprintf("%#v\n", s)
}
