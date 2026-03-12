package handler

import (
	agentHandler "github.com/lin-snow/ech0/internal/handler/agent"
	backupHandler "github.com/lin-snow/ech0/internal/handler/backup"
	commonHandler "github.com/lin-snow/ech0/internal/handler/common"
	connectHandler "github.com/lin-snow/ech0/internal/handler/connect"
	dashboardHandler "github.com/lin-snow/ech0/internal/handler/dashboard"
	echoHandler "github.com/lin-snow/ech0/internal/handler/echo"
	fileHandler "github.com/lin-snow/ech0/internal/handler/file"
	inboxHandler "github.com/lin-snow/ech0/internal/handler/inbox"
	initHandler "github.com/lin-snow/ech0/internal/handler/init"
	migrationHandler "github.com/lin-snow/ech0/internal/handler/migration"
	commentHandler "github.com/lin-snow/ech0/internal/handler/comment"
	settingHandler "github.com/lin-snow/ech0/internal/handler/setting"
	userHandler "github.com/lin-snow/ech0/internal/handler/user"
	webHandler "github.com/lin-snow/ech0/internal/handler/web"
)

// Bundle 聚合各业务 Handler。
type Bundle struct {
	WebHandler       *webHandler.WebHandler
	UserHandler      *userHandler.UserHandler
	EchoHandler      *echoHandler.EchoHandler
	FileHandler      *fileHandler.FileHandler
	CommentHandler   *commentHandler.CommentHandler
	InitHandler      *initHandler.InitHandler
	CommonHandler    *commonHandler.CommonHandler
	SettingHandler   *settingHandler.SettingHandler
	InboxHandler     *inboxHandler.InboxHandler
	ConnectHandler   *connectHandler.ConnectHandler
	BackupHandler    *backupHandler.BackupHandler
	MigrationHandler *migrationHandler.MigrationHandler
	DashboardHandler *dashboardHandler.DashboardHandler
	AgentHandler     *agentHandler.AgentHandler
}

// NewBundle 创建 Handler 聚合实例。
func NewBundle(
	webHandler *webHandler.WebHandler,
	userHandler *userHandler.UserHandler,
	echoHandler *echoHandler.EchoHandler,
	fileHandler *fileHandler.FileHandler,
	commentHandler *commentHandler.CommentHandler,
	initHandler *initHandler.InitHandler,
	commonHandler *commonHandler.CommonHandler,
	settingHandler *settingHandler.SettingHandler,
	inboxHandler *inboxHandler.InboxHandler,
	connectHandler *connectHandler.ConnectHandler,
	backupHandler *backupHandler.BackupHandler,
	migrationHandler *migrationHandler.MigrationHandler,
	dashboardHandler *dashboardHandler.DashboardHandler,
	agentHandler *agentHandler.AgentHandler,
) *Bundle {
	return &Bundle{
		WebHandler:       webHandler,
		UserHandler:      userHandler,
		EchoHandler:      echoHandler,
		FileHandler:      fileHandler,
		CommentHandler:   commentHandler,
		InitHandler:      initHandler,
		CommonHandler:    commonHandler,
		SettingHandler:   settingHandler,
		InboxHandler:     inboxHandler,
		ConnectHandler:   connectHandler,
		BackupHandler:    backupHandler,
		MigrationHandler: migrationHandler,
		DashboardHandler: dashboardHandler,
		AgentHandler:     agentHandler,
	}
}
