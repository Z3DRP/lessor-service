package dtos

import (
	"time"

	"github.com/Z3DRP/lessor-service/internal/model"
)

type TaskResponse struct {
	Tid           string          `json:"tid"`
	Name          string          `json:"name"`
	LessorId      string          `json:"lessorId"`
	Details       string          `json:"details"`
	Notes         string          `json:"notes"`
	PropertyId    string          `json:"propertyId"`
	Property      *model.Property `json:"property"`
	Category      string          `json:"category"`
	ScheduledAt   time.Time       `json:"scheduledAt"`
	StartedAt     time.Time       `json:"startedAt"`
	CompletedAt   time.Time       `json:"completedAt"`
	PausedAt      time.Time       `json:"pausedAt"`
	PausedReason  string          `json:"pausedReason"`
	FailedAt      time.Time       `json:"failedAt"`
	FailedReason  string          `json:"failedReason"`
	WorkerId      string          `json:"workerId"`
	Worker        *model.Worker   `json:"worker"`
	EstimatedCost float64         `json:"estimatedCost"`
	ActualCost    float64         `json:"actualCost"`
	Profit        float64         `json:"profit"`
	Priority      string          `json:"priority"`
	Image         string          `json:"image"`
	ImageUrl      *string         `json:"imageUrl"`
}

func (t TaskResponse) Validate() error {
	return nil
}

func NewTaskResposne(t *model.Task, url *string) TaskResponse {
	return TaskResponse{
		Tid:           t.Tid.String(),
		Name:          t.Name,
		LessorId:      t.LessorId.String(),
		Details:       t.Details,
		Notes:         t.Notes,
		PropertyId:    t.PropertyId.String(),
		Property:      t.Property,
		Category:      string(t.Category),
		ScheduledAt:   t.ScheduledAt,
		StartedAt:     t.StartedAt,
		CompletedAt:   t.CompletedAt,
		PausedAt:      t.PausedAt,
		PausedReason:  t.PausedReason,
		FailedAt:      t.FailedAt,
		FailedReason:  t.FailedReason,
		WorkerId:      t.WorkerId.String(),
		Worker:        t.Worker,
		EstimatedCost: t.EstimatedCost,
		ActualCost:    t.ActualCost,
		Priority:      string(t.Priority),
		Profit:        t.Profit,
		Image:         t.Image,
		ImageUrl:      url,
	}
}

func NewTaskResponseList(t []model.Task) []TaskResponse {
	response := make([]TaskResponse, len(t))
	for _, tsk := range t {
		response = append(response, NewTaskResposne(&tsk, nil))
	}

	return response
}

func NewTaskResposneFrmPntr(t *model.Task, url *string) TaskResponse {
	return TaskResponse{
		Tid:           t.Tid.String(),
		LessorId:      t.LessorId.String(),
		Name:          t.Name,
		Details:       t.Details,
		Notes:         t.Notes,
		PropertyId:    t.PropertyId.String(),
		Category:      string(t.Category),
		ScheduledAt:   t.ScheduledAt,
		StartedAt:     t.StartedAt,
		CompletedAt:   t.CompletedAt,
		PausedAt:      t.PausedAt,
		PausedReason:  t.PausedReason,
		FailedAt:      t.FailedAt,
		FailedReason:  t.FailedReason,
		WorkerId:      t.WorkerId.String(),
		Worker:        t.Worker,
		EstimatedCost: t.EstimatedCost,
		ActualCost:    t.ActualCost,
		Priority:      string(t.Priority),
		Profit:        t.Profit,
		Image:         t.Image,
		ImageUrl:      url,
	}
}

type TaskRequest struct {
	Tid          string    `json:"tid"`
	LessorId     string    `json:"lessorId"`
	Name         string    `json:"name"`
	Details      string    `json:"details"`
	Notes        string    `json:"notes"`
	PropertyId   string    `json:"propertyId"`
	Category     string    `json:"category"`
	ScheduledAt  time.Time `json:"scheduledAt"`
	WorkerId     string    `json:"workerId"`
	EstimateCost float64   `json:"estimatedCost"`
	ActualCost   float64   `json:"actualCost"`
	Profit       float64   `json:"profit"`
	Image        string    `json:"image"`
	Priority     string    `json:"priority"`
}

func (t TaskRequest) Validate() error {
	return nil
}

type TaskModRequest struct {
	Tid           string    `json:"tid"`
	LessorId      string    `json:"alessorId"`
	Name          string    `json:"name"`
	Details       string    `json:"details"`
	Notes         string    `json:"notes"`
	PropertyId    string    `json:"propertyId"`
	Category      string    `json:"category"`
	ScheduledAt   time.Time `json:"scheduledAt"`
	StartedAt     time.Time `json:"startedAt"`
	CompletedAt   time.Time `json:"completedAt"`
	PausedAt      time.Time `json:"pausedAt"`
	PausedReason  string    `json:"pausedReason"`
	FailedAt      time.Time `json:"failedAt"`
	FailedReason  string    `json:"failedReason"`
	WorkerId      string    `json:"workerId"`
	EstimatedCost float64   `json:"estimatedCost"`
	ActualCost    float64   `json:"actualCost"`
	Profit        float64   `json:"profit"`
	Image         string    `json:"image"`
	Priority      string    `json:"priority"`
}

func (t TaskModRequest) Validate() error {
	return nil
}
