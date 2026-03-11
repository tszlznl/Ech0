package service

import (
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	initModel "github.com/lin-snow/ech0/internal/model/init"
)

type InitService struct {
	repository  Repository
	userService UserService
}

func NewInitService(repository Repository, userService UserService) *InitService {
	return &InitService{
		repository:  repository,
		userService: userService,
	}
}

func (s *InitService) GetStatus() (initModel.Status, error) {
	initialized, err := s.repository.IsInitialized()
	if err != nil {
		return initModel.Status{}, err
	}

	_, ownerErr := s.repository.GetOwner()
	ownerExists := ownerErr == nil

	return initModel.Status{
		Initialized: initialized,
		OwnerExists: ownerExists,
	}, nil
}

func (s *InitService) InitOwner(registerDto *authModel.RegisterDto) error {
	initialized, err := s.repository.IsInitialized()
	if err != nil {
		return err
	}
	if initialized {
		return commonModel.NewBizError(commonModel.ErrCodeInitAlreadyDone, commonModel.SYSTEM_ALREADY_INITED)
	}
	return s.userService.InitOwner(registerDto)
}
