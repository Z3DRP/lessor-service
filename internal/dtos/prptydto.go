package dtos

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/google/uuid"
)

type PropertyDto struct {
	LessorId      string          `json:"alessorId"`
	Status        string          `json:"status"`
	Notes         string          `json:"notes"`
	Image         string          `json:"image"`
	Address       json.RawMessage `json:"address"`
	Bedrooms      float64         `json:"bedrooms"`
	Baths         float64         `json:"baths"`
	SquareFootage float64         `json:"squareFootage"`
	TaxAmountDue  float64         `json:"taxAmountDue"`
	TaxRate       float64         `json:"taxRate"`
	MaxOccupancy  int             `json:"maxOccupancy"`
	IsAvailable   bool            `json:"isAvailable"`
}

func NewPropertyDto(p model.Property) PropertyDto {
	return PropertyDto{
		LessorId:      p.LessorId.String(),
		Status:        string(p.Status),
		Notes:         p.Notes,
		Image:         p.Image,
		Address:       p.Address,
		Bedrooms:      p.Bedrooms,
		Baths:         p.Baths,
		SquareFootage: p.SquareFootage,
		TaxAmountDue:  p.TaxAmountDue,
		TaxRate:       p.TaxRate,
		MaxOccupancy:  p.MaxOccupancy,
		IsAvailable:   p.IsAvailable,
	}
}

type PropertyRequest struct {
	AlessorId    string          `json:"alessorId"`
	Status       string          `json:"status"`
	Notes        string          `json:"notes"`
	Image        string          `json:"image"`
	Address      json.RawMessage `json:"address"`
	Bedrooms     float64         `json:"bedrooms"`
	Baths        float64         `json:"baths"`
	SquareFt     float64         `json:"squareFootage"`
	TaxAmountDue float64         `json:"taxAmountDue"`
	TaxRate      float64         `json:"taxRate"`
	MaxOccupancy int             `json:"maxOccupancy"`
	IsAvailable  bool            `json:"isAvailable"`
}

func NewPropertyRequest(aid string, addr json.RawMessage, bdrm, bth, sqft float64, avb bool, stat, note, fileName string, txRate, txAmnt float64, occp int) PropertyRequest {
	return PropertyRequest{
		AlessorId:    aid,
		Address:      addr,
		Bedrooms:     bdrm,
		Baths:        bth,
		Image:        fileName,
		SquareFt:     sqft,
		IsAvailable:  avb,
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

	return basePropertyValidate(p)
}

type PropertyModificationRequest struct {
	Pid          string          `json:"pid"`
	AlessorId    string          `json:"alessorId"`
	Status       string          `json:"status"`
	Notes        string          `json:"notes"`
	Image        string          `json:"image"`
	Address      json.RawMessage `json:"address"`
	Bedrooms     float64         `json:"bedrooms"`
	Baths        float64         `json:"baths"`
	SquareFt     float64         `json:"squareFootage"`
	TaxAmountDue float64         `json:"taxAmountDue"`
	TaxRate      float64         `json:"taxRate"`
	MaxOccupancy int             `json:"maxOccupancy"`
	IsAvailable  bool            `json:"isAvailable"`
}

func (p *PropertyModificationRequest) Validate() error {
	if _, err := uuid.Parse(p.AlessorId); err != nil {
		return errors.New("invalid alessor id")
	}

	if _, err := uuid.Parse(p.Pid); err != nil {
		return errors.New("invalid pid")
	}

	return nil
}

type PropertyResponse struct {
	Pid          string          `json:"pid"`
	LessorId     string          `json:"alessorId"`
	Status       string          `json:"status"`
	Notes        string          `json:"notes"`
	Image        string          `json:"image"`
	Address      json.RawMessage `json:"address"`
	Bedrooms     float64         `json:"bedrooms"`
	Baths        float64         `json:"baths"`
	SquareFt     float64         `json:"squareFootage"`
	TaxAmountDue float64         `json:"taxAmountDue"`
	TaxRate      float64         `json:"taxRate"`
	MaxOccupancy int             `json:"maxOccupancy"`
	IsAvailable  bool            `json:"isAvailable"`
	ImageUrl     *string         `json:"imageUrl"`
}

func (p *PropertyResponse) Valiate() error {
	return nil
}

func NewPropertyResponse(p model.Property, url *string) PropertyResponse {
	return PropertyResponse{
		Pid:          p.Pid.String(),
		LessorId:     p.LessorId.String(),
		Status:       string(p.Status),
		Notes:        p.Notes,
		Image:        p.Image,
		Address:      p.Address,
		Bedrooms:     p.Bedrooms,
		Baths:        p.Baths,
		SquareFt:     p.SquareFootage,
		TaxAmountDue: p.TaxAmountDue,
		TaxRate:      p.TaxRate,
		MaxOccupancy: p.MaxOccupancy,
		IsAvailable:  p.IsAvailable,
		ImageUrl:     url,
	}
}

func NewPropertyResponseFrmPointer(p *model.Property, url *string) PropertyResponse {
	return PropertyResponse{
		Pid:          p.Pid.String(),
		LessorId:     p.LessorId.String(),
		Status:       string(p.Status),
		Notes:        p.Notes,
		Image:        p.Image,
		Address:      p.Address,
		Bedrooms:     p.Bedrooms,
		Baths:        p.Baths,
		SquareFt:     p.SquareFootage,
		TaxAmountDue: p.TaxAmountDue,
		TaxRate:      p.TaxRate,
		MaxOccupancy: p.MaxOccupancy,
		IsAvailable:  p.IsAvailable,
		ImageUrl:     url,
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
