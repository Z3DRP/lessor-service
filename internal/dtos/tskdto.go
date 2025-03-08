package dtos

import (
	"time"

	"github.com/Z3DRP/lessor-service/internal/model"
)

type TaskResponse struct {
	Tid          string        `json:"tid"`
	LessorId     string        `json:"alessorId"`
	Details      string        `json:"details"`
	Notes        string        `json:"notes"`
	PropertyId   string        `json:"propertyId"`
	ScheduledAt  time.Time     `json:"scheduledAt"`
	StartedAt    time.Time     `json:"startedAt"`
	CompletedAt  time.Time     `json:"completedAt"`
	PausedAt     time.Time     `json:"pausedAt"`
	PausedReason string        `json:"pausedReason"`
	FailedAt     time.Time     `json:"failedAt"`
	FailedReason string        `json:"failedReason"`
	WorkerId     string        `json:"workerId"`
	Worker       *model.Worker `json:"worker"`
	Image        string        `json:"image"`
	ImageUrl     *string       `json:"imageUrl"`
}

func (t TaskResponse) Validate() error {
	return nil
}

func NewTaskResposne(t model.Task, url *string) TaskResponse {
	return TaskResponse{
		Tid:          t.Tid.String(),
		LessorId:     t.LessorId.String(),
		Details:      t.Details,
		Notes:        t.Notes,
		PropertyId:   t.PropertyId.String(),
		ScheduledAt:  t.ScheduledAt,
		StartedAt:    t.StartedAt,
		CompletedAt:  t.CompletedAt,
		PausedAt:     t.PausedAt,
		PausedReason: t.PausedReason,
		FailedAt:     t.FailedAt,
		FailedReason: t.FailedReason,
		WorkerId:     t.WorkerId.String(),
		Worker:       t.Worker,
		Image:        t.Image,
		ImageUrl:     url,
	}
}

func NewTaskResposneFrmPntr(t *model.Task, url *string) TaskResponse {
	return TaskResponse{
		Tid:          t.Tid.String(),
		LessorId:     t.LessorId.String(),
		Details:      t.Details,
		Notes:        t.Notes,
		PropertyId:   t.PropertyId.String(),
		ScheduledAt:  t.ScheduledAt,
		StartedAt:    t.StartedAt,
		CompletedAt:  t.CompletedAt,
		PausedAt:     t.PausedAt,
		PausedReason: t.PausedReason,
		FailedAt:     t.FailedAt,
		FailedReason: t.FailedReason,
		WorkerId:     t.WorkerId.String(),
		Worker:       t.Worker,
		Image:        t.Image,
		ImageUrl:     url,
	}
}

type TaskRequest struct {
	Tid         string    `json:"tid"`
	LessorId    string    `json:"alessorId"`
	Details     string    `json:"details"`
	Notes       string    `json:"notes"`
	PropertyId  string    `json:"propertyId"`
	ScheduledAt time.Time `json:"scheduledAt"`
	WorkerId    string    `json:"workerId"`
	Image       string    `json:"image"`
}

func (t TaskRequest) Validate() error {
	return nil
}

type TaskModRequest struct {
	Tid          string    `json:"tid"`
	LessorId     string    `json:"alessorId"`
	Details      string    `json:"details"`
	Notes        string    `json:"notes"`
	PropertyId   string    `json:"propertyId"`
	ScheduledAt  time.Time `json:"scheduledAt"`
	StartedAt    time.Time `json:"startedAt"`
	CompletedAt  time.Time `json:"completedAt"`
	PausedAt     time.Time `json:"pausedAt"`
	PausedReason string    `json:"pausedReason"`
	FailedAt     time.Time `json:"failedAt"`
	FailedReason string    `json:"failedReason"`
	WorkerId     string    `json:"workerId"`
	Image        string    `json:"image"`
}

func (t TaskModRequest) Validate() error {
	return nil
}
