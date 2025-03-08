package task

import (
	"context"
	"database/sql"

	"github.com/Z3DRP/alessor-service/pkg/utils"
	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/google/uuid"
)

type TaskService struct {
	repo    dac.TaskRepo
	logger  *crane.Zlogrus
	s3Actor api.FilePersister
}

func (t TaskService) ServiceName() string {
	return "Task"
}

func NewTaskService(repo dac.TaskRepo, actr api.S3Actor, logr *crane.Zlogrus) TaskService {
	return TaskService{
		repo:    repo,
		s3Actor: actr,
		logger:  logr,
	}
}

func (t TaskService) GetTask(ctx context.Context, fltr filters.Filterer) (*dtos.TaskResponse, error) {
	uidFilter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	tsk, err := t.repo.Fetch(ctx, uidFilter)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	task, ok := tsk.(model.Task)

	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
	}

	var tskDto dtos.TaskResponse

	// if tsk.Image != "" {
	// 	fileUrl, err := t.s3Actor.Get(ctx, task.LessorId.String(), task.Tid.String(), task.Image)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	tskDto = dtos.NewTaskResponse(task, &fileUrl)
	// }
	tskDto = dtos.NewTaskResposne(task, nil)

	return &tskDto, nil
}

func (t TaskService) GetTasks(ctx context.Context, fltr filters.Filterer) ([]dtos.TaskResponse, error) {
	var tskReponses []dtos.TaskResponse
	filter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	tasks, err := t.repo.FetchAll(ctx, filter)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, dac.ErrNoResults{Err: err, Shape: tskReponses, Identifier: "all"}
		}
		return nil, err
	}

	// imgUrls, err := t.s3Actor.List(ctx, filter.Identifier)

	// if err != nil {
	// 	if err == api.ErrrNoImagesFound {
	// 		for _, tk := range tasks {
	// 			tskReponses = append(tskReponses, dtos.NewTaskResposne(tk, nil))
	// 		}
	// 	}
	// }

	for _, tk := range tasks {
		// if url, found := imgUrls[tk.Image]; found {
		// 	tskReponses = append(tskReponses, dtos.NewTaskResposne(tk, &url))
		// } else {
		// 	tskReponses = append(tskReponses, dtos.NewTaskResposne(tk, nil))
		// }
		tskReponses = append(tskReponses, dtos.NewTaskResposne(tk, nil))
	}

	return tskReponses, nil
}

// when add images need to pass in the file data here
func (t TaskService) CreateTask(ctx context.Context, tdata *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	tsk := newTaskFrmRequest(tdata)
	var err error

	tsk.Tid, err = uuid.NewRandom()

	if err != nil {
		return nil, err
	}

	// if fileData != nil && fileData.File != nil && fileData.Header != nil {
	// 	var fileName string
	// 	fileName, err = t.s3Actor.Upload(ctx, tsk.LessorId.String(), tsk.Tid.String(), fileData)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	tsk.Image = fileName
	// }

	nwTsk, err := t.repo.Insert(ctx, tsk)

	if err != nil {
		return nil, err
	}

	tk, ok := nwTsk.(*model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: &model.Task{}, Got: nwTsk}
	}

	// var psUrl *string
	// if tk.Image != "" {
	// 	url, err := t.s3Actor.GetFile(ctx, tk.Image)

	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	psUrl = &url
	// }

	// here you will have to pass in the presign url instead of nil
	response := dtos.NewTaskResposneFrmPntr(tk, nil)
	return &response, nil
}

func newTaskFrmRequest(data dtos.TaskRequest) *model.Task {
	return &model.Task{
		LessorId:    utils.ParseUuid(data.LessorId),
		PropertyId:  utils.ParseUuid(data.PropertyId),
		WorkerId:    utils.ParseUuid(data.WorkerId),
		Details:     data.Details,
		Notes:       data.Notes,
		ScheduledAt: data.ScheduledAt,
		Image:       data.Image,
	}
}

func newTaskFrmPtrRequest(data *dtos.TaskRequest) *model.Task {
	return &model.Task{
		LessorId:    utils.ParseUuid(data.LessorId),
		PropertyId:  utils.ParseUuid(data.PropertyId),
		WorkerId:    utils.ParseUuid(data.WorkerId),
		Details:     data.Details,
		Notes:       data.Notes,
		ScheduledAt: data.ScheduledAt,
		Image:       data.Image,
	}
}

func newTaskFrmModRequest(data *dtos.TaskModRequest) *model.Task {
	return &model.Task{
		LessorId:     utils.ParseUuid(data.LessorId),
		Tid:          utils.ParseUuid(data.Tid),
		PropertyId:   utils.ParseUuid(data.PropertyId),
		WorkerId:     utils.ParseUuid(data.WorkerId),
		Details:      data.Details,
		Notes:        data.Notes,
		ScheduledAt:  data.ScheduledAt,
		StartedAt:    data.StartedAt,
		CompletedAt:  data.CompletedAt,
		PausedAt:     data.PausedAt,
		PausedReason: data.PausedReason,
		FailedAt:     data.FailedAt,
		FailedReason: data.FailedReason,
		Image:        data.Image,
	}
}
