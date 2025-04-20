package dtos

import (
	"errors"
	"time"

	"github.com/Z3DRP/lessor-service/internal/model"
)

type NotificationDto struct {
	Id         int             `json:"id"`
	Title      string          `json:"title"`
	Message    string          `json:"message"`
	LessorId   string          `json:"lessorId"`
	TaskId     string          `json:"taskId"`
	Task       *model.Task     `json:"task"`
	UserId     string          `json:"userId"`
	User       *model.User     `json:"user"`
	PropertyId string          `json:"propertyId"`
	Property   *model.Property `json:"property"`
	Category   string          `json:"category"`
	Viewed     bool            `json:"viewed"`
	CreatedAt  time.Time       `json:"createdAt"`
	VoidAt     time.Time       `json:"voidAt"`
}

func (n NotificationDto) Validate() error {
	if n.Title == "" {
		return errors.New("error missing title")
	}

	if n.Message == "" {
		return errors.New("error missing message")
	}

	if n.Category == "" {
		return errors.New("error missing category")
	}

	if n.Category != "property" && n.Category != "task" && n.Category != "user" && n.Category != "wo	rker" && n.Category != "tenant" && n.Category != "general" {
		return errors.New("error invalid category")
	}
	return nil
}

func NewNotificationDto(n model.Notification) *NotificationDto {
	return &NotificationDto{
		Id:         n.Id,
		Title:      n.Title,
		Message:    n.Message,
		LessorId:   n.LessorId.String(),
		TaskId:     n.TaskId.String(),
		Task:       n.Task,
		UserId:     n.UserId.String(),
		User:       n.User,
		PropertyId: n.PropertyId.String(),
		Property:   n.Property,
		Category:   string(n.Category),
		Viewed:     n.Viewed,
		CreatedAt:  n.CreatedAt,
		VoidAt:     n.VoidAt,
	}
}

func NewNotificationDtoList(n []model.Notification) []*NotificationDto {
	responses := make([]*NotificationDto, len(n))
	for _, noti := range n {
		responses = append(responses, NewNotificationDto(noti))
	}
	return responses
}
