package service

import (
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	initModel "github.com/lin-snow/ech0/internal/model/init"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

type Service interface {
	GetStatus() (initModel.Status, error)
	InitOwner(registerDto *authModel.RegisterDto) error
}

type Repository interface {
	IsInitialized() (bool, error)
	GetOwner() (userModel.User, error)
}

type UserService = userService.Service
