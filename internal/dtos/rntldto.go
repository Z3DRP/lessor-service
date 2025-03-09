package dtos

import (
	"time"

	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/shopspring/decimal"
)

type RentalPropertyDto struct {
	Pid string `json:"pid"`
	//Property *PropertyResponse
	RentalPrice       decimal.Decimal `json:"rentalPrice"`
	RentDueDate       time.Time       `json:"rentDueDate"`
	LeaseSigned       bool            `json:"leaseSigned"`
	LeaseDuration     int             `json:"leaseDuration"`
	LeaseRenewDate    time.Time       `json:"leaseRenewDate"`
	IsVacant          bool            `json:"isVacant"`
	PetFriendly       bool            `json:"petFriendly"`
	NeedsEviction     bool            `json:"needsEviction"`
	EvictionStartDate time.Time       `json:"evictionStartDate"`
}

func (r *RentalPropertyDto) Validte() error {
	return nil
}

func NewRentalPropertyDto(r model.RentalProperty) RentalPropertyDto {
	return RentalPropertyDto{
		Pid:               r.Pid.String(),
		RentalPrice:       r.RentalPrice,
		RentDueDate:       r.RentDueDate,
		LeaseSigned:       r.LeaseSigned,
		LeaseDuration:     r.LeaseDuration,
		LeaseRenewDate:    r.LeaseRenewDate,
		IsVacant:          r.IsVacant,
		PetFriendly:       r.PetFriendly,
		NeedsEviction:     r.NeedsEviction,
		EvictionStartDate: r.EvictionStartDate,
	}
}

func NewRentalPropertyDtoFrmPtr(r *model.RentalProperty) RentalPropertyDto {
	return RentalPropertyDto{
		Pid:               r.Pid.String(),
		RentalPrice:       r.RentalPrice,
		RentDueDate:       r.RentDueDate,
		LeaseSigned:       r.LeaseSigned,
		LeaseDuration:     r.LeaseDuration,
		LeaseRenewDate:    r.LeaseRenewDate,
		IsVacant:          r.IsVacant,
		PetFriendly:       r.PetFriendly,
		NeedsEviction:     r.NeedsEviction,
		EvictionStartDate: r.EvictionStartDate,
	}
}

func NewRentalDtos(rs []model.RentalProperty) []RentalPropertyDto {
	rentals := make([]RentalPropertyDto, 0)
	for _, r := range rs {
		rentals = append(rentals, NewRentalPropertyDto(r))
	}

	return rentals
}
