package dtos

import (
	"time"

	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/shopspring/decimal"
)

type WorkerDto struct {
	Uid           string          `json:"uid"`
	User          *model.User     `json:"user"`
	StartDate     time.Time       `json:"startDate"`
	EndDate       time.Time       `json:"endDate"`
	Title         string          `json:"title"`
	Specilization string          `json:"specilization"`
	PayRate       decimal.Decimal `json:"payRate"`
	LessorId      string          `json:"lessorId"`
	PaymentMethod string          `json:"paymentMethod"`
	Image         string          `json:"image"`
	ImageUrl      string          `json:"imageUrl"`
}

func (w WorkerDto) Validate() error {
	return nil
}

func NewWorkerDto(w model.Worker, url *string) WorkerDto {
	return WorkerDto{
		Uid:           w.Uid.String(),
		User:          w.User,
		StartDate:     w.StartDate,
		EndDate:       w.EndDate,
		Title:         w.Title,
		Specilization: w.Specilization,
		PayRate:       w.PayRate,
		LessorId:      w.LessorId.String(),
		PaymentMethod: string(w.PaymentMethod),
		Image:         w.Image,
	}
}

func NewWorkerDtoFrmPtr(w *model.Worker, url *string) WorkerDto {
	return WorkerDto{
		Uid:           w.Uid.String(),
		User:          w.User,
		StartDate:     w.StartDate,
		EndDate:       w.EndDate,
		Title:         w.Title,
		Specilization: w.Specilization,
		PayRate:       w.PayRate,
		LessorId:      w.LessorId.String(),
		PaymentMethod: string(w.PaymentMethod),
		Image:         w.Image,
	}
}
