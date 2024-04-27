package service

import (
	"context"

	"github.com/pegov/fauth-backend-go/internal/repo"
)

type AdminService interface {
	Ban(id int32) error
	Unban(id int32) error
	Kick(id int32) error
	Unkick(id int32) error
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

func (s *adminService) actionOnID(id int32, action func(int32) error) error {
	user, err := s.userRepo.Get(id)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	if err := action(id); err != nil {
		return err
	}

	return nil
}

func (s *adminService) actionWithContextOnID(ctx context.Context, id int32, action func(context.Context, int32) error) error {
	user, err := s.userRepo.Get(id)
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

func (s *adminService) Ban(id int32) error {
	return s.actionOnID(id, s.userRepo.Ban)
}

func (s *adminService) Unban(id int32) error {
	return s.actionOnID(id, s.userRepo.Unban)
}

func (s *adminService) Kick(id int32) error {
	ctx := context.TODO()
	return s.actionWithContextOnID(ctx, id, s.userRepo.Kick)
}

func (s *adminService) Unkick(id int32) error {
	ctx := context.TODO()
	return s.actionWithContextOnID(ctx, id, s.userRepo.Unkick)
}
