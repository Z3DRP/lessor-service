package dtos

import (
	"errors"
	"time"

	"github.com/Z3DRP/lessor-service/pkg/utils"
)

type ProfileSearchDto struct {
	Id         int64
	Uid        string
	Username   string
	IsActive   bool
	SearchType map[string]bool
}

func (p *ProfileSearchDto) Validate() error {
	if _, ok := p.SearchType["uid"]; ok {
		if p.Uid != "" && !IsValidUUID(p.Uid) {
			return errors.New("invalid Uid for profile search")
		}
	}

	if _, ok := p.SearchType["id"]; ok {
		if p.Id == 0 {
			return errors.New("invalid id for profile search")
		}
	}
	return nil
}

type ProfileRequest struct {
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

// need to get the limits for the varchars

func (p *ProfileRequest) Validate() error {
	if p.Id == 0 {
		return errors.New("invalid profile id")
	}

	if p.Uid != "" && !IsValidUUID(p.Uid) {
		return errors.New("invalid profile uid")
	}

	if p.FirstName == "" || p.LastName == "" {
		return errors.New("first name and last name are required")
	}

	if utils.CharCount(p.FirstName) > maxFnameLen {
		return ErrMaxLength{Field: "first name", MaxLen: maxFnameLen}
	}

	if utils.CharCount(p.FirstName) < minStrLen {
		return ErrMinLength{Field: "first name", MinLen: maxFnameLen}
	}

	if utils.CharCount(p.LastName) > maxFnameLen {
		return ErrMaxLength{Field: "last name", MaxLen: maxFnameLen}
	}

	if utils.CharCount(p.LastName) < minStrLen {
		return ErrMinLength{Field: "last name", MinLen: maxFnameLen}
	}

	if utils.CharCount(p.Email) > maxEmalLen {
		return ErrMaxLength{Field: "email", MaxLen: maxFnameLen}
	}

	if !utils.IsValidEmail(p.Email) {
		return errors.New("invalid email address")
	}

	if utils.CharCount(p.Phone) > maxPhneLen {
		return ErrMaxLength{Field: "email", MaxLen: maxFnameLen}
	}

	if utils.CharCount(p.Phone) < maxPhneLen {
		return ErrMinLength{Field: "phone", MinLen: maxFnameLen}
	}

	if !utils.IsValidPhone(p.Phone) {
		return errors.New("invalid phone number")
	}

	if utils.CharCount(p.Password) > maxPwdLen {
		return ErrMaxLength{Field: "password", MaxLen: maxPwdLen}
	}

	if utils.CharCount(p.Password) < minPwdLen {
		return ErrMinLength{Field: "password", MinLen: maxPwdLen}
	}

	return nil
}

type ProfileSignUpRequest struct {
	FirstName   string
	LastName    string
	ProfileType string
	Username    string
	Phone       string
	Email       string
	Password    string
}

func (p *ProfileSignUpRequest) Validate() error {
	if utils.CharCount(p.FirstName) > maxFnameLen {
		return ErrMaxLength{Field: "first name", MaxLen: maxFnameLen}
	}
	if utils.CharCount(p.FirstName) < minStrLen {
		return ErrMinLength{Field: "first name", MinLen: maxFnameLen}
	}

	if utils.CharCount(p.LastName) > maxFnameLen {
		return ErrMaxLength{Field: "last name", MaxLen: maxFnameLen}
	}

	if utils.CharCount(p.LastName) < minStrLen {
		return ErrMinLength{Field: "last name", MinLen: maxLnameLen}
	}

	if utils.CharCount(p.Email) > maxEmalLen {
		return ErrMaxLength{Field: "email", MaxLen: maxFnameLen}
	}

	if !utils.IsValidEmail(p.Email) {
		return errors.New("invalid email address")
	}

	if utils.CharCount(p.Phone) > maxPhneLen {
		return ErrMaxLength{Field: "email", MaxLen: maxFnameLen}
	}

	if utils.CharCount(p.Phone) < maxPhneLen {
		return ErrMinLength{Field: "phone", MinLen: maxFnameLen}
	}

	if !utils.IsValidPhone(p.Phone) {
		return errors.New("invalid phone number")
	}

	if utils.CharCount(p.Password) > maxPwdLen {
		return ErrMaxLength{Field: "password", MaxLen: maxPwdLen}
	}

	if utils.CharCount(p.Password) < minPwdLen {
		return ErrMinLength{Field: "password", MinLen: maxPwdLen}
	}

	if utils.CharCount(p.Username) > maxUsrnameLen {
		return ErrMaxLength{Field: "username", MaxLen: maxUsrnameLen}
	}

	return nil
}
