package dtos

import (
	"errors"
)

type AlessorRequestDto struct {
	Id  int64
	Uid string
	Bid string
}

func (a *AlessorRequestDto) Validate() error {
	if !IsInBufferRange(a.Id) {
		return errors.New("invalid id, out of bounds")
	}
	if !IsInBufferRange(a.Uid) {
		return errors.New("invalid uid, out of bounds")
	}

	if a.Bid != "" && !IsValidUUID(a.Bid) {
		return errors.New("invalid Bid")
	}

	if a.Uid != "" && !IsValidUUID(a.Uid) {
		return errors.New("invalid uid")
	}

	return nil
}

type AlessorRequest struct {
	Id                        int64
	Uid                       string
	Bid                       string
	TotalProperties           int64
	SquareAccount             string
	PaymentIntegrationEnabled bool
	PaymentSchedule           map[string]interface{}
	CommunicationPrefrences   string
}

func (a *AlessorRequest) Validate() error {
	if !IsInBufferRange(a.Id) {
		return errors.New("invalid id, out of bounds")
	}

	if !IsInBufferRange(a.Uid) {
		return errors.New("invalid uid, out of bounds")
	}

	if a.Bid != "" && !IsValidUUID(a.Bid) {
		return errors.New("invalid bid")
	}

	if a.Uid != "" && !IsValidUUID(a.Uid) {
		return errors.New("invalid uid")
	}

	if !IsInBufferRange(a.TotalProperties) {
		return errors.New("invalid total properties")
	}

	if len(a.PaymentSchedule) == 0 {
		return errors.New("payment schedule must be defined")
	}

	return nil
}
