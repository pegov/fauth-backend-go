package service

import (
	"github.com/pegov/fauth-backend-go/internal/repo"
)

type AdminService interface {
	Ban(id int32) error
	Unban(id int32) error
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

func (s *adminService) Ban(id int32) error {
	user, err := s.userRepo.Get(id)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.Ban(id); err != nil {
		return err
	}

	return nil
}

func (s *adminService) Unban(id int32) error {
	user, err := s.userRepo.Get(id)
	if err != nil {
		return err
	}

	if user == nil {
		return ErrUserNotFound
	}

	if err := s.userRepo.Unban(id); err != nil {
		return err
	}

	return nil
}
