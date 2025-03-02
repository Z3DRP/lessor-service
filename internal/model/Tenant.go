package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Tenant struct {
	bun.BaseModel `bun:"table:tenants,alias:tnt"`

	Id          int64     `bun:"column:id,pk,autoincrement"`
	Uid         uuid.UUID `bun:"type:uuid,notnull,unique"`
	User        *User     `bun:"rel:belongs-to,join:uid=uid"`
	LessorId    uuid.UUID `bun:"type:uuid,notnull,unique"`
	Lessor      *User     `bun:"rel:belongs-to,join:lessor_id=uid"`
	MoveInDate  time.Time `bun:"type:timestamptz,nullzero"`
	MoveOutDate time.Time `bun:"type:timestamptz,nullzero"`
	PropertyId  uuid.UUID `bun:"type:uuid,notnull"`
	Property    *Property `bun:"rel:belongs-to,join:property_id=pid"`
}

func (t Tenant) Info() string {
	return fmt.Sprintf("%#v\n", t)
}
