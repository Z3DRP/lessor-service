package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

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

type TaskService struct {
	repo   dac.TaskRepo
	logger *crane.Zlogrus
	//s3Actor api.FilePersister
}

func (t TaskService) ServiceName() string {
	return "Task"
}

func NewTaskService(repo dac.TaskRepo, logr *crane.Zlogrus) TaskService {
	return TaskService{
		repo: repo,
		//s3Actor: actr,
		logger: logr,
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

	// if tsk.Image != "" {
	// 	fileUrl, err := t.s3Actor.Get(ctx, task.LessorId.String(), task.Tid.String(), task.Image)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	tskDto = dtos.NewTaskResponse(task, &fileUrl)
	// }
	tskDto := dtos.NewTaskResposne(&task, nil)

	return &tskDto, nil
}

func (t TaskService) GetTasks(ctx context.Context, fltr filters.Filterer) ([]dtos.TaskResponse, error) {
	var tskReponses []dtos.TaskResponse
	filter, ok := fltr.(filters.Filter)

	if !ok {
		log.Println("invlaid filter assert ")
		return nil, filters.NewFailedToMakeFilterErr("uuid filter")
	}

	log.Println("calling service repo method")
	tasks, err := t.repo.FetchAll(ctx, filter)

	if err != nil {
		log.Printf("db err in service %v", err)
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
		tskReponses = append(tskReponses, dtos.NewTaskResposne(&tk, nil))
	}

	log.Println("returning from service")

	return tskReponses, nil
}

// when add images need to pass in the file data here
func (t TaskService) CreateTask(ctx context.Context, tdata *dtos.TaskRequest) (*dtos.TaskResponse, error) {
	tsk := newTaskFrmPtrRequest(tdata)
	log.Println("created task from request")
	var err error

	tsk.Tid, err = uuid.NewRandom()
	if err != nil {
		log.Printf("failed to create tid %v", err)
		return nil, err
	}

	log.Printf("created tid %v", tsk.Tid)
	// if fileData != nil && fileData.File != nil && fileData.Header != nil {
	// 	var fileName string
	// 	fileName, err = t.s3Actor.Upload(ctx, tsk.LessorId.String(), tsk.Tid.String(), fileData)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	tsk.Image = fileName
	// }

	nwTsk, err := t.repo.Insert(ctx, tsk)
	log.Printf("created new task %#v", nwTsk)

	if err != nil {
		return nil, err
	}

	tk, ok := nwTsk.(*model.Task)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: &model.Task{}, Got: nwTsk}
	}

	log.Println("type assertion passed")

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
	log.Printf("returning response %#v", response)
	return &response, nil
}

// when refactoring to use images will have to pass in *ztype.FileUploadDto
func (t TaskService) ModifyTask(ctx context.Context, tdto *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	tsk := newTaskFrmModRequest(tdto)

	if tsk.Tid == uuid.Nil {
		log.Printf("could not parse task pid as uuid")
		return nil, services.ErrInvalidRequest{ServiceType: t.ServiceName(), RequestType: "update"}
	}

	updatedTask, err := t.repo.Update(ctx, tsk)

	if err != nil {
		log.Printf("fialed to update task %v", err)
		return nil, err
	}

	task, ok := updatedTask.(*model.Task)

	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: updatedTask}
		log.Printf("type assertion failed %v", err)
		return nil, err
	}

	response := dtos.NewTaskResposne(task, nil)
	return &response, nil
}

func (t TaskService) ModifyTaskPririty(ctx context.Context, tdo *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	pid, err := uuid.Parse(tdo.Tid)
	if err != nil {
		return nil, fmt.Errorf("invalid task id %v", err)
	}

	tsk, err := t.repo.UpdatePriority(ctx, model.Task{PropertyId: pid, Priority: model.PriorityLevel(tdo.Priority)})
	if err != nil {
		return nil, fmt.Errorf("failed to modify priority %v", err)
	}

	task, ok := tsk.(model.Task)
	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
		log.Printf("type assertion failed %v", err)
		return nil, err
	}

	response := dtos.NewTaskResposne(&task, nil)
	return &response, nil
}

func (t TaskService) AssignTask(ctx context.Context, tdo *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	pid, err := uuid.Parse(tdo.Tid)
	if err != nil {
		return nil, fmt.Errorf("invalid task id %v", err)
	}

	wid, err := uuid.Parse(tdo.WorkerId)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	tsk, err := t.repo.UpdateStartedAt(ctx, model.Task{PropertyId: pid, WorkerId: wid, StartedAt: time.Now()})
	if err != nil {
		return nil, fmt.Errorf("failed to modify priority %v", err)
	}

	task, ok := tsk.(model.Task)
	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
		log.Printf("type assertion failed %v", err)
		return nil, err
	}

	response := dtos.NewTaskResposne(&task, nil)
	return &response, nil
}

func (t TaskService) CompleteTask(ctx context.Context, tdo *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	pid, err := uuid.Parse(tdo.Tid)
	if err != nil {
		return nil, fmt.Errorf("invalid task id %v", err)
	}

	tsk, err := t.repo.UpdateCompletedAt(ctx, model.Task{PropertyId: pid, CompletedAt: time.Now()})
	if err != nil {
		return nil, fmt.Errorf("failed to modify priority %v", err)
	}

	task, ok := tsk.(model.Task)
	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
		log.Printf("type assertion failed %v", err)
		return nil, err
	}

	response := dtos.NewTaskResposne(&task, nil)
	return &response, nil
}

func (t TaskService) PauseTask(ctx context.Context, tdo *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	pid, err := uuid.Parse(tdo.Tid)
	if err != nil {
		return nil, fmt.Errorf("invalid task id %v", err)
	}

	tsk, err := t.repo.UpdatePausedAt(ctx, model.Task{PropertyId: pid, StartedAt: time.Now()})
	if err != nil {
		return nil, fmt.Errorf("failed to modify priority %v", err)
	}

	task, ok := tsk.(model.Task)
	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
		log.Printf("type assertion failed %v", err)
		return nil, err
	}

	response := dtos.NewTaskResposne(&task, nil)
	return &response, nil
}

func (t TaskService) UnPauseTask(ctx context.Context, tdo *dtos.TaskModRequest) (*dtos.TaskResponse, error) {
	pid, err := uuid.Parse(tdo.Tid)
	if err != nil {
		return nil, fmt.Errorf("invalid task id %v", err)
	}

	var nilTime time.Time

	tsk, err := t.repo.UpdatePausedAt(ctx, model.Task{PropertyId: pid, StartedAt: time.Now(), PausedAt: nilTime})
	if err != nil {
		return nil, fmt.Errorf("failed to modify priority %v", err)
	}

	task, ok := tsk.(model.Task)
	if !ok {
		err = cmerr.ErrUnexpectedData{Wanted: model.Task{}, Got: tsk}
		log.Printf("type assertion failed %v", err)
		return nil, err
	}

	response := dtos.NewTaskResposne(&task, nil)
	return &response, nil
}

func (t TaskService) UpdatePriorities(ctx context.Context, tdata []*dtos.TaskModRequest) (*[]dtos.TaskResponse, error) {
	tasks := make([]any, len(tdata))
	for _, tres := range tdata {
		tasks = append(tasks, *newTaskFrmModRequest(tres))
	}

	tsks, err := t.repo.BulkPriorityUpdate(ctx, tasks)

	if err != nil {
		return nil, err
	}

	response := dtos.NewTaskResponseList(tsks)
	return &response, nil
}

func (t TaskService) DeleteTask(ctx context.Context, f filters.Filterer) error {
	fltr, ok := f.(filters.IdFilter)

	if !ok {
		return errors.New("fialed to create id filter")
	}

	if err := fltr.Validate(); err != nil {
		return fmt.Errorf("invalid request %v", err)
	}

	tid, _ := uuid.Parse(fltr.Identifier)
	err := t.repo.Delete(ctx, model.Task{Tid: tid})

	if err != nil {
		return err
	}

	return nil
}

func NewTaskFrmRequest(data dtos.TaskRequest) *model.Task {
	return &model.Task{
		LessorId:      utils.ParseUuid(data.LessorId),
		PropertyId:    utils.ParseUuid(data.PropertyId),
		WorkerId:      utils.ParseUuid(data.WorkerId),
		Name:          data.Name,
		Category:      model.TaskCategory(data.Category),
		Priority:      model.PriorityLevel(data.Priority),
		Details:       data.Details,
		Notes:         data.Notes,
		ScheduledAt:   data.ScheduledAt,
		ActualCost:    data.ActualCost,
		EstimatedCost: data.EstimateCost,
		Image:         data.Image,
	}
}

func newTaskFrmPtrRequest(data *dtos.TaskRequest) *model.Task {
	return &model.Task{
		LessorId:      utils.ParseUuid(data.LessorId),
		PropertyId:    utils.ParseUuid(data.PropertyId),
		WorkerId:      utils.ParseUuid(data.WorkerId),
		Category:      model.TaskCategory(data.Category),
		Name:          data.Name,
		Priority:      model.PriorityLevel(data.Priority),
		Details:       data.Details,
		Notes:         data.Notes,
		ScheduledAt:   data.ScheduledAt,
		EstimatedCost: data.EstimateCost,
		ActualCost:    data.ActualCost,
		Image:         data.Image,
	}
}

func newTaskFrmModRequest(data *dtos.TaskModRequest) *model.Task {
	return &model.Task{
		LessorId:      utils.ParseUuid(data.LessorId),
		Tid:           utils.ParseUuid(data.Tid),
		Name:          data.Name,
		PropertyId:    utils.ParseUuid(data.PropertyId),
		Category:      model.TaskCategory(data.Category),
		WorkerId:      utils.ParseUuid(data.WorkerId),
		Priority:      model.PriorityLevel(data.Priority),
		Details:       data.Details,
		Notes:         data.Notes,
		ScheduledAt:   data.ScheduledAt,
		StartedAt:     data.StartedAt,
		CompletedAt:   data.CompletedAt,
		PausedAt:      data.PausedAt,
		PausedReason:  data.PausedReason,
		FailedAt:      data.FailedAt,
		FailedReason:  data.FailedReason,
		EstimatedCost: data.EstimatedCost,
		ActualCost:    data.ActualCost,
		Image:         data.Image,
	}
}
