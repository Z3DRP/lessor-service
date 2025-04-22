package notification

import (
	"context"
	"database/sql"

	"github.com/Z3DRP/lessor-service/internal/cmerr"
	"github.com/Z3DRP/lessor-service/internal/crane"
	"github.com/Z3DRP/lessor-service/internal/dac"
	"github.com/Z3DRP/lessor-service/internal/dtos"
	"github.com/Z3DRP/lessor-service/internal/filters"
	"github.com/Z3DRP/lessor-service/internal/model"
	"github.com/Z3DRP/lessor-service/pkg/utils"
)

type NotificationService struct {
	repo   dac.NotificationRepo
	logger *crane.Zlogrus
}

func (n NotificationService) ServiceName() string {
	return "Notification"
}

func NewNotificationService(repo dac.NotificationRepo, logr *crane.Zlogrus) NotificationService {
	return NotificationService{
		repo:   repo,
		logger: logr,
	}
}

func (n NotificationService) GetNotification(ctx context.Context, fltr filters.Filterer) (*dtos.NotificationDto, error) {
	flter, ok := fltr.(filters.Filter)
	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("id filter")
	}

	noti, err := n.repo.Fetch(ctx, flter)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	notification, ok := noti.(model.Notification)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Notification{}, Got: noti}
	}

	return dtos.NewNotificationDto(notification), nil
}

// TODO: need to parse lessorIds out of notification get requests

func (n NotificationService) GetNotifications(ctx context.Context, fltr filters.Filterer) ([]dtos.NotificationDto, error) {
	var response []dtos.NotificationDto
	filter, ok := fltr.(filters.Filter)

	if !ok {
		return nil, filters.NewFailedToMakeFilterErr("id filter")
	}

	notifs, err := n.repo.FetchAll(ctx, filter)
	if err != nil {
		if err == sql.ErrNoRows {
			return make([]dtos.NotificationDto, 0), nil
		}
		return nil, err
	}

	response = dtos.NewNotificationDtoList(notifs)
	return response, nil
}

func (n NotificationService) CreateNotification(ctx context.Context, data *dtos.NotificationDto) (*dtos.NotificationDto, error) {
	notif := newNotification(data)
	nwNotif, err := n.repo.Insert(ctx, notif)
	if err != nil {
		return nil, err
	}

	noti, ok := nwNotif.(model.Notification)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Notification{}, Got: noti}
	}

	return dtos.NewNotificationDto(noti), nil
}

func (n NotificationService) UpdateViewed(ctx context.Context, nid int) (*dtos.NotificationDto, error) {
	nwNtf, err := n.repo.UpdateViewed(ctx, nid)

	if err != nil {
		return nil, err
	}

	noti, ok := nwNtf.(model.Notification)
	if !ok {
		return nil, cmerr.ErrUnexpectedData{Wanted: model.Notification{}, Got: nwNtf}
	}

	return dtos.NewNotificationDto(noti), nil
}

func newNotification(n *dtos.NotificationDto) model.Notification {
	return model.Notification{
		Title:      n.Title,
		Message:    n.Message,
		LessorId:   utils.ParseUuid(n.LessorId),
		TaskId:     utils.ParseUuid(n.TaskId),
		UserId:     utils.ParseUuid(n.UserId),
		PropertyId: utils.ParseUuid(n.PropertyId),
		Category:   model.NotificationType(n.Category),
	}
}
