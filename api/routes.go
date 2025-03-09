package api

import (
	"chatweb/config"
	"chatweb/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, cfg *config.Config, handlers *Handlers) {
	// 公开路由
	public := r.Group("/api/v1")
	{
		public.POST("/register", handlers.User.Register)
		public.POST("/login", handlers.User.Login)
		// WebSocket连接
		public.GET("/ws", handlers.Chat.HandleWebSocket)
	}

	// 需要认证的路由
	authorized := r.Group("/api/v1")
	authorized.Use(middleware.Auth(cfg))
	{
		// 用户相关
		authorized.GET("/user/profile", handlers.User.GetProfile)
		authorized.PUT("/user/updateprofile", handlers.User.UpdateProfile)
		authorized.GET("/user/search", handlers.User.SearchUser)

		// 好友相关路由
		authorized.POST("/friendship/request", handlers.Friendship.SendRequest)
		authorized.GET("/friendship/list", handlers.Friendship.GetFriendsList)

		// 聊天相关
		authorized.POST("/chat/message", handlers.Chat.SendMessage)
		authorized.POST("/chat/getAllLastMessages", handlers.Chat.getAllLastMessages)
		// authorized.GET("/chat/messages", handlers.Chat.GetMessages)
		authorized.PUT("/messages/:id/read", handlers.Message.MarkAsRead)
		authorized.PUT("/messages/read", handlers.Message.MarkMultipleAsRead)
		authorized.GET("/messages/unread", handlers.Message.GetUnreadMessages)
		authorized.GET("/groups/:group_id/messages/unread", handlers.Message.GetGroupUnreadMessages)
		authorized.PUT("/groups/messages/:id/read", handlers.Message.MarkGroupMessageAsRead)

		// 群聊相关
		authorized.POST("/group", handlers.Group.Create)
		authorized.GET("/groups", handlers.Group.List)

		authorized.GET("/group/:id", handlers.Group.Get)
		authorized.POST("/group/:id/join", handlers.Group.Join)
		authorized.POST("/group/:id/leave", handlers.Group.Leave)

		// 文件相关路由
		authorized.POST("/files", handlers.File.Upload)
		authorized.GET("/files", handlers.File.GetUserFiles)
		authorized.DELETE("/files/:id", handlers.File.Delete)

		// 通知相关路由
		authorized.GET("/notifications", handlers.Notification.GetNotifications)
		authorized.PUT("/notifications/:id/read", handlers.Notification.MarkAsRead)
		authorized.PUT("/notifications/read-all", handlers.Notification.MarkAllAsRead)
		authorized.DELETE("/notifications/:id", handlers.Notification.DeleteNotification)
		authorized.DELETE("/notifications", handlers.Notification.DeleteAllNotifications)

		// 在线状态相关路由
		authorized.GET("/online/users", handlers.Online.GetOnlineUsers)
		authorized.GET("/online/users/:id", handlers.Online.CheckUserOnline)
	}
}

// Handlers 结构体用于组织所有的处理器
type Handlers struct {
	User         *UserHandler
	Chat         *ChatHandler
	Group        *GroupHandler
	File         *FileHandler
	Notification *NotificationHandler
	Online       *OnlineHandler
	Message      *MessageHandler
	Friendship   *FriendshipHandler
}
