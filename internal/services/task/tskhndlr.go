package task

import (
	"errors"
	"log"
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/sirupsen/logrus"
)

type TaskHandler struct {
	TaskService
}

func NewHandler(service TaskService) TaskHandler {
	return TaskHandler{
		TaskService: service,
	}
}

func (t TaskHandler) HandlerName() string {
	return "Task"
}

func (t TaskHandler) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeoutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		payload := &dtos.TaskRequest{}

		if err := utils.ParseJSON(r, payload); err != nil {
			t.logger.LogFields(logrus.Fields{"msg": "failed to parse request body", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			log.Printf("failed to parse request %v", err)
			log.Printf("request body failed: %v", r.Body)
			return
		}

		log.Println()
		log.Printf("payload %#v\n", payload)

		task, err := t.CreateTask(r.Context(), payload)

		if err != nil {
			t.logger.LogFields(logrus.Fields{"msg": "failed to create task", "err": err})
			log.Printf("failed to create task db err %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		log.Printf("creat task success %#v", task)

		log.Println("fetching new task data")

		nwTask, err := t.repo.Fetch(r.Context(), filters.Filter{Identifier: task.Tid, Page: 1, Limit: 1})

		if err != nil {
			t.logger.LogFields(logrus.Fields{"msg": "failed to fetch newest task", "err": err})
			log.Printf("failed to fetch new task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    nwTask,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			t.logger.LogFields(logrus.Fields{"msg": "faild to write json response", "err": err})
			log.Printf("failed to write json response %v", err)
		}
	}
}

func (t TaskHandler) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeOutErr := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeOutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeOutErr)
	default:
		fltr, err := filters.GenFilter(r)

		if err != nil {
			log.Printf("failed to gen filter %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		tsk, err := t.GetTask(r.Context(), fltr)

		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"task":    nil,
					"success": true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			t.logger.LogFields(logrus.Fields{"msg": "database err", "err": err})
			log.Printf("database err %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    tsk,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusInternalServerError, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (t TaskHandler) HandleGetTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeOutErr := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeOutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeOutErr)
	default:
		log.Println()
		log.Println("fetching tasks")
		fltr, err := filters.GenFilter(r)

		if err != nil {
			t.logger.LogFields(logrus.Fields{"msg": "failed to generate fileter", "err": err})
			log.Printf("failed to make filter %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		log.Println("about to call service db method")
		tasks, err := t.repo.FetchAll(r.Context(), fltr)

		if err != nil {
			var noResults *dac.ErrNoResults
			if errors.As(err, noResults) {
				res := ztype.JsonResponse{
					"tasks":   make([]dtos.TaskResponse, 0),
					"success": true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
				}
			}
			log.Printf("database err failed to fetch tasks %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"tasks":   tasks,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (t TaskHandler) HandleUpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.TaskModRequest{}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		task, err := t.ModifyTask(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    task,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (t TaskHandler) HandleUpdatePriority(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.TaskModRequest{}

		tid := r.PathValue("id")
		if tid == "" {
			log.Println("invalid request missing tid path value")
			utils.WriteErr(w, http.StatusBadRequest, errors.New("invalid request missing tid in url"))
			return
		}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		payload.Tid = tid
		task, err := t.ModifyTask(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    task,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}

}

func (t TaskHandler) HandleAssignTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.TaskModRequest{}

		tid := r.PathValue("id")
		if tid == "" {
			log.Println("invalid request missing tid path value")
			utils.WriteErr(w, http.StatusBadRequest, errors.New("invalid request missing tid in url"))
			return
		}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		payload.Tid = tid
		task, err := t.AssignTask(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    task,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}

}

func (t TaskHandler) HandleCompleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.TaskModRequest{}

		tid := r.PathValue("id")
		if tid == "" {
			log.Println("invalid request missing tid path value")
			utils.WriteErr(w, http.StatusBadRequest, errors.New("invalid request missing tid in url"))
			return
		}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		payload.Tid = tid
		task, err := t.CompleteTask(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    task,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}

}

func (t TaskHandler) HandlePauseTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.TaskModRequest{}

		tid := r.PathValue("id")
		if tid == "" {
			log.Println("invalid request missing tid path value")
			utils.WriteErr(w, http.StatusBadRequest, errors.New("invalid request missing tid in url"))
			return
		}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		payload.Tid = tid
		task, err := t.PauseTask(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    task,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}

}

func (t TaskHandler) HandleUnPauseTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.TaskModRequest{}

		tid := r.PathValue("id")
		if tid == "" {
			log.Println("invalid request missing tid path value")
			utils.WriteErr(w, http.StatusBadRequest, errors.New("invalid request missing tid in url"))
			return
		}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		payload.Tid = tid
		task, err := t.UnPauseTask(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"task":    task,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}

}

func (t TaskHandler) HandleBulkPriorityUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := []*dtos.TaskModRequest{}
		if err := utils.ParseJSON(r, &payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		var bErr error
		for _, load := range payload {
			if err := load.Validate(); err != nil {
				log.Printf("failed to validate request %v", err)
				bErr = err
				break
			}
		}

		if bErr != nil {
			utils.WriteErr(w, http.StatusBadRequest, bErr)
		}

		tasks, err := t.UpdatePriorities(r.Context(), payload)

		if err != nil {
			log.Printf("db error %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"tasks":   tasks,
			"success": true,
		}

		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed to write resposne %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (t TaskHandler) HandleDeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		t.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		fltr := filters.NewIdFilter(r)
		err := t.DeleteTask(r.Context(), fltr)

		if err != nil {
			log.Printf("database err failed to delete task %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"taskId":  fltr.Identifier,
			"success": err == nil,
		}

		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
