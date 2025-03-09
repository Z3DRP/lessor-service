package factories

import (
	"context"
	"fmt"
	"strings"

	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/services"
	"github.com/Z3DRP/lessor-service/internal/services/alssr"
	"github.com/Z3DRP/lessor-service/internal/services/property"
	rentalproperty "github.com/Z3DRP/lessor-service/internal/services/rentalProperty"
	"github.com/Z3DRP/lessor-service/internal/services/task"
	"github.com/Z3DRP/lessor-service/internal/services/usr"
	"github.com/Z3DRP/lessor-service/internal/services/worker"
)

func ServiceFactory(serviceName string, store dac.Persister, logger *crane.Zlogrus) (services.Service, error) {
	switch strings.ToLower(serviceName) {
	case "alessor":
		repo := dac.InitAlsrRepo(store)
		return alssr.NewAlsrService(repo, logger), nil
	case "user":
		repo := dac.InitUsrRepo(store)
		return usr.NewUserService(repo, logger), nil
	case "property":
		// needs to update to use actor inbox to send msg and init actor
		repo := dac.InitPrptyRepo(store)
		s3Dir, err := ServiceS3Dir(serviceName)

		if err != nil {
			return nil, err
		}

		actor, err := api.NewS3Actor(context.TODO(), s3Dir)

		if err != nil {
			return nil, err
		}
		return property.NewPropertyService(repo, actor, logger), nil
	case "task":
		repo := dac.InitTskRepo(store)
		s3Dir, err := ServiceS3Dir(serviceName)

		if err != nil {
			return nil, err
		}

		actor, err := api.NewS3Actor(context.TODO(), s3Dir)

		if err != nil {
			return nil, err
		}

		return task.NewTaskService(repo, actor, logger), nil
	case "rental property":
		repo := dac.InitRentalPrptyRepo(store)
		return rentalproperty.NewRentalPropertyService(repo, logger), nil
	case "worker":
		repo := dac.InitWorkerRepo(store)
		return worker.NewWorkerService(repo, logger), nil
	default:
		return nil, nil
	}
}

func HandlerFactory(handlerName string, service services.Service) (services.Handler, error) {
	switch strings.ToLower(handlerName) {
	case "alessor":
		alsSrvc, ok := service.(alssr.AlessorService)
		if !ok {
			return nil, ErrWrongServiceInject{ServiceName: service.ServiceName(), HandlerName: "alessor"}
		}
		return alssr.NewHandler(alsSrvc), nil
	case "user":
		usrSrvc, ok := service.(usr.UserService)
		if !ok {
			return nil, ErrWrongServiceInject{ServiceName: service.ServiceName(), HandlerName: "user"}
		}
		return usr.NewHandler(usrSrvc), nil
	case "property":
		pSrvc, ok := service.(property.PropertyService)
		if !ok {
			return nil, ErrWrongServiceInject{ServiceName: service.ServiceName(), HandlerName: "property"}
		}
		return property.NewHandler(pSrvc), nil
	case "tasks":
		tskSrvc, ok := service.(task.TaskService)
		if !ok {
			return nil, ErrWrongServiceInject{ServiceName: service.ServiceName(), HandlerName: "task"}
		}
		return task.NewHandler(tskSrvc), nil
	case "rental property":
		rentalPropSrc, ok := service.(rentalproperty.RentalPropertyService)
		if !ok {
			return nil, ErrWrongServiceInject{ServiceName: service.ServiceName(), HandlerName: "rental property"}
		}
		return rentalproperty.NewHandler(rentalPropSrc), nil
	case "worker":
		workerServer, ok := service.(worker.WorkerService)
		if !ok {
			return nil, ErrWrongServiceInject{ServiceName: service.ServiceName(), HandlerName: "worker"}
		}
		return worker.NewHandler(workerServer), nil
	default:
		return nil, fmt.Errorf("handler not found for %v", handlerName)
	}
}

func ServiceS3Dir(service string) (string, error) {
	switch strings.ToLower(service) {
	case "alessor", "user":
		return "USERS_DIR", nil
	case "property":
		return "PROPERTIES_DIR", nil
	case "task":
		return "TASK_DIR", nil
	default:
		return "", fmt.Errorf("service %v does not have a s3 location", service)
	}
}

type ErrWrongServiceInject struct {
	ServiceName string
	HandlerName string
}

func (e ErrWrongServiceInject) Error() string {
	return fmt.Sprintf("incorrect service passed into %v handler wanted %v", e.ServiceName, e.HandlerName)
}

type ErrFailedServiceStart struct {
	ServiceName string
	Err         error
}

func (e ErrFailedServiceStart) Error() string {
	return fmt.Sprintf("failed to create %v service", e.ServiceName)
}

func (e ErrFailedServiceStart) Unwrap() error {
	return e.Err
}
