package worker

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

type WorkerHandler struct {
	WorkerService
}

func NewHandler(service WorkerService) WorkerHandler {
	return WorkerHandler{
		WorkerService: service,
	}
}

func (w WorkerHandler) HandlerName() string {
	return "Worker"
}

func (wk WorkerHandler) HandleCreateWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		wk.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeoutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		payload := &dtos.WorkerDto{}

		if err := utils.ParseJSON(r, payload); err != nil {
			wk.logger.LogFields(logrus.Fields{"msg": "failed to parse request body", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			log.Printf("failed to parse reqiest %v", err)
			return
		}

		worker, err := wk.CreateWorker(r.Context(), payload)

		if err != nil {
			wk.logger.LogFields(logrus.Fields{"msg": "failed to create worker", "err": err})
			log.Printf("failed to create worker db err %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"worker":  worker,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			wk.logger.LogFields(logrus.Fields{"msg": "faild to write json response", "err": err})
			log.Printf("failed to write json response %v", err)
		}
	}
}

func (wk WorkerHandler) HandleGetWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeOutErr := utils.ErrRequestTimeout{Request: r}
		wk.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeOutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeOutErr)
	default:
		fltr, err := filters.GenFilter(r)

		if err != nil {
			log.Printf("failed to gen filter %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		wrk, err := wk.GetWorker(r.Context(), fltr)

		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"worker":  nil,
					"success": true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			wk.logger.LogFields(logrus.Fields{"msg": "database err", "err": err})
			log.Printf("database err %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"worker":  wrk,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusInternalServerError, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (wk WorkerHandler) HandleGetWorkers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeOutErr := utils.ErrRequestTimeout{Request: r}
		wk.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeOutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeOutErr)
	default:
		fltr, err := filters.GenFilter(r)

		if err != nil {
			wk.logger.LogFields(logrus.Fields{"msg": "failed to generate fileter", "err": err})
			log.Printf("failed to make filter %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		workers, err := wk.repo.FetchAll(r.Context(), fltr)

		if err != nil {
			var noResults *dac.ErrNoResults
			if errors.As(err, noResults) {
				log.Println("no workers found")
				res := ztype.JsonResponse{
					"workers": make([]dtos.WorkerDto, 0),
					"success": true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
				}
			}
			log.Printf("database err failed to fetch workers %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"workers": workers,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (wk WorkerHandler) HandleUpdateWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		wk.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		payload := &dtos.WorkerDto{}

		if err := payload.Validate(); err != nil {
			log.Printf("failed to validate request")
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := utils.ParseJSON(r, payload); err != nil {
			log.Printf("failed to parse request body %v", err)
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		worker, err := wk.ModifyWorker(r.Context(), payload)

		if err != nil {
			log.Printf("database error failed to update worker %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"worker":  worker,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (wk WorkerHandler) HandleDeleteWorker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		wk.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		fltr := filters.NewIdFilter(r)
		err := wk.DeleteWorker(r.Context(), fltr)

		if err != nil {
			log.Printf("database err failed to delete worker %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"workerId": fltr.Identifier,
			"success":  err == nil,
		}

		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed writing json response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
