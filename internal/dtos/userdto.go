package dtos

import (
	"fmt"
	"time"

	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/utils"
)

var (
	minStrLen     = 2
	maxPwdLen     = 30
	minPwdLen     = 2
	maxFnameLen   = 30
	maxLnameLen   = 30
	maxPhneLen    = 10
	maxEmalLen    = 75
	maxUsrnameLen = 75
)

type UserSignupRequest struct {
	FirstName   string
	LastName    string
	ProfileType string
	Username    string
	Phone       string
	Email       string
	Password    string
}

func (u *UserSignupRequest) Validate() error {
	if u.ProfileType != "alessor" && u.ProfileType != "worker" {
		return fmt.Errorf("invalid profile type %v not supported", u.ProfileType)
	}
	return baseUserValidate(UserRequest{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Password:  u.Password,
		Phone:     u.Phone,
		Email:     u.Email,
	})
}

type UserSigninRequest struct {
	Uid         string
	FirstName   string
	LastName    string
	IsActive    bool
	Phone       string
	Email       string
	ProfileType string
	Username    string
}

func (u *UserSigninRequest) Validate() error {
	return nil
}

func NewSigninRequest(usr model.User) UserSigninRequest {
	return UserSigninRequest{
		Uid:         usr.Uid.String(), // alessor id
		FirstName:   usr.FirstName,
		LastName:    usr.LastName,
		IsActive:    usr.IsActive,
		Phone:       usr.Phone,
		Email:       usr.Email,
		ProfileType: usr.ProfileType,
		Username:    usr.Username,
	}
}

type UserRequest struct {
	Id         int64
	Uid        string
	FirstName  string
	LastName   string
	Email      string
	Phone      string
	Username   string
	Password   string
	IsActive   bool
	AvatarFile string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (u *UserRequest) Validate() error {
	return baseUserValidate(UserRequest{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Password:  u.Password,
		Phone:     u.Phone,
		Email:     u.Email,
	})
}

func baseUserValidate(u UserRequest) error {
	if utils.CharCount(u.FirstName) > maxFnameLen {
		return ErrMaxLength{Field: "first name", MaxLen: maxFnameLen}
	}
	if utils.CharCount(u.FirstName) < minStrLen {
		return ErrMinLength{Field: "first name", MinLen: maxFnameLen}
	}

	if utils.CharCount(u.LastName) > maxLnameLen {
		return ErrMaxLength{Field: "last name", MaxLen: maxFnameLen}
	}

	if utils.CharCount(u.LastName) < minStrLen {
		return ErrMinLength{Field: "last name", MinLen: maxLnameLen}
	}

	if utils.CharCount(u.Email) > maxEmalLen {
		return ErrMaxLength{Field: "email", MaxLen: maxFnameLen}
	}

	// if !utils.IsValidEmail(u.Email) {
	// 	return errors.New("invalid email address")
	// }

	if utils.CharCount(u.Phone) != maxPhneLen {
		return fmt.Errorf("invalid phone length, must be %v digits", maxPhneLen)
	}

	// if !utils.IsValidPhone(u.Phone) {
	// 	return errors.New("invalid phone number")
	// }

	if utils.CharCount(u.Password) > maxPwdLen {
		return ErrMaxLength{Field: "password", MaxLen: maxPwdLen}
	}

	if utils.CharCount(u.Password) < minPwdLen {
		return ErrMinLength{Field: "password", MinLen: maxPwdLen}
	}

	if utils.CharCount(u.Username) > maxUsrnameLen {
		return ErrMaxLength{Field: "username", MaxLen: maxUsrnameLen}
	}

	if utils.CharCount(u.Username) < minStrLen {
		return ErrMinLength{Field: "username", MinLen: minStrLen}
	}

	return nil
}
