package handler

import (
	"github.com/google/wire"
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	fediverseHandler "github.com/lin-snow/ech0/internal/handler/fediverse"
	inboxHandler "github.com/lin-snow/ech0/internal/handler/inbox"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	todoHandler "github.com/lin-snow/ech0/internal/handler/todo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
)

var WebSet = wire.NewSet(webHandler.ProviderSet)
var UserSet = wire.NewSet(userHandler.ProviderSet)
var EchoSet = wire.NewSet(echoHandler.ProviderSet)
var CommonSet = wire.NewSet(commonHandler.ProviderSet)
var SettingSet = wire.NewSet(settingHandler.ProviderSet)
var TodoSet = wire.NewSet(todoHandler.ProviderSet)
var ConnectSet = wire.NewSet(connectHandler.ProviderSet)
var BackupSet = wire.NewSet(backupHandler.ProviderSet)
var DashboardSet = wire.NewSet(dashboardHandler.ProviderSet)
var AgentSet = wire.NewSet(agentHandler.ProviderSet)
var InboxSet = wire.NewSet(inboxHandler.ProviderSet)
var FediverseSet = wire.NewSet(fediverseHandler.ProviderSet)

var ProviderSet = wire.NewSet(
	WebSet,
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
	FediverseSet,
)
