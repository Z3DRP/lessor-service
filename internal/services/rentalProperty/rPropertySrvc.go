package rentalproperty

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/internal/services"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type RentalPropertyService struct {
	repo   dac.RentalPropertyRepo
	logger *crane.Zlogrus
	//s3Actor api.FilePersister
}

func (p RentalPropertyService) ServiceName() string {
	return "Rental Property"
}

func NewRentalPropertyService(repo dac.RentalPropertyRepo, logr *crane.Zlogrus) RentalPropertyService {
	return RentalPropertyService{
		repo:   repo,
		logger: logr,
	}
}

func (p RentalPropertyService) GetRentalProperty(ctx context.Context, fltr filters.Filterer) (*dtos.RentalPropertyDto, error) {
	uidFilter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	prpty, err := p.repo.Fetch(ctx, uidFilter)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	property, ok := prpty.(model.RentalProperty)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	reqDto := dtos.NewRentalPropertyDto(property)

	return &reqDto, nil
}

func (p RentalPropertyService) GetRentalProperties(ctx context.Context, fltr filters.Filterer) ([]dtos.RentalPropertyDto, error) {
	// need to add a uuid filter for all repos because that way it limits the results in multi tenant db
	var propResponses []dtos.RentalPropertyDto
	filter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	properties, err := p.repo.FetchAll(ctx, filter)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dac.ErrNoResults{Err: err, Shape: propResponses, Identifier: "all"}
		}
		return nil, err
	}

	return dtos.NewRentalDtos(properties), nil
}

func (p RentalPropertyService) CreateRentalProperty(ctx context.Context, pdata *dtos.RentalPropertyDto) (*dtos.RentalPropertyDto, error) {
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

	prpty, ok := newPrpty.(*model.RentalProperty)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: &model.RentalProperty{}, Got: newPrpty}
	}

	response := dtos.NewRentalPropertyDtoFrmPtr(prpty)
	return &response, nil
}

func (p RentalPropertyService) ModifyRentalProperty(ctx context.Context, pdto *dtos.RentalPropertyDto) (*dtos.RentalPropertyDto, error) {
	prpty := newPropertyModRequest(pdto)

	if prpty.Pid == uuid.Nil {
		log.Printf("could not parse property pid as uuid")
		return nil, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Upate", Err: nil}
	}

	updatePrpty, err := p.repo.Update(ctx, prpty)

	if err != nil {
		log.Printf("err updating property %v", err)
		return nil, err
	}

	property, ok := updatePrpty.(model.RentalProperty)

	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.RentalProperty{}, Got: updatePrpty}
		log.Printf("type assertion failed %v", err)
		return nil, cmerr.ErrUnexpectedData{Wanted: model.RentalProperty{}, Got: updatePrpty}
	}

	response := dtos.NewRentalPropertyDto(property)
	return &response, nil
}

func (p RentalPropertyService) DeleteRentalProperty(ctx context.Context, f filters.Filterer) error {
	fltr, ok := f.(filters.IdFilter)
	if !ok {
		return errors.New("failed to create id filter")
	}

	if err := fltr.Validate(); err != nil {
		return fmt.Errorf("invalid request, %v", err)
	}

	pid, _ := uuid.Parse(fltr.Identifier)
	err := p.repo.Delete(ctx, model.RentalProperty{Pid: pid})

	if err != nil {
		return err
	}

	return nil
}

func newPropertyRequest(data *dtos.RentalPropertyDto) *model.RentalProperty {
	return &model.RentalProperty{
		RentalPrice:       data.RentalPrice,
		RentDueDate:       data.RentDueDate,
		LeaseSigned:       data.LeaseSigned,
		LeaseDuration:     data.LeaseDuration,
		LeaseRenewDate:    data.LeaseRenewDate,
		IsVacant:          data.IsVacant,
		PetFriendly:       data.PetFriendly,
		NeedsEviction:     data.NeedsEviction,
		EvictionStartDate: data.EvictionStartDate,
	}
}

func newPropertyModRequest(data *dtos.RentalPropertyDto) model.RentalProperty {
	return model.RentalProperty{
		Pid:               utils.ParseUuid(data.Pid),
		RentalPrice:       data.RentalPrice,
		RentDueDate:       data.RentDueDate,
		LeaseSigned:       data.LeaseSigned,
		LeaseDuration:     data.LeaseDuration,
		LeaseRenewDate:    data.LeaseRenewDate,
		IsVacant:          data.IsVacant,
		PetFriendly:       data.PetFriendly,
		NeedsEviction:     data.NeedsEviction,
		EvictionStartDate: data.EvictionStartDate,
	}
}
