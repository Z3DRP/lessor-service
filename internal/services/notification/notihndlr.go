package notification

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/ztype"
	"github.com/Z3DRP/lessor-service/pkg/utils"
	"github.com/sirupsen/logrus"
)

type NotificationHandler struct {
	NotificationService
}

func NewHandler(service NotificationService) NotificationHandler {
	return NotificationHandler{
		NotificationService: service,
	}
}

func (n NotificationHandler) HandlerName() string {
	return "Notification"
}

func (n NotificationHandler) HandleCreateNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		n.logger.LogFields(logrus.Fields{"msg": "request timeout", "err": timeoutErr})
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		payload := &dtos.NotificationDto{}

		if err := utils.ParseJSON(r, payload); err != nil {
			n.logger.LogFields(logrus.Fields{"msg": "failed to parse request body", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		notif, err := n.CreateNotification(r.Context(), payload)
		if err != nil {
			n.logger.LogFields(logrus.Fields{"msg": "failed to create notification", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"notification": notif,
			"success":      true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			n.logger.LogFields(logrus.Fields{"msg": "failed to write json response", "err": err})
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (n NotificationHandler) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		n.logger.MustDebug("request timeout")
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		fltr, err := filters.GenFilter(r)

		if err != nil {
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		nofi, err := n.GetNotifications(r.Context(), fltr)

		if err != nil {
			if err == sql.ErrNoRows {
				res := ztype.JsonResponse{
					"notifications": make([]dtos.NotificationDto, 0),
					"success":       true,
				}

				if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
					utils.WriteErr(w, http.StatusInternalServerError, err)
					return
				}
			}
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}

		res := ztype.JsonResponse{
			"notifications": nofi,
			"success":       true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (n NotificationHandler) HandleUpdateViewed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	select {
	case <-r.Context().Done():
		timeoutErr := utils.ErrRequestTimeout{Request: r}
		n.logger.MustDebug("request timeout")
		utils.WriteErr(w, http.StatusRequestTimeout, timeoutErr)
	default:
		id := r.PathValue("id")

		if id == "" {
			utils.WriteErr(w, http.StatusBadRequest, errors.New("missing identifier in request"))
			return
		}

		ntfId, err := utils.ParseIntOrZero(id)

		if err != nil {
			utils.WriteErr(w, http.StatusBadRequest, err)
			return
		}

		noti, err := n.UpdateViewed(r.Context(), ntfId)

		if err != nil {
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}

		res := ztype.JsonResponse{
			"notification": noti,
			"success":      true,
		}

		if err = utils.WriteJSON(w, http.StatusOK, res); err != nil {
			n.logger.Zlog(map[string]interface{}{
				"msg": "error occurred while writing json response",
				"err": err,
			})
			utils.WriteErr(w, http.StatusInternalServerError, err)
		}
	}
}
