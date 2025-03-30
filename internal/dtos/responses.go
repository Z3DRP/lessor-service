package dtos

import (
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/google/uuid"
)

type SigninResponse struct {
	Uid         uuid.UUID `json:"uid"`
	LessorId    uuid.UUID `json:"lessorId"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	ProfileType string    `json:"profileType"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	IsActive    bool      `json:"isActive"`
	Phone       string    `json:"phone"`
}

func NewSigninResponse(usr *model.User) SigninResponse {
	return SigninResponse{
		Uid:         usr.Uid,
		Username:    usr.Username,
		Email:       usr.Email,
		ProfileType: usr.ProfileType,
		FirstName:   usr.FirstName,
		LastName:    usr.LastName,
		IsActive:    usr.IsActive,
		Phone:       usr.Phone,
	}
}

func NewWorkerSignUpResponse(usr *model.User, lessorId uuid.UUID) SigninResponse {
	return SigninResponse{
		Uid:         usr.Uid,
		LessorId:    lessorId,
		Username:    usr.Username,
		Email:       usr.Email,
		ProfileType: usr.ProfileType,
		FirstName:   usr.FirstName,
		LastName:    usr.LastName,
		IsActive:    usr.IsActive,
		Phone:       usr.Phone,
	}
}
