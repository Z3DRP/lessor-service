package alssr

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
)

type AlessorHandler struct {
	AlessorService
}

func NewHandler(service AlessorService) AlessorHandler {
	return AlessorHandler{AlessorService: service}
}

func (a AlessorHandler) HandlerName() string {
	return "Alessor"
}

func (a AlessorHandler) HandleCreateAlessor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		a.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		var payload dtos.AlessorRequest
		if err := utils.ParseJSON(r, payload); err != nil {
			a.logger.MustDebug(fmt.Sprintf("failed to parse request body %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			a.logger.MustDebug(fmt.Sprintf("alessor create payload failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		alessor, err := a.CreateAlsr(r.Context(), payload)
		log.Println("returned from service")
		if err != nil {
			a.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		log.Println("callilng println on alsr type")

		log.Println(fmt.Printf("Type of alsr: %T", alessor))
		res := ztype.JsonResponse{
			"alessor": alessor,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			a.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (a AlessorHandler) HandleGetAlessor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		// w.WriteHeader(http.StatusRequestTimeout)
		// errMsg := map[string]string{
		// 	"error":  "request timeout",
		// 	"status": fmt.Sprintf("%v", http.StatusRequestTimeout),
		// }
		// encoder := json.NewEncoder(w)

		// err := encoder.Encode(errMsg)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		a.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			a.logger.MustDebug(fmt.Sprintf("error: %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		alessor, err := a.GetAlsr(r.Context(), fltr)
		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"alessor": nil,
					"success": true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			a.logger.MustDebug(fmt.Sprintf("failed to fetch alessor, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"alessor": alessor,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			a.logger.MustDebug(fmt.Sprintf("internal server error, %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (a AlessorHandler) HandleGetAlessors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		a.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"alessors": make([]model.Alessor, 0),
					"success":  true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			a.logger.MustDebug(fmt.Sprintf("failed create alessor query filter, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		alsrs, err := a.GetAlsrs(r.Context(), fltr)
		if err != nil {
			a.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"alessors": alsrs,
			"success":  true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (a AlessorHandler) HandleUpdateAlessor(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		a.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)

	default:
		var alsrPayload dtos.AlessorRequest
		if err := utils.ParseJSON(r, alsrPayload); err != nil {
			a.logger.MustDebug(fmt.Sprintf("failed to parse alessor dto from request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := alsrPayload.Validate(); err != nil {
			a.logger.MustDebug(utils.FormatErrMsg("alessor dto failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		updatedAlsr, err := a.ModifyAlsr(r.Context(), alsrPayload)
		if err != nil {
			a.logger.MustDebug(utils.FormatErrMsg("database err, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"alessor": updatedAlsr,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			a.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (a AlessorHandler) HandleDeleteAlessor(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		a.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)

	default:
		delReq, err := dtos.BuildDeleteRequest(r)
		if err != nil {
			a.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err = delReq.Validate(); err != nil {
			a.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, fmt.Errorf("invalid delete request, %v", err))
			return
		}

		req, ok := delReq.(dtos.DeleteRequest)
		if !ok {
			invalidData := cmerr.ErrUnexpectedData{Wanted: dtos.DeleteRequest{}, Got: delReq}
			a.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", invalidData))
			utils.WriteErr(w, http.StatusBadRequest, errors.New("invalid request"))
			return
		}

		if err := a.DeleteAlsr(r.Context(), req); err != nil {
			a.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"success": true,
		}
		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			a.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
