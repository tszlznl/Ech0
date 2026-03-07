package repository

import (
	"github.com/google/wire"
	commonRepository "github.com/lin-snow/ech0/internal/repository/common"
	connectRepository "github.com/lin-snow/ech0/internal/repository/connect"
	echoRepository "github.com/lin-snow/ech0/internal/repository/echo"
	fileRepository "github.com/lin-snow/ech0/internal/repository/file"
	inboxRepository "github.com/lin-snow/ech0/internal/repository/inbox"
	keyvalueRepository "github.com/lin-snow/ech0/internal/repository/keyvalue"
	queueRepository "github.com/lin-snow/ech0/internal/repository/queue"
	settingRepository "github.com/lin-snow/ech0/internal/repository/setting"
	todoRepository "github.com/lin-snow/ech0/internal/repository/todo"
	userRepository "github.com/lin-snow/ech0/internal/repository/user"
	webhookRepository "github.com/lin-snow/ech0/internal/repository/webhook"
)

var UserSet = wire.NewSet(userRepository.ProviderSet)
var EchoSet = wire.NewSet(echoRepository.ProviderSet)
var CommonSet = wire.NewSet(commonRepository.ProviderSet)
var FileSet = wire.NewSet(fileRepository.ProviderSet)
var KeyValueSet = wire.NewSet(keyvalueRepository.ProviderSet)
var SettingSet = wire.NewSet(settingRepository.ProviderSet)
var TodoSet = wire.NewSet(todoRepository.ProviderSet)
var ConnectSet = wire.NewSet(connectRepository.ProviderSet)
var WebhookSet = wire.NewSet(webhookRepository.ProviderSet)
var InboxSet = wire.NewSet(inboxRepository.ProviderSet)
var QueueSet = wire.NewSet(queueRepository.ProviderSet)

var ProviderSet = wire.NewSet(
	UserSet,
	EchoSet,
	CommonSet,
	FileSet,
	KeyValueSet,
	SettingSet,
	TodoSet,
	ConnectSet,
	WebhookSet,
	InboxSet,
	QueueSet,
)
