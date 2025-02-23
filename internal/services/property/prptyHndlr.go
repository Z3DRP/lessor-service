package property

import (
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/sirupsen/logrus"
)

var (
	maxSize = int64(1024000)
)

type PropertyHandler struct {
	PropertyService
}

func NewHandler(service PropertyService) PropertyHandler {
	return PropertyHandler{PropertyService: service}
}

func (p PropertyHandler) HandlerName() string {
	return "Property"
}

func (p PropertyHandler) HandleCreateProperty(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeoutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		var payload dtos.PropertyRequest
		w.Header().Set("Content-Type", "application/json")
		err := r.ParseMultipartForm(maxSize)

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "file to large",
				"err": api.ErrMaxSize{Err: err},
			})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		// need to pass file and header to createProperty
		file, header, err := r.FormFile("image")

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "could not parse file from request",
				"err": api.ErrFileRead{Err: err},
			})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		defer file.Close()

		if err := utils.ParseJSON(r, &payload); err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to parse json", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		if err := payload.Validate(); err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "property validation failed",
				"err": err,
			})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		property, err := p.CreateProperty(r.Context(), payload)

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "failed to create property",
				"err": err,
			})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"property": property,
			"success":  true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to encode json resposne", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p PropertyHandler) HandleGetProperty(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeOutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeOutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeOutErr)
	default:
		fltr, err := filters.GenFilter(r)
		if err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to make filter", "err": err})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		prprty, err := p.GetProperty(r.Context(), fltr)

		if err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "database error", "err": err})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		res := ztype.JsonResponse{
			"property": prprty,
			"success":  true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "failed to encode json response",
				"err": err,
			})
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p PropertyHandler) HandleGetProperties(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": timeoutErr,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "failed to create request filter",
				"err": err,
			})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		properties, err := p.GetProperties(r.Context(), fltr)

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "database err",
				"err": err,
			})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"properties": properties,
			"success":    true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p PropertyHandler) HandleUpdateProperty(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		err := utils.ErrRequestTimeout{Request: r}
		p.logger.LogFields(logrus.Fields{
			"msg": "request timeout",
			"err": err,
		})
		utils.WriteErr(w, http.StatusRequestTimeout, err)
	default:
		var payload dtos.PropertyModificationRequest

		if err := utils.ParseJSON(r, &payload); err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to parse required dto fields", "err": err})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		if err := payload.Validate(); err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "dto failed validation",
				"err": err,
			})

			utils.WriteErr(w, http.StatusBadRequest, err)
		}

		property, err := p.ModifyProperty(r.Context(), payload)

		if err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "database err", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"property": property,
			"success":  true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to encode json response", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
