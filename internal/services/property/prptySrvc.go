package property

import (
	"context"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/internal/services"
	"github.com/Z3DRP/lessor-service/internal/services/property"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type PropertyService struct {
	repo   dac.PropertyRepo
	logger *crane.Zlogrus
}

func (p PropertyService) ServiceName() string {
	return "Property"
}

func NewPropertyService(repo dac.PropertyRepo, logr *crane.Zlogrus) *PropertyService {
	return &PropertyService{
		repo:   repo,
		logger: logr,
	}
}

func (p PropertyService) GetProperty(ctx context.Context, fltr filters.Filterer) (model.Property, error) {
	uidFilter, ok := fltr.(filters.UuidFilter)

	if !ok {
		return model.Property{}, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	prpty, err := p.repo.Fetch(ctx, uidFilter)

	if err != nil {
		return model.Property{}, err
	}

	property, ok := prpty.(model.Property)

	if !ok {
		return model.Property{}, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	return property, nil
}

func (p PropertyService) GetProperties(ctx context.Context, fltr filters.Filterer) ([]model.Property, error) {
	// need to add a uuid filter for all repos because that way it limits the results in multi tenant db
	filter, ok := fltr.(filters.UuidFilter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	properties, err := p.repo.FetchAll(ctx, filter)

	if err != nil {
		return nil, err
	}

	return properties, nil
}

func (p PropertyService) CreateProperty(ctx context.Context, pdata dtos.PropertyRequest) (*model.Property, error) {
	property := newPropertyRequest(pdata)
	var err error

	property.Pid, err = uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	newPrpty, err := p.repo.Insert(ctx, property)

	if err != nil {
		return nil, err
	}

	prpty, ok := newPrpty.(*model.Property)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: &model.Property{}, Got: newPrpty}
	}

	return prpty, nil
}

func (p PropertyService) ModifyProperty(ctx context.Context, pdto dtos.PropertyModificationRequest) (model.Property, error) {
	prpty := newPropertyModRequest(pdto)

	if prpty.Pid == uuid.Nil {
		return model.Property{}, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Upate", Err: nil}
	}

	updatePrpty, err := p.repo.Update(ctx, prpty)

	if err != nil {
		return model.Property{}, err
	}

	property, ok := updatePrpty.(model.Property)

	if !ok {
		return model.Property{}, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: updatePrpty}
	}

	return property, nil
}

func (p PropertyService) DeleteProperty(ctx context.Context, delReq dtos.DeleteRequest) error {
	pid, _ := uuid.Parse(delReq.Identifer)
	err := p.repo.Delete(ctx, model.Property{Pid: pid})

	if err != nil {
		return err
	}

	return nil
}

func newPropertyRequest(data dtos.PropertyRequest) *model.Property {
	return &model.Property{
		AlessorId:     utils.ParseUuid(data.AlessorId),
		Address:       data.Address,
		Bedrooms:      data.Bedrooms,
		Baths:         data.Baths,
		SquareFootage: data.SquareFt,
		IsAvailable:   data.Available,
		Status:        model.PropertyStatus(data.Status),
		Notes:         data.Notes,
		TaxRate:       data.TaxRate,
		TaxAmountDue:  data.TaxAmountDue,
		MaxOccupancy:  data.MaxOccupancy,
	}
}

func newPropertyModRequest(data dtos.PropertyModificationRequest) *model.Property {
	return &model.Property{
		Pid:           utils.ParseUuid(data.Pid),
		AlessorId:     utils.ParseUuid(data.AlessorId),
		Address:       data.Address,
		Bedrooms:      data.Bedrooms,
		Baths:         data.Baths,
		SquareFootage: data.SquareFt,
		IsAvailable:   data.Available,
		Status:        model.PropertyStatus(data.Status),
		Notes:         data.Notes,
		TaxRate:       data.TaxRate,
		TaxAmountDue:  data.TaxAmountDue,
		MaxOccupancy:  data.MaxOccupancy,
	}
}
