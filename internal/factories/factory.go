package factories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/services"
	"github.com/Z3DRP/lessor-service/internal/services/alssr"
	"github.com/Z3DRP/lessor-service/internal/services/property"
	"github.com/Z3DRP/lessor-service/internal/services/usr"
)

func ServiceFactory(serviceName string, store dac.Store, logger *crane.Zlogrus) (services.Service, error) {
	switch strings.ToLower(serviceName) {
	case "alessor":
		repo := dac.InitAlsrRepo(store)
		return alssr.NewAlsrService(repo, logger), nil
	case "user":
		repo := dac.InitUsrRepo(store)
		return usr.NewUserService(repo, logger), nil
	case "property":
		repo := dac.InitPrptyRepo(store)
		actor, err := api.NewS3Actor(context.TODO(), serviceName)
		if err != nil {
			return nil, err
		}
		return property.NewPropertyService(repo, actor, logger), nil
	default:
		return nil, nil
	}
}

func HandlerFactory(handlerName string, service services.Service) (services.Handler, error) {
	switch strings.ToLower(handlerName) {
	case "alessor":
		alsSrvc, ok := service.(alssr.AlessorService)
		if !ok {
			return nil, errors.New("incorrect service passed to alessor handler")
		}
		return alssr.NewHandler(alsSrvc), nil
	case "user":
		usrSrvc, ok := service.(usr.UserService)
		if !ok {
			return nil, errors.New("incorrect service passed to user handler")
		}
		return usr.NewHandler(usrSrvc), nil
	case "property":
		pSrvc, ok := service.(property.PropertyService)
		if !ok {
			return nil, errors.New("incorrect service passed to property handler")
		}
		return property.NewHandler(pSrvc), nil
	default:
		return nil, fmt.Errorf("handler not found for %v", handlerName)
	}
}
