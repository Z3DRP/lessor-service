package dtos

import (
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/google/uuid"
)

type SigninResponse struct {
	Uid         uuid.UUID `json:"uid"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	ProfileType string    `json:"profile_type"`
}

func NewSigninResponse(usr *model.User) SigninResponse {
	return SigninResponse{
		usr.Uid,
		usr.Username,
		usr.Email,
		usr.ProfileType,
	}
}
