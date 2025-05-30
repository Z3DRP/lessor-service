package adapters

import (
	"encoding/json"
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/geo"
	"github.com/Z3DRP/lessor-service/pkg/utils"
)

func ParsePropertyUpdateForm(r *http.Request) (*dtos.PropertyModificationRequest, error) {
	beds, err := utils.ParseFloatOrZero(r.FormValue("bedrooms"))

	if err != nil {
		return nil, err
	}

	baths, err := utils.ParseFloatOrZero(r.FormValue("baths"))

	if err != nil {
		return nil, err
	}

	squareFt, err := utils.ParseFloatOrZero(r.FormValue("squareFootage"))

	if err != nil {
		return nil, err
	}

	taxRate, err := utils.ParseFloatOrZero(r.FormValue("taxRate"))

	if err != nil {
		return nil, err
	}

	taxDue, err := utils.ParseFloatOrZero(r.FormValue("taxAmountDue"))

	if err != nil {
		return nil, err
	}

	maxOpp, err := utils.ParseIntOrZero(r.FormValue("maxOccupancy"))

	if err != nil {
		return nil, err
	}

	isAvailable := utils.ParseBool(r.FormValue("isAvailable"))

	return &dtos.PropertyModificationRequest{
		Pid:          r.FormValue("pid"),
		Address:      json.RawMessage(r.FormValue("address")),
		AlessorId:    r.FormValue("alessorId"),
		IsAvailable:  isAvailable,
		Bedrooms:     beds,
		Baths:        baths,
		SquareFt:     squareFt,
		Status:       r.FormValue("status"),
		Notes:        r.FormValue("notes"),
		TaxRate:      taxRate,
		TaxAmountDue: taxDue,
		MaxOccupancy: maxOpp,
		Image:        r.FormValue("image"),
	}, nil

}

func ParsePropertyForm(r *http.Request) (*dtos.PropertyRequest, error) {
	beds, err := utils.ParseFloatOrZero(r.FormValue("bedrooms"))

	if err != nil {
		return nil, err
	}

	baths, err := utils.ParseFloatOrZero(r.FormValue("baths"))

	if err != nil {
		return nil, err
	}

	squareFt, err := utils.ParseFloatOrZero(r.FormValue("squareFootage"))

	if err != nil {
		return nil, err
	}

	taxRate, err := utils.ParseFloatOrZero(r.FormValue("taxRate"))

	if err != nil {
		return nil, err
	}

	taxDue, err := utils.ParseFloatOrZero(r.FormValue("taxAmountDue"))

	if err != nil {
		return nil, err
	}

	maxOpp, err := utils.ParseIntOrZero(r.FormValue("maxOccupancy"))

	if err != nil {
		return nil, err
	}

	isAvailable := utils.ParseBool(r.FormValue("isAvailable"))

	return &dtos.PropertyRequest{
		Address:      json.RawMessage(r.FormValue("address")),
		AlessorId:    r.FormValue("alessorId"),
		IsAvailable:  isAvailable,
		Bedrooms:     beds,
		Baths:        baths,
		SquareFt:     squareFt,
		Status:       r.FormValue("status"),
		Notes:        r.FormValue("notes"),
		TaxRate:      taxRate,
		TaxAmountDue: taxDue,
		MaxOccupancy: maxOpp,
		Image:        r.FormValue("image"),
	}, nil
}

func AddressAdapter(adr model.Address) *geo.GAddress {
	return &geo.GAddress{
		Street:  adr.Street,
		City:    adr.City,
		State:   adr.State,
		Country: adr.Country,
		Zipcode: adr.Zipcode,
	}
}
