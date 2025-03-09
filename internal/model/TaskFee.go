package model

import "github.com/google/uuid"

type TaskFee struct {
	Id        int64     `bun:"column:id,pk,autoincrement" json:"-"`
	TaskId    uuid.UUID `bun:"type:uuid,notnull,unique" json:"taskId"`
	Task      *Task     `bun:"rel:belongs-to,join:task_id=tid" json:"task"`
	Material  string    `bun:"type:varchar(255)" json:"material"`
	Cost      float64   `bun:"type:money,notnull,nullzero" json:"cost"`
	Details   string    `bun:"type:varchar(255)" json:"details"`
	Processed bool      `bun:"type:boolean,notnull,nullzero,default:false" json:"processed"`
}
