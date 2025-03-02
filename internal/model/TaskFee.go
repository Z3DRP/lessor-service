package model

import "github.com/google/uuid"

type TaskFee struct {
	Id        int64     `bun:"column:id,pk,autoincrement"`
	Task_id   uuid.UUID `bun:"type:uuid,notnull,unique"`
	Task      *Task     `bun:"rel:belongs-to,join:task_id=tid"`
	Material  string    `bun:"type:varchar(255)"`
	Cost      float64   `bun:"type:money,notnull,nullzero"`
	Details   string    `bun:"type:varchar(255)"`
	Processed bool      `bun:"type:boolean,notnull,nullzero,default:false"`
}
