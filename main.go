package main

import (
	"log"

	"chatweb/api"
	"chatweb/config"
	"chatweb/internal/repository"
	"chatweb/internal/repository/mongodb"
	"chatweb/internal/service"
	"chatweb/middleware"
	"chatweb/pkg/event"
	"chatweb/pkg/storage"
	"chatweb/pkg/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	cfg := config.LoadConfig()

	// 初始化数据库连接
	mongodb.InitMongoDB(cfg.MongoDB.URI, cfg.MongoDB.Database)

	// 初始化MinIO客户端
	minioClient, err := storage.NewMinioClient(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.Bucket,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// 初始化仓库
	userRepo := repository.NewUserRepository()
	messageRepo := repository.NewMessageRepository()
	groupRepo := repository.NewGroupRepository()
	fileRepo := repository.NewFileRepository()
	notificationRepo := repository.NewNotificationRepository()
	friendshipRepo := repository.NewFriendshipRepository()

	// 创建事件总线
	eventBus := event.NewEventBus()

	// 初始化服务
	userService := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireTime)
	messageService := service.NewMessageService(messageRepo, nil, eventBus)
	groupService := service.NewGroupService(groupRepo)
	fileService := service.NewFileService(fileRepo, minioClient)
	notificationService := service.NewNotificationService(notificationRepo, eventBus)
	friendshipService := service.NewFriendshipService(friendshipRepo, userRepo)
	// 创建WebSocket hub
	wsHub := websocket.NewHub(eventBus)
	onlineService := service.NewOnlineService(userRepo, eventBus)
	go wsHub.Run()

	// 初始化处理器
	userHandler := api.NewUserHandler(userService)
	chatHandler := api.NewChatHandler(messageService, notificationService, groupService, onlineService, eventBus)
	groupHandler := api.NewGroupHandler(groupService)
	fileHandler := api.NewFileHandler(fileService)
	notificationHandler := api.NewNotificationHandler(notificationService)
	onlineHandler := api.NewOnlineHandler(onlineService)
	friendshipHandler := api.NewFriendshipHandler(friendshipService)

	// 设置gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.Default()

	// 添加全局中间件
	r.Use(middleware.Cors())
	r.Use(middleware.Logger())

	// 初始化路由处理器
	handlers := &api.Handlers{
		User:         userHandler,
		Chat:         chatHandler,
		Group:        groupHandler,
		File:         fileHandler,
		Notification: notificationHandler,
		Online:       onlineHandler,
		Friendship:   friendshipHandler,
	}

	// 初始化路由
	api.InitRoutes(r, cfg, handlers)

	// 启动服务器
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
