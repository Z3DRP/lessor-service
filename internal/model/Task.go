package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PriorityLevel string

const (
	Low       PriorityLevel = "low"
	Medium    PriorityLevel = "medium"
	High      PriorityLevel = "high"
	Immediate PriorityLevel = "immediate"
)

type Task struct {
	bun.BaseModel `bun:"table:tasks,alias:tsk"`

	Id             int64         `bun:"column:id,pk,autoincrement" json:"-"`
	Tid            uuid.UUID     `bun:"type:uuid,notnull,unique" json:"tid"`
	LessorId       uuid.UUID     `bun:"type:uuid,notnull,unique" json:"lessorId"`
	Alessor        *Alessor      `bun:"rel:belongs-to,join:lessor_id=uid" json:"alessor"`
	Details        string        `bun:"type:varchar(255),notnull" json:"details"`
	Notes          string        `bun:"type:varchar(255)" json:"notes"`
	Priority       PriorityLevel `bun:"type:priority_level,notnull" json:"prioriy"`
	TakePrecedence bool          `bun:"type:bool" json:"takePrecedence"`
	PropertyId     uuid.UUID     `bun:"type:uuid,notnull" json:"propertyId"`
	Property       *Property     `bun:"rel:belongs-to,join:property_id=pid" json:"property"`
	ScheduledAt    time.Time     `bun:"type:timestamptz,nullzero" json:"scheduledAt"`
	StartedAt      time.Time     `bun:"type:timestamptz,nullzero" json:"startedAt"`
	CompletedAt    time.Time     `bun:"type:timestamptz,nullzero" json:"completedAt"`
	PausedAt       time.Time     `bun:"type:timestamptz,nullzero" json:"pausedAt"`
	PausedReason   string        `bun:"type:varchar(255)" json:"pausedReason"`
	FailedAt       time.Time     `bun:"type:timestamptz,nullzero" json:"failedAt"`
	FailedReason   string        `bun:"type:varchar(255)" json:"failedReason"`
	WorkerId       uuid.UUID     `bun:"type:uuid,nullzero" json:"workerId"`
	Worker         *Worker       `bun:"rel:belongs-to,join:WorkerId=uid" json:"worker"`
	Image          string        `bun:"type:text,nullzero" json:"image"`
}

func (t Task) Info() string {
	return fmt.Sprintf("%#v\n", t)
}
