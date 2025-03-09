package rentalproperty

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

type RentalPropertyHandler struct {
	RentalPropertyService
}

func NewHandler(service RentalPropertyService) RentalPropertyHandler {
	return RentalPropertyHandler{RentalPropertyService: service}
}

func (p RentalPropertyHandler) HandlerName() string {
	return "Rental Property"
}

func (p RentalPropertyHandler) HandleCreateRentalProperty(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		p.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeoutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		payload := dtos.RentalPropertyDto{}

		if err := utils.ParseJSON(r, payload); err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to parse json", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
		property, err := p.CreateRentalProperty(r.Context(), &payload)

		if err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "failed to create property",
				"err": err,
			})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"rentalProperty": property,
			"success":        true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			p.logger.LogFields(logrus.Fields{"msg": "failed to encode json resposne", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}

func (p RentalPropertyHandler) HandleGetRentalProperty(w http.ResponseWriter, r *http.Request) {
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

		prprty, err := p.GetRentalProperty(r.Context(), fltr)

		if err != nil {
			var noResults dac.ErrNoResults
			if errors.As(err, &noResults) {
				res := ztype.JsonResponse{
					"rentalProperty": nil,
					"success":        true,
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

func (p RentalPropertyHandler) HandleGetRentalProperties(w http.ResponseWriter, r *http.Request) {
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

		// need to update this because if no rows then GetProperties call is returning nil, nil
		properties, err := p.GetRentalProperties(r.Context(), fltr)

		if err != nil {
			var noResults *dac.ErrNoResults
			if errors.As(err, noResults) {
				res := ztype.JsonResponse{
					"properties": make([]dtos.RentalPropertyDto, 0),
					"success":    true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
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

func (p RentalPropertyHandler) HandleUpdateRentalProperty(w http.ResponseWriter, r *http.Request) {
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
		payload := &dtos.RentalPropertyDto{}

		if err := utils.ParseJSON(r, payload); err != nil {
			p.logger.LogFields(logrus.Fields{
				"msg": "fialed to parse json",
				"err": err,
			})
			log.Printf("failed to parse json")
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		property, err := p.ModifyRentalProperty(r.Context(), payload)

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

func (p *RentalPropertyHandler) HandleDeleteRentalProperty(w http.ResponseWriter, r *http.Request) {
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
		err := p.DeleteRentalProperty(r.Context(), filter)

		if err != nil {
			log.Printf("error deleting rental property: %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}

		res := ztype.JsonResponse{
			"propertyId": filter.Identifier,
			"success":    err == nil,
		}

		if err := utils.WriteJSON(w, http.StatusOK, res); err != nil {
			log.Printf("failed to write response %v", err)
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
