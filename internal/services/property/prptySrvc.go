package property

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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
	repo   dac.PropertyRepo
	logger *crane.Zlogrus
	//s3Actor api.S3Actor
	s3Actor api.FilePersister
}

func (p PropertyService) ServiceName() string {
	return "Property"
}

func NewPropertyService(repo dac.PropertyRepo, actr api.S3Actor, logr *crane.Zlogrus) PropertyService {
	return PropertyService{
		repo:    repo,
		s3Actor: actr,
		logger:  logr,
	}
}

func (p PropertyService) GetProperty(ctx context.Context, fltr filters.Filterer) (*dtos.PropertyResponse, error) {
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

	property, ok := prpty.(model.Property)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: prpty}
	}

	var reqDto dtos.PropertyResponse

	if property.Image != "" {
		fileUrl, err := p.s3Actor.Get(ctx, property.LessorId.String(), property.Pid.String(), property.Image)

		if err != nil {
			return nil, err
		}

		reqDto = dtos.NewPropertyResponse(property, &fileUrl)
	}

	return &reqDto, nil
}

func (p PropertyService) GetProperties(ctx context.Context, fltr filters.Filterer) ([]dtos.PropertyResponse, error) {
	// need to add a uuid filter for all repos because that way it limits the results in multi tenant db
	var propResponses []dtos.PropertyResponse
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

	imageUrls, err := p.s3Actor.List(ctx, filter.Identifier)

	if err != nil {
		if err == api.ErrrNoImagesFound {
			// no images so return properties found
			for _, prop := range properties {
				propResponses = append(propResponses, dtos.NewPropertyResponse(prop, nil))
			}
			return propResponses, nil
		}
		return nil, err
	}

	for _, prop := range properties {
		// prop.Image has the entire s3 path and file key i.e. property/{ownerId}/{objId}/filename
		if url, found := imageUrls[prop.Image]; found {
			propResponses = append(propResponses, dtos.NewPropertyResponse(prop, &url))
		} else {
			propResponses = append(propResponses, dtos.NewPropertyResponse(prop, nil))
		}
	}

	return propResponses, nil
}

func (p PropertyService) CreateProperty(ctx context.Context, pdata *dtos.PropertyRequest, fileData *ztype.FileUploadDto) (*dtos.PropertyResponse, error) {
	property := newPropertyRequest(pdata)
	var err error

	property.Pid, err = uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	if fileData != nil && fileData.File != nil && fileData.Header != nil {
		var fileName string
		fileName, err = p.s3Actor.Upload(ctx, property.LessorId.String(), property.Pid.String(), fileData)

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

	var psUrl *string
	if property.Image != "" {
		url, err := p.s3Actor.GetFile(ctx, property.Image)

		if err != nil {
			return nil, err
		}
		psUrl = &url
	}

	response := dtos.NewPropertyResponseFrmPointer(prpty, psUrl)
	return &response, nil
}

func (p PropertyService) ModifyProperty(ctx context.Context, pdto *dtos.PropertyModificationRequest, fileData *ztype.FileUploadDto) (*dtos.PropertyResponse, error) {
	prpty := newPropertyModRequest(pdto)

	if prpty.Pid == uuid.Nil {
		log.Printf("could not parse property pid as uuid")
		return nil, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Upate", Err: nil}
	}

	existingPropertyImg, err := p.getExistingPropertyImage(ctx, prpty.Pid.String())

	if err != nil {
		log.Printf("failed to fetch existing property to check for image")
		return nil, err
	}

	if fileData != nil && fileData.File != nil && fileData.Header != nil {
		fileName, err := p.s3Actor.Upload(ctx, prpty.LessorId.String(), prpty.Pid.String(), fileData)

		if err != nil {
			log.Printf("error uploading file %v", err)
			return nil, err
		}

		prpty.Image = fileName
	} else {
		prpty.Image = existingPropertyImg
	}

	updatePrpty, err := p.repo.Update(ctx, prpty)

	if err != nil {
		log.Printf("err updating property %v", err)
		return nil, err
	}

	property, ok := updatePrpty.(model.Property)

	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: updatePrpty}
		log.Printf("type assertion failed %v", err)
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Property{}, Got: updatePrpty}
	}

	var psUrl *string
	if property.Image != "" {
		url, err := p.s3Actor.GetFile(ctx, property.Image)

		if err != nil {
			log.Printf("failed to get file %v", err)
			return nil, err
		}

		psUrl = &url
	}

	response := dtos.NewPropertyResponse(property, psUrl)
	return &response, nil
}

func (p PropertyService) DeleteProperty(ctx context.Context, f filters.Filterer) error {
	fltr, ok := f.(filters.IdFilter)
	if !ok {
		return errors.New("failed to create id filter")
	}

	if err := fltr.Validate(); err != nil {
		return fmt.Errorf("invalid request, %v", err)
	}

	pid, _ := uuid.Parse(fltr.Identifier)
	err := p.repo.Delete(ctx, model.Property{Pid: pid})

	if err != nil {
		return err
	}

	return nil
}

func (p PropertyService) getExistingPropertyImage(ctx context.Context, id string) (string, error) {
	property, err := p.repo.GetExisting(ctx, id)
	if err != nil {
		return "", nil
	}

	return property.Image, nil
}

func newPropertyRequest(data *dtos.PropertyRequest) *model.Property {
	return &model.Property{
		LessorId:      utils.ParseUuid(data.AlessorId),
		Address:       data.Address,
		Bedrooms:      data.Bedrooms,
		Baths:         data.Baths,
		SquareFootage: data.SquareFt,
		IsAvailable:   data.IsAvailable,
		Status:        model.PropertyStatus(data.Status),
		Notes:         data.Notes,
		TaxRate:       data.TaxRate,
		TaxAmountDue:  data.TaxAmountDue,
		MaxOccupancy:  data.MaxOccupancy,
	}
}

func newPropertyModRequest(data *dtos.PropertyModificationRequest) model.Property {
	return model.Property{
		Pid:           utils.ParseUuid(data.Pid),
		LessorId:      utils.ParseUuid(data.AlessorId),
		Address:       data.Address,
		Bedrooms:      data.Bedrooms,
		Baths:         data.Baths,
		SquareFootage: data.SquareFt,
		IsAvailable:   data.IsAvailable,
		Status:        model.PropertyStatus(data.Status),
		Notes:         data.Notes,
		TaxRate:       data.TaxRate,
		TaxAmountDue:  data.TaxAmountDue,
		MaxOccupancy:  data.MaxOccupancy,
	}
}
