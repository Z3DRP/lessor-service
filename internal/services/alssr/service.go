package alssr

import (
	"context"
	"errors"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type AlessorService struct {
	repo   dac.AlessorRepo
	logger *crane.Zlogrus
}

func (a AlessorService) ServiceName() string {
	return "Alessor"
}

func NewAlsrService(repo dac.AlessorRepo, logr *crane.Zlogrus) AlessorService {
	return AlessorService{
		repo:   repo,
		logger: logr,
	}
}

func (a *AlessorService) GetAlsr(ctx context.Context, fltr filters.Filterer) (model.Alessor, error) {
	uidFltr, ok := fltr.(filters.Filter)
	if !ok {
		return model.Alessor{}, filters.NewFailedToMakeFilterErr("uuid")
	}

	alsr, err := a.repo.Fetch(ctx, uidFltr)

	if err != nil {
		return model.Alessor{}, err
	}

	alessor, ok := alsr.(model.Alessor)
	if !ok {
		return model.Alessor{}, cmerr.ErrUnexpectedData{Wanted: model.Alessor{}, Got: alsr}
	}
	return alessor, nil
}

func (a *AlessorService) GetAlsrs(ctx context.Context, fltr filters.Filter) ([]model.Alessor, error) {
	alsrs, err := a.repo.FetchAll(ctx, fltr)
	if err != nil {
		return nil, err
	}

	return alsrs, nil
}

func (a *AlessorService) CreateAlsr(ctx context.Context, adto dtos.AlessorRequest) (model.Alessor, error) {
	al := newAlessor(adto)

	if al.Uid == uuid.Nil {
		return model.Alessor{}, errors.New("missing profile id")
	}

	alsr, err := a.repo.Insert(ctx, al)
	if err != nil {
		return model.Alessor{}, err
	}

	alessor, ok := alsr.(model.Alessor)
	if !ok {
		return model.Alessor{}, cmerr.ErrUnexpectedData{Wanted: model.Alessor{}, Got: alsr}
	}
	return alessor, nil
}

func (a *AlessorService) ModifyAlsr(ctx context.Context, adto dtos.AlessorRequest) (model.Alessor, error) {
	al := newAlessor(adto)

	if al.Uid == uuid.Nil {
		return model.Alessor{}, errors.New("missing profile id")
	}

	alsr, err := a.repo.Update(ctx, al)
	if err != nil {
		return model.Alessor{}, err
	}

	alessor, ok := alsr.(model.Alessor)
	if !ok {
		return model.Alessor{}, cmerr.ErrUnexpectedData{Wanted: model.Alessor{}, Got: alsr}
	}

	return alessor, nil
}

func (a *AlessorService) DeleteAlsr(ctx context.Context, delRequest dtos.DeleteRequest) error {
	uid, err := uuid.Parse(delRequest.Identifer)
	if err != nil {
		return err
	}

	err = a.repo.Delete(ctx, model.Alessor{Uid: uid})
	if err != nil {
		return err
	}

	return nil
}

func newAlessor(adto dtos.AlessorRequest) *model.Alessor {
	return &model.Alessor{
		Id:                        adto.Id,
		Uid:                       utils.ParseUuid(adto.Uid),
		TotalProperties:           adto.TotalProperties,
		SquareAccount:             adto.SquareAccount,
		PaymentIntegrationEnabled: adto.PaymentIntegrationEnabled,
		PaymentSchedule:           adto.PaymentSchedule,
		ComunicationPreference:    model.CommunicationPreference(adto.CommunicationPrefrences),
	}
}
