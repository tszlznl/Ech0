package service

import (
	"github.com/lin-snow/ech0/internal/fediverse"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	repository "github.com/lin-snow/ech0/internal/repository/fediverse"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	"github.com/lin-snow/ech0/internal/transaction"
)

type FediverseService struct {
	core                *fediverse.FediverseCore
	txManager           transaction.TransactionManager
	fediverseRepository repository.FediverseRepositoryInterface
	userRepository      userRepository.UserRepositoryInterface
	echoRepository      echoRepository.EchoRepositoryInterface
}

func NewFediverseService(
	core *fediverse.FediverseCore,
	txManager transaction.TransactionManager,
	fediverseRepository repository.FediverseRepositoryInterface,
	userRepository userRepository.UserRepositoryInterface,
	echoRepository echoRepository.EchoRepositoryInterface,
) *FediverseService {
	return &FediverseService{
		core:                core,
		txManager:           txManager,
		fediverseRepository: fediverseRepository,
		userRepository:      userRepository,
		echoRepository:      echoRepository,
	}
}
