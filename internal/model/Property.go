package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PropertyStatus string

const (
	Pending    PropertyStatus = "pending"
	InProgress PropertyStatus = "in-progress"
	Completed  PropertyStatus = "completed"
	Unknown    PropertyStatus = "unknown"
)

type Property struct {
	bun.BaseModel `bun:"table:properties,alias:p"`

	Id            int64                  `bun:"column:id,pk,autoincrement"`
	Pid           uuid.UUID              `bun:"type:uuid,notnull,unique"`
	AlessorId     uuid.UUID              `bun:"type:uuid,notnull"`
	Alessor       *Alessor               `bun:"rel:belongs-to,join:alessor_id=bid"`
	Address       map[string]interface{} `bun:"type:jsonb,json_use_number"`
	Bedrooms      int                    `bun:",notnull,nullzero,default:0"`
	Baths         int                    `bun:",notnull,nullzero,default:0"`
	SquareFootage float64                `bun:"type:numeric(10,2),nullzero"`
	IsAvailable   bool                   `bun:",nullzero"`
	Status        PropertyStatus         `bun:"type:property_status,default:'unknown'"`
	Notes         string                 `bun:"type:varchar(255),nullzero"`
	Image         string                 `bun:"type:varchar(255),nullzero"`
	TaxRate       float64                `bun:"type:numeric(5,2),nullzero"`
	TaxAmountDue  float64                `bun:"type:numeric(10,2)"`
	MaxOccupancy  int                    `bun:",nullzero"`
}

func (p Property) Info() string {
	return fmt.Sprintf("%#v\n", p)
}
