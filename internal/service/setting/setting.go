package service

import (
	publisher "github.com/lin-snow/ech0/internal/event/publisher"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/internal/transaction"
)

type SettingService struct {
	transactor         transaction.Transactor
	commonService      CommonService
	fileService        FileService
	storageManager     *storage.Manager
	keyvalueRepository KeyValueRepository
	settingRepository  SettingRepository
	webhookRepository  WebhookRepository
	publisher          *publisher.Publisher
}

func NewSettingService(
	tx transaction.Transactor,
	commonService CommonService,
	fileService FileService,
	storageManager *storage.Manager,
	keyvalueRepository KeyValueRepository,
	settingRepository SettingRepository,
	webhookRepository WebhookRepository,
	publisher *publisher.Publisher,
) *SettingService {
	return &SettingService{
		transactor:         tx,
		commonService:      commonService,
		fileService:        fileService,
		storageManager:     storageManager,
		keyvalueRepository: keyvalueRepository,
		webhookRepository:  webhookRepository,
		settingRepository:  settingRepository,
		publisher:          publisher,
	}
}
