package services

import "fmt"

type Service interface {
	ServiceName() string
}

type ErrInvalidRequest struct {
	ServiceType string
	RequestType string
	Err         error
}

func (e ErrInvalidRequest) Error() string {
	return fmt.Sprintf("invalid %v service %v request", e.ServiceType, e.RequestType)
}

func (e ErrInvalidRequest) Unwrap() error {
	return e.Err
}
