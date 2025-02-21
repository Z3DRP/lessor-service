package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type NotificationSettings struct {
	bun.BaseModel `bun:"table:notification_settings,alias:ns"`

	Id      int64                  `bun:"column:id,pk,autoincrement"`
	Uid     uuid.UUID              `bun:"type:uuid,notnull,unique"`
	User    *User                  `bun:"rel:belongs-to,join:uid=uid"`
	Setting map[string]interface{} `bun:"type:jsonb,json_use_number"`
}

func (n NotificationSettings) Info() string {
	return fmt.Sprintf("%#v\n", n)
}
