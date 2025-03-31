package property

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"github.com/Z3DRP/lessor-service/pkg/geo"
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
	// mutates the address adding the lat and lng of addrss
	if err := p.GeocodeAddress(pdata); err != nil {
		return nil, fmt.Errorf("could not determine property coordinates %v", err)
	}

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
	isAddrsUpdated, err := p.isAddressDif(ctx, pdto.Pid, pdto.Address)

	if err != nil {
		log.Printf("could not compare previous address with new address %v", err)
		return nil, fmt.Errorf("could not compare previous address with new address %v", err)
	}

	log.Printf("address was updated %v", isAddrsUpdated)
	if isAddrsUpdated {
		// mutates the address adding the lat and lng of addrss
		if err = p.GeocodeAddress(pdto); err != nil {
			return nil, fmt.Errorf("could not determine new property coordinates %v", err)
		}
	}

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
		// do because the compiler is complaining about err below shadowing err above even though i thought that the variable at second pos in := is not initilization but assignment
		var fileName string
		fileName, err = p.s3Actor.Upload(ctx, prpty.LessorId.String(), prpty.Pid.String(), fileData)

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

func (p PropertyService) getExistingAddress(ctx context.Context, id string) (model.Address, error) {
	log.Println("fetching existing address")
	property, err := p.repo.GetExisting(ctx, id)
	if err != nil {
		return model.Address{}, err
	}

	addrs, err := p.decodeAddrs(property.Address)
	if err != nil {
		log.Printf("failed to decode existing address %v", err)
		return model.Address{}, err
	}
	log.Printf("existing address %v", addrs)

	return addrs, nil
}

func (p PropertyService) isAddressDif(ctx context.Context, pid string, addr json.RawMessage) (bool, error) {
	existingAddrs, err := p.getExistingAddress(ctx, pid)
	if err != nil {
		log.Printf("failed to get exisitn address %v", err)
		return false, err
	}

	addrss, err := p.decodeAddrs(addr)
	if err != nil {
		log.Printf("failed to decode address attached to proeprty being updated %v", err)
		return false, err
	}

	log.Printf("address attached to updated property %v", addrss)

	return addrss.Street != existingAddrs.Street ||
		addrss.City != existingAddrs.City ||
		addrss.State != existingAddrs.State ||
		addrss.Country != existingAddrs.Country, nil
}

func (p PropertyService) decodeAddrs(addr json.RawMessage) (model.Address, error) {
	var addrs model.Address
	if err := json.Unmarshal(addr, &addrs); err != nil {
		return model.Address{}, err
	}

	return addrs, nil
}

func (p PropertyService) GeocodeAddress(payload interface{}) error {
	var payloadAddrs *json.RawMessage
	switch v := payload.(type) {
	case *dtos.PropertyRequest:
		payloadAddrs = &v.Address
	case *dtos.PropertyModificationRequest:
		payloadAddrs = &v.Address
	default:
		return errors.New("unsupported dto type")
	}

	gAddrs, err := geo.NewGAddress(*payloadAddrs)

	if err != nil {
		return err
	}

	gActor := geo.NewGeoActor()
	locCoordinates, err := gActor.GeoCode(gAddrs)

	if err != nil {
		// %s will print out the actual json of the bytes
		return fmt.Errorf("failed to get geo coordinates for %s err %v", gAddrs, err)
	}

	var updatedAddrs model.Address
	if err = json.Unmarshal(*payloadAddrs, &updatedAddrs); err != nil {
		return fmt.Errorf("failed to decode address frm payload %v", err)
	}

	updatedAddrs.Lat = locCoordinates.Latitude
	updatedAddrs.Lng = locCoordinates.Longitude
	encodedAddrs, err := json.Marshal(updatedAddrs)
	if err != nil {
		return fmt.Errorf("failed to encode udpated address %v", err)
	}

	*payloadAddrs = encodedAddrs
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
