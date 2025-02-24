package property

import (
	"context"
	"strings"

	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/internal/services"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type PropertyService struct {
	repo    dac.PropertyRepo
	logger  *crane.Zlogrus
	s3Actor api.S3Actor
}

func (p PropertyService) ServiceName() string {
	return "Property"
}

func NewPropertyService(repo dac.PropertyRepo, actr api.S3Actor, logr *crane.Zlogrus) *PropertyService {
	return &PropertyService{
		repo:    repo,
		s3Actor: actr,
		logger:  logr,
	}
}

func (p PropertyService) GetProperty(ctx context.Context, fltr filters.Filterer) (dtos.PropertyResponse, error) {
	uidFilter, ok := fltr.(filters.Filter)

	if !ok {
		return dtos.PropertyResponse{}, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	prpty, err := p.repo.Fetch(ctx, uidFilter)

	if err != nil {
		return dtos.PropertyResponse{}, err
	}

	property, ok := prpty.(model.Property)

	if !ok {
		return dtos.PropertyResponse{}, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	reqDto := dtos.PropertyResponse{Property: property}
	if property.Image != "" {
		fileUrl, err := p.s3Actor.Get(ctx, property.AlessorId.String(), property.Pid.String(), property.Image)

		if err != nil {
			return dtos.PropertyResponse{}, err
		}

		reqDto = dtos.NewPropertyResposne(property, &fileUrl)
	}

	return reqDto, nil
}

func (p PropertyService) GetProperties(ctx context.Context, fltr filters.Filterer) ([]dtos.PropertyResponse, error) {
	// need to add a uuid filter for all repos because that way it limits the results in multi tenant db
	filter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	properties, err := p.repo.FetchAll(ctx, filter)

	if err != nil {
		return nil, err
	}

	propertyImgs := make(map[string]string)
	imageUrls, err := p.s3Actor.GetAll(ctx, filter.Identifier)

	if err != nil {
		return nil, err
	}

	for key, url := range imageUrls {
		parts := strings.Split(key, "/")
		if len(parts) < 3 {
			continue
		}

		propertyImgs[key] = url
	}

	var propResponses []dtos.PropertyResponse
	for _, prop := range properties {
		if url, found := propertyImgs[prop.Pid.String()]; found {
			propResponses = append(propResponses, dtos.NewPropertyResposne(prop, &url))
		}
	}

	return propResponses, nil
}

func (p PropertyService) CreateProperty(ctx context.Context, pdata dtos.PropertyRequest, fileData *ztype.FileUploadDto) (*model.Property, error) {
	property := newPropertyRequest(pdata)
	var err error

	property.Pid, err = uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	if fileData != nil {
		fileName, err := p.s3Actor.Upload(ctx, property.AlessorId.String(), property.Pid.String(), fileData)

		if err != nil {
			return nil, err
		}

		property.Image = fileName
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

func (p PropertyService) ModifyProperty(ctx context.Context, pdto dtos.PropertyModificationRequest, fileData *ztype.FileUploadDto) (model.Property, error) {
	prpty := newPropertyModRequest(pdto)

	if prpty.Pid == uuid.Nil {
		return model.Property{}, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Upate", Err: nil}
	}

	if fileData != nil {
		fileName, err := p.s3Actor.Upload(ctx, prpty.AlessorId.String(), prpty.Pid.String(), fileData)

		if err != nil {
			return model.Property{}, err
		}

		prpty.Image = fileName
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
		AlessorId:     utils.ParseUuid(data.Request.AlessorId),
		Address:       data.Request.Address,
		Bedrooms:      data.Request.Bedrooms,
		Baths:         data.Request.Baths,
		SquareFootage: data.Request.SquareFt,
		IsAvailable:   data.Request.Available,
		Status:        model.PropertyStatus(data.Request.Status),
		Notes:         data.Request.Notes,
		TaxRate:       data.Request.TaxRate,
		TaxAmountDue:  data.Request.TaxAmountDue,
		MaxOccupancy:  data.Request.MaxOccupancy,
	}
}
