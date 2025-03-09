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

	Id            int64           `bun:"column:id,pk,autoincrement" json:"-"`
	Pid           uuid.UUID       `bun:"type:uuid,notnull,unique" json:"pid"`
	LessorId      uuid.UUID       `bun:"type:uuid,notnull" json:"alessorId"`
	Alessor       *Alessor        `bun:"rel:belongs-to,join:lessor_id=uid" json:"alessor"`
	Address       json.RawMessage `bun:"type:jsonb,json_use_number" json:"address"`
	Bedrooms      float64         `bun:"type:numeric(5,2),notnull,nullzero,default:0" json:"bedrooms"`
	Baths         float64         `bun:"type:numeric(5,2),notnull,nullzero,default:0" json:"baths"`
	SquareFootage float64         `bun:"type:numeric(10,4),nullzero" json:"squareFootage"`
	IsAvailable   bool            `bun:",nullzero" json:"isAvailable"`
	Status        PropertyStatus  `bun:"type:property_status,default:'unknown'" json:"status"`
	Notes         string          `bun:"type:text,nullzero" json:"notes"`
	Image         string          `bun:"type:varchar(255),nullzero" json:"image"`
	TaxRate       float64         `bun:"type:numeric(10,4),nullzero" json:"taxRate"`
	TaxAmountDue  float64         `bun:"type:numeric(10,2)" json:"taxAmountDue"`
	MaxOccupancy  int             `bun:",nullzero" json:"maxOccupancy"`
}

func (p Property) Info() string {
	return fmt.Sprintf("%#v\n", p)
}
