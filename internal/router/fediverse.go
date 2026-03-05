package router

import "github.com/lin-snow/ech0/internal/handler"

func setupFediverseRoutes(appRouterGroup *AppRouterGroup, h *handler.Bundle) {
	//==============
	// Fediverse 必须要的路由
	//==============

	// ActivityPub discovery
	appRouterGroup.ResourceGroup.GET("/.well-known/webfinger", h.FediverseHandler.Webfinger)

	// Actor (用户资料)
	appRouterGroup.ResourceGroup.GET("/users/:username", h.FediverseHandler.GetActor)

	// Inbox (接收 ActivityPub 消息)
	appRouterGroup.ResourceGroup.POST("/users/:username/inbox", h.FediverseHandler.PostInbox)

	// Outbox (发布消息)
	appRouterGroup.ResourceGroup.GET("/users/:username/outbox", h.FediverseHandler.GetOutbox)

	// Followers list
	appRouterGroup.ResourceGroup.GET("/users/:username/followers", h.FediverseHandler.GetFollowers)

	// Following list
	appRouterGroup.ResourceGroup.GET("/users/:username/following", h.FediverseHandler.GetFollowing)

	// Objects (内容对象访问)
	appRouterGroup.ResourceGroup.GET("/objects/:id", h.FediverseHandler.GetObject)

	//==============
	// 前端自用的相关路由
	//==============

	// // Search Actor By Actor ID
	// appRouterGroup.AuthRouterGroup.GET("/search/actor", h.FediverseHandler.SearchActorByActorID)

	// // Get Follow Status (获取关注状态)
	// appRouterGroup.AuthRouterGroup.GET("/follow/status", h.FediverseHandler.GetFollowStatus)

	// // Follow (发起关注请求)
	// appRouterGroup.AuthRouterGroup.POST("/follow", h.FediverseHandler.PostFollow)

	// // Unfollow (取消关注请求)
	// appRouterGroup.AuthRouterGroup.POST("/unfollow", h.FediverseHandler.PostUnfollow)

	// // Get Timeline (获取关注的Actor的推文)
	// appRouterGroup.AuthRouterGroup.GET("/timeline", h.FediverseHandler.GetTimeline)

	// List Following (列出当前关注的所有Actor)
	// appRouterGroup.AuthRouterGroup.GET("/following", h.FediverseHandler.GetFollowingList)

	// Post Like (点赞请求)
	// appRouterGroup.AuthRouterGroup.POST("/like", h.FediverseHandler.PostLike)

	// Post Undo Like (取消点赞请求)
	// appRouterGroup.AuthRouterGroup.POST("/undo-like", h.FediverseHandler.PostUndoLike)
}
