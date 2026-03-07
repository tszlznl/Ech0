package handler

import (
	"github.com/google/wire"
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	inboxHandler "github.com/lin-snow/ech0/internal/handler/inbox"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	todoHandler "github.com/lin-snow/ech0/internal/handler/todo"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
)

var (
	WebSet       = wire.NewSet(webHandler.ProviderSet)
	UserSet      = wire.NewSet(userHandler.ProviderSet)
	EchoSet      = wire.NewSet(echoHandler.ProviderSet)
	CommonSet    = wire.NewSet(commonHandler.ProviderSet)
	SettingSet   = wire.NewSet(settingHandler.ProviderSet)
	TodoSet      = wire.NewSet(todoHandler.ProviderSet)
	ConnectSet   = wire.NewSet(connectHandler.ProviderSet)
	BackupSet    = wire.NewSet(backupHandler.ProviderSet)
	DashboardSet = wire.NewSet(dashboardHandler.ProviderSet)
	AgentSet     = wire.NewSet(agentHandler.ProviderSet)
	InboxSet     = wire.NewSet(inboxHandler.ProviderSet)
)

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
)
