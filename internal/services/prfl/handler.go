package prfl

import (
	"fmt"
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
)

type ProfileHandler struct {
	ProfileService
}

func NewHandler(service ProfileService) ProfileHandler {
	return ProfileHandler{ProfileService: service}
}

func (p ProfileHandler) HandlerName() string {
	return "Profile"
}

func (p ProfileHandler) HandleCreateProfile(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		var payload dtos.ProfileSignUpRequest
		if err := utils.ParseJSON(r, payload); err != nil {
			p.logger.MustDebug(fmt.Sprintf("failed to parse request body, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			p.logger.MustDebug(fmt.Sprintf("profile create payload failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		profile, err := p.CreatePrfl(r.Context(), payload)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("database err %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"profile": profile,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p ProfileHandler) HandleGetProfiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("failed create profile filter, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		prfls, err := p.GetPrfl(r.Context(), fltr)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"profiles": prfls,
			"success":  true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p ProfileHandler) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("failed to create profile filter, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		prfl, err := p.GetPrfl(r.Context(), fltr)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("database errr, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"profile": prfl,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.MustDebug(fmt.Sprintf("internal server error, %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

	}
}

func (p ProfileHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		var payload dtos.ProfileRequest
		if err := utils.ParseJSON(r, payload); err != nil {
			p.logger.MustDebug(fmt.Sprintf("failed to parse profile dto from request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			p.logger.MustDebug(fmt.Sprintf("profile dto failed validation, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		updatePrfl, err := p.ModifyProfile(r.Context(), payload)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"profile": updatePrfl,
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p ProfileHandler) HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.MustDebug(timeoutErr.Error())
		utils.WriteErr(w, http.StatusRequestTimeout, &timeoutErr)
	default:
		delReq, err := dtos.BuildDeleteRequest(r)
		if err != nil {
			p.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err = delReq.Validate(); err != nil {
			p.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", err))
			utils.WriteErr(w, http.StatusBadRequest, fmt.Errorf("invalid delete request, %v", err))
			return
		}

		req, ok := delReq.(dtos.DeleteRequest)
		if !ok {
			invalidData := cmerr.ErrUnexpectedData{Wanted: dtos.DeleteRequest{}, Got: delReq}
			p.logger.MustDebug(fmt.Sprintf("invalid delete request, %v", invalidData))
			utils.WriteErr(w, http.StatusBadRequest, invalidData)
			return
		}

		if err = p.DeletePrfl(r.Context(), req); err != nil {
			p.logger.MustDebug(fmt.Sprintf("database err, %v", err))
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"success": true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.MustDebug(err.Error())
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
