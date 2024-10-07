package service

import (
	"context"
	"time"

	"github.com/pegov/fauth-backend-go/internal/model"
	"github.com/pegov/fauth-backend-go/internal/repo"
)

type AdminService interface {
	GetMassLogout(ctx context.Context) (model.MassLogoutStatus, error)
	ActivateMassLogout(ctx context.Context) error
	DeactivateMassLogout(ctx context.Context) error
	Ban(ctx context.Context, id int32) error
	Unban(ctx context.Context, id int32) error
	Kick(ctx context.Context, id int32) error
	Unkick(ctx context.Context, id int32) error
}

type adminService struct {
	userRepo repo.UserRepo
}

func NewAdminService(
	userRepo repo.UserRepo,
) AdminService {
	return &adminService{
		userRepo: userRepo,
	}
}

func (s *adminService) GetMassLogout(ctx context.Context) (model.MassLogoutStatus, error) {
	date, err := s.userRepo.GetMassLogout(ctx)
	if err != nil {
		return model.MassLogoutStatus{}, err
	}

	if date == nil {
		return model.MassLogoutStatus{}, nil
	}

	return model.MassLogoutStatus{
		Date:   date,
		Active: true,
	}, nil
}

func (s *adminService) ActivateMassLogout(ctx context.Context) error {
	return s.userRepo.ActivateMassLogout(ctx, 60*60*24*31*time.Second)
}

func (s *adminService) DeactivateMassLogout(ctx context.Context) error {
	return s.userRepo.DeactivateMassLogout(ctx)
}

func (s *adminService) actionOnID(
	ctx context.Context,
	id int32,
	action func(context.Context, int32) error,
) error {
	user, err := s.userRepo.Get(ctx, id)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	if err := action(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *adminService) Ban(ctx context.Context, id int32) error {
	return s.actionOnID(ctx, id, s.userRepo.Ban)
}

func (s *adminService) Unban(ctx context.Context, id int32) error {
	return s.actionOnID(ctx, id, s.userRepo.Unban)
}

func (s *adminService) Kick(ctx context.Context, id int32) error {
	return s.actionOnID(ctx, id, s.userRepo.Kick)
}

func (s *adminService) Unkick(ctx context.Context, id int32) error {
	return s.actionOnID(ctx, id, s.userRepo.Unkick)
}
