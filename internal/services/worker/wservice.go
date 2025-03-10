package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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

type WorkerService struct {
	repo   dac.WorkerRepo
	logger *crane.Zlogrus
	//s3Actor api.FilePersister
}

func (p WorkerService) ServiceName() string {
	return "Worker"
}

// TODO when adding images will have to pass in s3Actor
func NewWorkerService(repo dac.WorkerRepo, logr *crane.Zlogrus) WorkerService {
	return WorkerService{
		repo: repo,
		//s3Actor: actr,
		logger: logr,
	}
}

func (p WorkerService) GetWorker(ctx context.Context, fltr filters.Filterer) (*dtos.WorkerDto, error) {
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

	worker, ok := prpty.(model.Worker)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: prpty}
	}

	//var workerDto dtos.WorkerDto

	// if worker.Image != "" {
	// 	fileUrl, err := p.s3Actor.Get(ctx, worker.LessorId.String(), worker.Pid.String(), worker.Image)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	workerDto = dtos.NewWorkerDto(worker, nil)
	// }

	workerDto := dtos.NewWorkerDto(worker, nil)

	return &workerDto, nil
}

func (p WorkerService) GetProperties(ctx context.Context, fltr filters.Filterer) ([]dtos.WorkerDto, error) {
	// need to add a uuid filter for all repos because that way it limits the results in multi tenant db
	var workersRes []dtos.WorkerDto
	filter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	properties, err := p.repo.FetchAll(ctx, filter)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dac.ErrNoResults{Err: err, Shape: workersRes, Identifier: "all"}
		}
		return nil, err
	}

	//imageUrls, err := p.s3Actor.List(ctx, filter.Identifier)

	// if err != nil {
	// 	if err == api.ErrrNoImagesFound {
	// 		// no images so return properties found
	// 		for _, prop := range properties {
	// 			workersRes = append(workersRes, dtos.NewWorkerResponse(prop, nil))
	// 		}
	// 		return workersRes, nil
	// 	}
	// 	return nil, err
	// }

	for _, worker := range properties {
		// prop.Image has the entire s3 path and file key i.e. property/{ownerId}/{objId}/filename
		// if url, found := imageUrls[prop.Image]; found {
		// 	workersRes = append(workersRes, dtos.NewWorkerResponse(prop, &url))
		// } else {
		// 	workersRes = append(workersRes, dtos.NewWorkerResponse(prop, nil))
		// }
		workersRes = append(workersRes, dtos.NewWorkerDto(worker, nil))
	}

	return workersRes, nil
}

// TODO will have to add fileData fileDto back here
func (p WorkerService) CreateWorker(ctx context.Context, pdata *dtos.WorkerDto) (*dtos.WorkerDto, error) {
	worker := newWorkerRequest(pdata)
	var err error

	worker.Uid, err = uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	// if fileData != nil && fileData.File != nil && fileData.Header != nil {
	// 	var fileName string
	// 	fileName, err = p.s3Actor.Upload(ctx, worker.LessorId.String(), worker.Pid.String(), fileData)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	worker.Image = fileName
	// }

	nwWorker, err := p.repo.Insert(ctx, worker)

	if err != nil {
		return nil, err
	}

	prpty, ok := nwWorker.(*model.Worker)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: &model.Worker{}, Got: nwWorker}
	}

	// var psUrl *string
	// if worker.Image != "" {
	// 	url, err := p.s3Actor.GetFile(ctx, worker.Image)

	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	psUrl = &url
	// }

	response := dtos.NewWorkerDtoFrmPtr(prpty, nil)
	return &response, nil
}

// when add image will have to pass this back in fileData *ztype.FileUploadDto
func (w WorkerService) ModifyWorker(ctx context.Context, pdto *dtos.WorkerDto) (*dtos.WorkerDto, error) {
	wrkr := newWorkerModRequest(pdto)

	if wrkr.Uid == uuid.Nil {
		log.Printf("could not parse property pid as uuid")
		return nil, services.ErrInvalidRequest{ServiceType: w.ServiceName(), RequestType: "Upate", Err: nil}
	}

	//existingWorkerImg, err := p.getExistingWorkerImage(ctx, wrkr.Pid.String())

	// if err != nil {
	// 	log.Printf("failed to fetch existing property to check for image")
	// 	return nil, err
	// }

	// if fileData != nil && fileData.File != nil && fileData.Header != nil {
	// 	fileName, err := p.s3Actor.Upload(ctx, prpty.LessorId.String(), prpty.Pid.String(), fileData)

	// 	if err != nil {
	// 		log.Printf("error uploading file %v", err)
	// 		return nil, err
	// 	}

	// 	prpty.Image = fileName
	// } else {
	// 	prpty.Image = existingWorkerImg
	// }

	updateWrkr, err := w.repo.Update(ctx, wrkr)

	if err != nil {
		log.Printf("err updating property %v", err)
		return nil, err
	}

	worker, ok := updateWrkr.(model.Worker)

	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: updateWrkr}
		log.Printf("type assertion failed %v", err)
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Worker{}, Got: updateWrkr}
	}

	// var psUrl *string
	// if worker.Image != "" {
	// 	url, err := p.s3Actor.GetFile(ctx, worker.Image)

	// 	if err != nil {
	// 		log.Printf("failed to get file %v", err)
	// 		return nil, err
	// 	}

	// 	psUrl = &url
	// }

	response := dtos.NewWorkerDto(worker, nil)
	return &response, nil
}

func (w WorkerService) DeleteWorker(ctx context.Context, f filters.Filterer) error {
	fltr, ok := f.(filters.IdFilter)
	if !ok {
		return errors.New("failed to create id filter")
	}

	if err := fltr.Validate(); err != nil {
		return fmt.Errorf("invalid request, %v", err)
	}

	wid, _ := uuid.Parse(fltr.Identifier)
	err := w.repo.Delete(ctx, model.Worker{Uid: wid})

	if err != nil {
		return err
	}

	return nil
}

func (w WorkerService) GetExistingWorkerImage(ctx context.Context, id string) (string, error) {
	worker, err := w.repo.GetExisting(ctx, id)
	if err != nil {
		return "", nil
	}

	return worker.Image, nil
}

func newWorkerRequest(w *dtos.WorkerDto) *model.Worker {
	return &model.Worker{
		Uid:           utils.ParseUuid(w.Uid),
		StartDate:     w.StartDate,
		EndDate:       w.EndDate,
		Title:         w.Title,
		Specilization: w.Specilization,
		PayRate:       w.PayRate,
		LessorId:      utils.ParseUuid(w.LessorId),
		PaymentMethod: model.MethodOfPayment(w.PaymentMethod),
	}
}

func newWorkerModRequest(w *dtos.WorkerDto) model.Worker {
	return model.Worker{
		Uid:           utils.ParseUuid(w.Uid),
		StartDate:     w.StartDate,
		EndDate:       w.EndDate,
		Title:         w.Title,
		Specilization: w.Specilization,
		PayRate:       w.PayRate,
		LessorId:      utils.ParseUuid(w.LessorId),
		PaymentMethod: model.MethodOfPayment(w.PaymentMethod),
	}
}
