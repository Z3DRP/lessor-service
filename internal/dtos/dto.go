package dtos

import (
	"fmt"
	"math"
	"net/http"

	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type Dto interface {
	Validate() error
}

type DeleteRequest struct {
	Identifer string
}

func (d DeleteRequest) Validate() error {
	if d.Identifer == "" {
		return utils.ErrMissingId{Obj: "profile request dto", FieldName: "uid"}
	}

	if !IsValidUUID(d.Identifer) {
		return fmt.Errorf("invalid identifier")
	}

	return nil
}

func BuildDeleteRequest(r *http.Request) (Dto, error) {
	id := r.PathValue("identifier")

	return &DeleteRequest{Identifer: id}, nil
}

type ErrInvalidDto struct {
	DtoType string
	Field   string
	Err     error
}

func (e ErrInvalidDto) Error() string {
	return fmt.Sprintf("invalid %v DTO the following field was incorrect: %v", e.DtoType, e.Field)
}

func (e ErrInvalidDto) Unwrap() error {
	return e.Err
}

func IsValidUUID(uid string) bool {
	_, err := uuid.Parse(uid)
	return err == nil
}

func IsInBufferRange(val interface{}) bool {
	switch v := val.(type) {
	case int64:
		return v <= math.MaxInt64 && v >= math.MinInt64
	case int32:
		return v <= math.MaxInt32 && v >= math.MinInt32
	case int:
		return v <= math.MaxInt && v >= math.MinInt
	}
	return false
}

type ErrMaxLength struct {
	Field  string
	MaxLen int
}

func (e ErrMaxLength) Error() string {
	return fmt.Sprintf("%v must be less than %v characters", e.Field, e.MaxLen)
}

type ErrMinLength struct {
	Field  string
	MinLen int
}

func (e ErrMinLength) Error() string {
	return fmt.Sprintf("%v must be greater than %v characters", e.Field, e.MinLen)
}
