package prfl

import (
	"context"

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

type ProfileService struct {
	repo   dac.ProfileRepo
	logger *crane.Zlogrus
}

func (p ProfileService) ServiceName() string {
	return "Profile"
}

func NewProfileService(repo dac.ProfileRepo, logr *crane.Zlogrus) ProfileService {
	return ProfileService{
		repo:   repo,
		logger: logr,
	}
}

func (p ProfileService) GetPrfl(ctx context.Context, fltr filters.Filterer) (model.Profile, error) {
	uidFltr, ok := fltr.(filters.UuidFilter)
	if !ok {
		return model.Profile{}, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	prfl, err := p.repo.Fetch(ctx, uidFltr)
	if err != nil {
		return model.Profile{}, err
	}

	profile, ok := prfl.(model.Profile)
	if !ok {
		return model.Profile{}, cmerr.ErrUnexpectedData{Wanted: model.Profile{}, Got: prfl}
	}

	return profile, nil
}

func (p ProfileService) GetPrfls(ctx context.Context, fltr filters.Filter) ([]model.Profile, error) {
	prfls, err := p.repo.FetchAll(ctx, fltr)

	if err != nil {
		return nil, err
	}

	return prfls, nil
}

func (p ProfileService) CreatePrfl(ctx context.Context, pdto dtos.ProfileSignUpRequest) (model.Profile, error) {
	pfl := newSignupRequest(pdto)
	var err error

	pfl.Uid, err = uuid.NewRandom()
	if err != nil {
		return model.Profile{}, err
	}

	newPrfl, err := p.repo.Insert(ctx, pfl)
	if err != nil {
		return model.Profile{}, err
	}

	profile, ok := newPrfl.(model.Profile)
	if !ok {
		return model.Profile{}, cmerr.ErrUnexpectedData{Wanted: model.Profile{}, Got: newPrfl}
	}

	return profile, nil
}

func (p ProfileService) ModifyProfile(ctx context.Context, pdto dtos.ProfileRequest) (model.Profile, error) {
	pf := newProfile(pdto)

	if pf.Uid == uuid.Nil {
		return model.Profile{}, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Update", Err: nil}
	}

	prfl, err := p.repo.Update(ctx, pf)
	if err != nil {
		return model.Profile{}, err
	}

	profile, ok := prfl.(model.Profile)
	if !ok {
		return model.Profile{}, cmerr.ErrUnexpectedData{Wanted: model.Profile{}, Got: prfl}
	}

	return profile, nil
}

func (p ProfileService) DeletePrfl(ctx context.Context, delReq dtos.DeleteRequest) error {
	uid, _ := uuid.Parse(delReq.Identifer)
	err := p.repo.Delete(ctx, model.Profile{Uid: uid})
	if err != nil {
		return err
	}
	return nil
}

func newProfile(pdto dtos.ProfileRequest) *model.Profile {
	return &model.Profile{
		Id:         pdto.Id,
		Uid:        utils.ParseUuid(pdto.Uid),
		FirstName:  pdto.FirstName,
		LastName:   pdto.LastName,
		Email:      pdto.Email,
		Phone:      pdto.Phone,
		Username:   pdto.Username,
		Password:   pdto.Password,
		IsActive:   pdto.IsActive,
		AvatarFile: pdto.AvatarFile,
		CreatedAt:  pdto.CreatedAt,
		UpdatedAt:  pdto.UpdatedAt,
	}
}

func newSignupRequest(data dtos.ProfileSignUpRequest) *model.Profile {
	return &model.Profile{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Phone:     data.Phone,
		Username:  data.Username,
		Password:  data.Password,
	}
}
