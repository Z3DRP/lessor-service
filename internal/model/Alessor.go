package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CommunicationPreference string

const (
	Email CommunicationPreference = "email"
	Phone CommunicationPreference = "phone"
	Text  CommunicationPreference = "text"
)

type Alessor struct {
	bun.BaseModel `bun:"table:alessors,alias:alsr"`

	Id                        int64                   `bun:"column:id,pk,autoincrement"`
	Uid                       uuid.UUID               `bun:"type:uuid,unique,notnull"`
	User                      *User                   `bun:"rel:belongs-to,join:uid=uid"`
	Bid                       uuid.UUID               `bun:"type:uuid,unique,notnull"`
	TotalProperties           int64                   `bun:"type:int,notnull,nullzero,default:0"`
	SquareAccount             string                  `bun:"type:varchar(255),nullzero"`
	PaymentIntegrationEnabled bool                    `bun:"type:boolean,notnull,nullzero,default:false"`
	PaymentSchedule           map[string]interface{}  `bun:"type:jsonb,json_use_number"`
	ComunicationPreference    CommunicationPreference `bun:"type:communication_preference"`
	NumberOfEmployees         int                     `bun:"type:int"`
}

func (a Alessor) Info() string {
	return fmt.Sprintf("%#v\n", a)
}
