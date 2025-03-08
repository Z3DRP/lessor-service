package property

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Z3DRP/lessor-service/internal/adapters"
	"github.com/Z3DRP/lessor-service/internal/api"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/sirupsen/logrus"
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
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeoutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		var (
			fileUpload *ztype.FileUploadDto
			payload    *dtos.PropertyRequest
		)

		contentType := r.Header.Get("Content-Type")

		if strings.HasPrefix(contentType, "multipart/form-data") {
			file, header, err := utils.ParseFile(r)

			if err != nil {
				p.logger.LogFields(logrus.Fields{
					"msg": "error occurred while parsing file from request",
					"err": err,
				})
				utils.WriteErr(w, http.StatusBadRequest, err)
				return
			}

			payload, err = adapters.ParsePropertyForm(r)
			if err != nil {
				log.Printf("failed to parse property form %v", err)
				utils.WriteErr(w, http.StatusBadRequest, err)
				return
			}

			if err = payload.Validate(); err != nil {
				p.logger.LogFields(logrus.Fields{
					"msg": "property validation failed",
					"err": err,
				})
				utils.WriteErr(w, http.StatusBadRequest, err)
				return
			}

			if file != nil && header != nil {
				fileUpload = &ztype.FileUploadDto{File: file, FileKey: payload.Image, Header: header}
				defer file.Close()

				if err = fileUpload.Validate(); err != nil {
					p.logger.LogFields(logrus.Fields{
						"msg": "an error occurred while create file upload dto",
						"err": err,
					})
					utils.WriteErr(w, http.StatusBadRequest, err)
					return
				}
			}
		} else {
			payload = &dtos.PropertyRequest{}
			if err := utils.ParseJSON(r, payload); err != nil {
				p.logger.LogFields(logrus.Fields{"msg": "failed to parse json", "err": err})
				utils.WriteErr(w, http.StatusInternalServerError, err)
				return
			}
		}

		property, err := p.CreateProperty(r.Context(), payload, fileUpload)

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
		log.Printf("fetching single user with id")
		fltr, err := filters.GenFilter(r)
		if err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to make filter", "err": err})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		prprty, err := p.GetProperty(r.Context(), fltr)

		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"property": nil,
					"success":  true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			p.logger.LogFields(logrus.Fields{"msg": "database error", "err": err})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		props, err := json.Marshal(prprty)
		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		// res := ztype.JsonResponse{
		// 	"property": props,
		// 	"success":  true,
		// }

		if err = utils.WriteJSON(w, http.StatusOK, struct {
			Property interface{} `json:"property"`
			Success  bool        `json:"success"`
		}{
			Property: props,
			Success:  true,
		}); err != nil {
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
		log.Printf("fetching all propeties")
		fltr, err := filters.GenFilter(r)

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "failed to create request filter",
				"err": err,
			})
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		// need to update this because if no rows then GetProperties call is returning nil, nil
		properties, err := p.GetProperties(r.Context(), fltr)

		if err != nil {
			var noResults *dac.ErrNoResults
			if errors.As(err, noResults) {
				res := ztype.JsonResponse{
					"properties": make([]dtos.PropertyResponse, 0),
					"success":    true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}

			if err == api.ErrrNoImagesFound {
				log.Printf("no images err as blck")
				res := ztype.JsonResponse{
					"properties": properties,
					"success":    true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
				}
				return
			}

			p.logger.LogFields(logrus.Fields{
				"msg": "service err",
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
		log.Printf("update ep hit")
		var (
			fileUpload *ztype.FileUploadDto
			payload    *dtos.PropertyModificationRequest
		)

		contentType := r.Header.Get("Content-Type")

		if strings.HasPrefix(contentType, "multipart/form-data") {
			log.Printf("is multipart form")
			file, header, err := utils.ParseFile(r)

			if err != nil {
				p.logger.LogFields(logrus.Fields{
					"msg": "error occurred while parsing file update",
					"err": err,
				})
				log.Printf("failed to pasre file update %v", err)

				utils.WriteErr(w, http.StatusBadRequest, err)
				return
			}

			payload, err = adapters.ParsePropertyUpdateForm(r)

			if err != nil {
				log.Printf("fialed to parse property fomr %v", err)
				utils.WriteErr(w, http.StatusBadRequest, err)
				return
			}

			if err = payload.Validate(); err != nil {
				log.Printf("failed to validate update request")
				utils.WriteErr(w, http.StatusBadRequest, err)
				return
			}

			if file != nil && header != nil {
				fileUpload = &ztype.FileUploadDto{File: file, FileKey: payload.Image, Header: header}
				defer file.Close()

				if err = fileUpload.Validate(); err != nil {
					log.Printf("failed to validate file")
					utils.WriteErr(w, http.StatusBadRequest, err)
					return
				}
			}
		} else {
			log.Printf("handling as json")
			payload = &dtos.PropertyModificationRequest{}
			if err := utils.ParseJSON(r, payload); err != nil {
				p.logger.LogFields(logrus.Fields{
					"msg": "fialed to parse json",
					"err": err,
				})
				log.Printf("failed to parse json")
				utils.WriteErr(w, http.StatusInternalServerError, err)
				return
			}
		}

		property, err := p.ModifyProperty(r.Context(), payload, fileUpload)

		if err != nil {
			log.Printf("failed to update property")
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"property": property,
			"success":  true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed to create json response")
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p *PropertyHandler) HandleDeleteProperty(w http.ResponseWriter, r *http.Request) {
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
		filter := filters.NewIdFilter(r)
		err := p.DeleteProperty(r.Context(), filter)

		if err != nil {
			log.Printf("error deleting property: %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}

		res := ztype.JsonResponse{
			"propertyId": filter.Identifier,
			"success":    err == nil,
		}

		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("fialed to write response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
