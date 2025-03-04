package model

import (
	"encoding/json"
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

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Zipcode string `json:"zipcode"`
}

type Property struct {
	bun.BaseModel `bun:"table:properties,alias:p"`

	Id            int64           `bun:"column:id,pk,autoincrement"`
	Pid           uuid.UUID       `bun:"type:uuid,notnull,unique"`
	LessorId      uuid.UUID       `bun:"type:uuid,notnull"`
	Alessor       *Alessor        `bun:"rel:belongs-to,join:lessor_id=uid"`
	Address       json.RawMessage `bun:"type:jsonb,json_use_number"`
	Bedrooms      float64         `bun:"type:numeric(5,2),notnull,nullzero,default:0"`
	Baths         float64         `bun:"type:numeric(5,2),notnull,nullzero,default:0"`
	SquareFootage float64         `bun:"type:numeric(10,4),nullzero"`
	IsAvailable   bool            `bun:",nullzero"`
	Status        PropertyStatus  `bun:"type:property_status,default:'unknown'"`
	Notes         string          `bun:"type:varchar(255),nullzero"`
	Image         string          `bun:"type:varchar(255),nullzero"`
	TaxRate       float64         `bun:"type:numeric(10,4),nullzero"`
	TaxAmountDue  float64         `bun:"type:numeric(10,2)"`
	MaxOccupancy  int             `bun:",nullzero"`
}

func (p Property) Info() string {
	return fmt.Sprintf("%#v\n", p)
}
