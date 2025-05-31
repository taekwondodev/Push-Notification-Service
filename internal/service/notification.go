package service

import (
	"context"

	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationServiceInterface interface {
	CreateNotification(ctx context.Context, notification *models.Notification) error
	GetNotificationsByReceiver(ctx context.Context, receiver string, unreadOnly bool) ([]models.Notification, error)
	MarkAsRead(ctx context.Context, id string) error
	GetNotificationByID(ctx context.Context, id string) (*models.Notification, error)
}

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
	if notification.ID.IsZero() {
		notification.ID = primitive.NewObjectID()
	}

	if err := s.repo.Save(ctx, notification); err != nil {
		return err
	}

	return nil
}

func (s *NotificationService) GetNotificationsByReceiver(ctx context.Context, receiver string, unread bool) ([]models.Notification, error) {
	notifications, err := s.repo.FindByReceiver(ctx, receiver, unread)
	if err != nil {
		return nil, err
	}

	return notifications, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id string) error {
	if err := s.repo.MarkAsRead(ctx, id); err != nil {
		return err
	}

	return nil
}
