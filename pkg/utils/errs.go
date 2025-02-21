package utils

import (
	"fmt"
	"net/http"
)

type ErrRequestTimeout struct {
	Request *http.Request
}

func (e *ErrRequestTimeout) Error() string {
	return fmt.Sprintf("request: %s method: %s timed out", e.Request.URL, e.Request.Method)
}

func NewRequestTimeoutErr(r *http.Request) *ErrRequestTimeout {
	return &ErrRequestTimeout{
		Request: r,
	}
}

type ErrMissingId struct {
	Obj       string
	FieldName string
}

func (e ErrMissingId) Error() string {
	return fmt.Sprintf("%v is missing the id field %v", e.Obj, e.FieldName)
}
