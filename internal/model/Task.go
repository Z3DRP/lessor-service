package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Task struct {
	bun.BaseModel `bun:"table:tasks,alias:tsk"`

	Id           int64     `bun:"column:id,pk,autoincrement"`
	Tid          uuid.UUID `bun:"type:uuid,notnull,unique"`
	OwnerId      uuid.UUID `bun:"type:uuid,notnull"`
	Alessor      *Alessor  `bun:"rel:belongs-to,join:owner_id=bid"`
	Details      string    `bun:"type:varchar(255),notnull"`
	Notes        string    `bun:"type:varchar(255)"`
	PropertyId   uuid.UUID `bun:"type:uuid,notnull"`
	Property     *Property `bun:"rel:belongs-to,join:property_id=pid"`
	ScheduledAt  time.Time `bun:"type:timestamptz,nullzero"`
	StartedAt    time.Time `bun:"type:timestamptz,nullzero"`
	CompletedAt  time.Time `bun:"type:timestamptz,nullzero"`
	PausedAt     time.Time `bun:"type:timestamptz,nullzero"`
	PausedReason string    `bun:"type:varchar(255)"`
	FailedAt     time.Time `bun:"type:timestamptz,nullzero"`
	FailedReason string    `bun:"type:varchar(255)"`
}

func (t Task) Info() string {
	return fmt.Sprintf("%#v\n", t)
}
