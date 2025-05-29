package service

import (
	"context"

	"github.com/taekwondodev/push-notification-service/internal/models"
	"github.com/taekwondodev/push-notification-service/internal/repository"
)

type NotificationService struct {
    repo   repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
    return &NotificationService{
        repo:   repo,
    }
}

func (s *NotificationService) CreateNotification(ctx context.Context, notification *models.Notification) error {
    if err := s.repo.Save(ctx, notification); err != nil {
        return err
    }
    
    return nil
}

func (s *NotificationService) GetNotificationsByReceiver(ctx context.Context, receiver string) ([]models.Notification, error) {
    notifications, err := s.repo.FindByReceiver(ctx, receiver)
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