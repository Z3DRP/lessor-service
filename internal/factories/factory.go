package factories

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/services"
	"github.com/Z3DRP/lessor-service/internal/services/alssr"
	"github.com/Z3DRP/lessor-service/internal/services/usr"
)

func ServiceFactory(serviceName string, store dac.Store, logger *crane.Zlogrus) services.Service {
	switch strings.ToLower(serviceName) {
	case "alessor":
		repo := dac.InitAlsrRepo(store)
		return alssr.NewAlsrService(repo, logger)
	case "user":
		repo := dac.InitUsrRepo(store)
		return usr.NewUserService(repo, logger)
	default:
		return nil
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
			return nil, errors.New("incorrect service passed to profile handler")
		}
		return usr.NewHandler(usrSrvc), nil
	default:
		return nil, fmt.Errorf("handler not found for %v", handlerName)
	}
}
