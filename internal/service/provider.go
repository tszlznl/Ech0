package service

import (
	"github.com/google/wire"
	agentService "github.com/lin-snow/ech0/internal/service/agent"
	backupService "github.com/lin-snow/ech0/internal/service/backup"
	commonService "github.com/lin-snow/ech0/internal/service/common"
	connectService "github.com/lin-snow/ech0/internal/service/connect"
	dashboardService "github.com/lin-snow/ech0/internal/service/dashboard"
	echoService "github.com/lin-snow/ech0/internal/service/echo"
	inboxService "github.com/lin-snow/ech0/internal/service/inbox"
	settingService "github.com/lin-snow/ech0/internal/service/setting"
	todoService "github.com/lin-snow/ech0/internal/service/todo"
	userService "github.com/lin-snow/ech0/internal/service/user"
)

var UserSet = wire.NewSet(userService.ProviderSet)
var EchoSet = wire.NewSet(echoService.ProviderSet)
var CommonSet = wire.NewSet(commonService.ProviderSet)
var SettingSet = wire.NewSet(settingService.ProviderSet)
var TodoSet = wire.NewSet(todoService.ProviderSet)
var ConnectSet = wire.NewSet(connectService.ProviderSet)
var BackupSet = wire.NewSet(backupService.ProviderSet)
var DashboardSet = wire.NewSet(dashboardService.ProviderSet)
var AgentSet = wire.NewSet(agentService.ProviderSet)
var InboxSet = wire.NewSet(inboxService.ProviderSet)

var ProviderSet = wire.NewSet(
	UserSet,
	EchoSet,
	CommonSet,
	SettingSet,
	TodoSet,
	ConnectSet,
	BackupSet,
	DashboardSet,
	AgentSet,
	InboxSet,
)
