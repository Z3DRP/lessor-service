package property

import (
	"context"
	"database/sql"
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

	//propertyImgs := make(map[string]string)
	log.Printf("finding images for owner : %v", filter.Identifier)

	imageUrls, err := p.s3Actor.List(ctx, filter.Identifier)
	log.Printf("image urls check %+v", imageUrls)

	for k, url := range imageUrls {
		log.Printf("image url k: %v, url: %v", k, url)
	}

	if err != nil {
		if err == api.ErrrNoImagesFound {
			// no images so return properties found
			log.Print("properties found but no iamges")
			for _, prop := range properties {
				propResponses = append(propResponses, dtos.NewPropertyResponse(prop, nil))
			}
			return propResponses, nil
		}
		log.Printf("list images response err props found but err occurred %v", err)
		return nil, err
	}

	log.Println("all good here about to loop presigned image urls")

	for _, prop := range properties {
		// prop.Image has the entire s3 path and file key i.e. property/{ownerId}/{objId}/filename
		log.Printf("looking for property image %v", prop.Image)
		if url, found := imageUrls[prop.Image]; found {
			log.Println("found image for property")
			propResponses = append(propResponses, dtos.NewPropertyResponse(prop, &url))
		}
	}

	log.Printf("returning responses %+v", propResponses)

	return propResponses, nil
}

func (p PropertyService) CreateProperty(ctx context.Context, pdata *dtos.PropertyRequest, fileData *ztype.FileUploadDto) (*model.Property, error) {
	property := newPropertyRequest(pdata)
	var err error

	property.Pid, err = uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	if fileData != nil {
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

	return prpty, nil
}

func (p PropertyService) ModifyProperty(ctx context.Context, pdto dtos.PropertyModificationRequest, fileData *ztype.FileUploadDto) (model.Property, error) {
	prpty := newPropertyModRequest(pdto)

	if prpty.Pid == uuid.Nil {
		return model.Property{}, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Upate", Err: nil}
	}

	if fileData != nil {
		fileName, err := p.s3Actor.Upload(ctx, prpty.LessorId.String(), prpty.Pid.String(), fileData)

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

func newPropertyModRequest(data dtos.PropertyModificationRequest) *model.Property {
	return &model.Property{
		Pid:           utils.ParseUuid(data.Pid),
		LessorId:      utils.ParseUuid(data.Request.AlessorId),
		Address:       data.Request.Address,
		Bedrooms:      data.Request.Bedrooms,
		Baths:         data.Request.Baths,
		SquareFootage: data.Request.SquareFt,
		IsAvailable:   data.Request.IsAvailable,
		Status:        model.PropertyStatus(data.Request.Status),
		Notes:         data.Request.Notes,
		TaxRate:       data.Request.TaxRate,
		TaxAmountDue:  data.Request.TaxAmountDue,
		MaxOccupancy:  data.Request.MaxOccupancy,
	}
}
