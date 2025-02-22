package dtos

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type PropertyRequest struct {
	AlessorId    string
	Address      map[string]interface{}
	Bedrooms     int
	Baths        int
	SquareFt     float64
	Available    bool
	Status       string
	Notes        string
	TaxRate      float64
	TaxAmountDue float64
	MaxOccupancy int
}

func NewPropertyRequest(aid string, addr map[string]interface{}, bdrm, bth int, sqft float64, avb bool, stat, note string, txRate, txAmnt float64, occp int) PropertyRequest {
	return PropertyRequest{
		AlessorId:    aid,
		Address:      addr,
		Bedrooms:     bdrm,
		Baths:        bth,
		SquareFt:     sqft,
		Available:    avb,
		Status:       stat,
		Notes:        note,
		TaxRate:      txRate,
		TaxAmountDue: txAmnt,
		MaxOccupancy: occp,
	}
}

func (p *PropertyRequest) Validate() error {
	if _, err := uuid.Parse(p.AlessorId); err != nil {
		return fmt.Errorf("invalid uuid %v", err)
	}

	if p.Status != "pending" && p.Status != "in-progress" && p.Status != "completed" && p.Status != "unknown" {
		return fmt.Errorf("invalid property status %v not supported", p.Status)
	}
	return basePropertyValidate(p)
}

type PropertyModificationRequest struct {
	Pid     string
	Request PropertyRequest
}

func (p *PropertyModificationRequest) Validate() error {
	if _, err := uuid.Parse(p.Request.AlessorId); err != nil {
		return errors.New("invalid alessor id")
	}

	if _, err := uuid.Parse(p.Pid); err != nil {
		return errors.New("invalid pid")
	}

	return p.Request.Validate()
}

func NewPropertyModRequest(id string, p PropertyRequest) PropertyModificationRequest {
	return PropertyModificationRequest{
		Pid:     id,
		Request: p,
	}
}

func basePropertyValidate(p *PropertyRequest) error {
	if p.Address == nil {
		return errors.New("address is required")
	}

	if p.MaxOccupancy <= 0 {
		return errors.New("max occupancy is required")
	}

	if p.Baths <= 0 {
		return errors.New("number of baths is requried")
	}

	if p.Bedrooms <= 0 {
		return errors.New("number of bedrooms is required")
	}

	return nil
}
