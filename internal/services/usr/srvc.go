package usr

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Z3DRP/lessor-service/internal/auth"
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

type UserService struct {
	repo   dac.UserRepo
	logger *crane.Zlogrus
}

func (p UserService) ServiceName() string {
	return "User"
}

func NewUserService(repo dac.UserRepo, logr *crane.Zlogrus) UserService {
	return UserService{
		repo:   repo,
		logger: logr,
	}
}

func (u UserService) AuthenticateUser(ctx context.Context, fltr filters.Filterer) (bool, model.User, error) {
	credentials, ok := fltr.(filters.Creds)

	if !ok {
		return false, model.User{}, filters.NewFailedToMakeFilterErr("credential filter")
	}
	log.Println("auth srvc valid creds")

	usr, err := u.repo.GetCredentials(ctx, credentials.Email)

	if err != nil {
		log.Printf("failed to get credentials %v", err)
		return false, model.User{}, err
	}

	if usr == nil {
		return false, model.User{}, errors.New("invalid credentials")
	}

	user, ok := usr.(model.User)
	if !ok {
		return false, model.User{}, cmerr.ErrUnexpectedData{Wanted: model.User{}, Got: usr}
	}

	isMatch, err := auth.VerifyHash(user.Password, credentials.Password)
	if err != nil {
		return false, model.User{}, err
	}

	return isMatch, user, nil
}

func (p UserService) GetUsr(ctx context.Context, fltr filters.Filterer) (model.User, error) {
	uidFltr, ok := fltr.(filters.Filter)
	if !ok {
		return model.User{}, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	prfl, err := p.repo.Fetch(ctx, uidFltr)
	if err != nil {
		return model.User{}, err
	}

	profile, ok := prfl.(model.User)
	if !ok {
		return model.User{}, cmerr.ErrUnexpectedData{Wanted: model.User{}, Got: prfl}
	}

	return profile, nil
}

func (p UserService) GetUsrs(ctx context.Context, fltr filters.Filter) ([]model.User, error) {
	prfls, err := p.repo.FetchAll(ctx, fltr)

	if err != nil {
		return nil, err
	}

	return prfls, nil
}

func (u UserService) CreateUsr(ctx context.Context, udata dtos.UserSignupRequest) (*model.User, error) {
	usr := newSignupRequest(udata)
	var err error

	usr.Uid, err = uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	hashPwd, err := auth.HashString(usr.Password)
	if err != nil {
		return nil, fmt.Errorf("could not create user safely, %v", err)
	}

	usr.Password = hashPwd
	usr.IsActive = true

	newUsr, err := u.repo.Insert(ctx, usr)
	if err != nil {
		return nil, err
	}

	user, ok := newUsr.(*model.User)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: &model.User{}, Got: newUsr}
	}

	return user, nil
}

func (p UserService) ModifyUser(ctx context.Context, pdto dtos.UserRequest) (model.User, error) {
	pf := newUser(pdto)

	if pf.Uid == uuid.Nil {
		return model.User{}, services.ErrInvalidRequest{ServiceType: p.ServiceName(), RequestType: "Update", Err: nil}
	}

	prfl, err := p.repo.Update(ctx, pf)
	if err != nil {
		return model.User{}, err
	}

	profile, ok := prfl.(model.User)
	if !ok {
		return model.User{}, cmerr.ErrUnexpectedData{Wanted: model.User{}, Got: prfl}
	}

	return profile, nil
}

func (p UserService) DeleteUsr(ctx context.Context, delReq dtos.DeleteRequest) error {
	uid, _ := uuid.Parse(delReq.Identifer)
	err := p.repo.Delete(ctx, model.User{Uid: uid})
	if err != nil {
		return err
	}
	return nil
}

func newUser(pdto dtos.UserRequest) *model.User {
	return &model.User{
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

func newSignupRequest(data dtos.UserSignupRequest) *model.User {
	return &model.User{
		FirstName:   data.FirstName,
		LastName:    data.LastName,
		Email:       data.Email,
		Phone:       data.Phone,
		ProfileType: data.ProfileType,
		Username:    data.Username,
		Password:    data.Password,
	}
}
