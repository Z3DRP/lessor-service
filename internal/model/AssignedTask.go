package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AssignedTask struct {
	bun.BaseModel `bun:"table:assigned_tasks,alias:at"`

	Id       int64     `bun:"column:id,pk,autoincrement"`
	TaskId   uuid.UUID `bun:"type:uuid,notnull"`
	Task     *Task     `bun:"rel:belongs-to,join:task_id=tid"`
	WorkerId uuid.UUID `bun:"type:uuid,notnull"`
	Worker   *Worker   `bun:"rel:belongs-to,join:worker_id=uid"`
}

func (a AssignedTask) Info() string {
	return fmt.Sprintf("%#v\n", a)
}
