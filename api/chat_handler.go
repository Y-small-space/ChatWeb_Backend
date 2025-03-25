package api

import (
	"chatweb/internal/model"
	"chatweb/internal/service"
	"chatweb/pkg/event"
	"chatweb/pkg/websocketM"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatHandler 用于处理与聊天相关的 HTTP 请求
type ChatHandler struct {
	messageService      *service.MessageService      // 消息服务
	notificationService *service.NotificationService // 通知服务
	groupService        *service.GroupService        // 群组服务
	onlineService       *service.OnlineService       // 在线状态服务
	wsHub               *websocketM.Hub              // WebSocket Hub
	upgrader            websocket.Upgrader           // WebSocket 升级器
}

// NewChatHandler 创建一个新的 ChatHandler 实例
func NewChatHandler(
	messageService *service.MessageService,
	notificationService *service.NotificationService,
	groupService *service.GroupService,
	onlineService *service.OnlineService,
	eventBus *event.EventBus,
) *ChatHandler {
	return &ChatHandler{
		messageService:      messageService,
		notificationService: notificationService,
		groupService:        groupService,
		onlineService:       onlineService,
		wsHub:               websocketM.NewHub(eventBus), // 初始化 WebSocket Hub
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024, // 设置 WebSocket 读取缓冲区大小
			WriteBufferSize: 1024, // 设置 WebSocket 写入缓冲区大小
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应更加严格地检查 Origin
			},
		},
	}
}

// HandleWebSocket 处理 WebSocket 连接的建立和消息处理
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	log.Println("WebSocket connection established")
	userID := c.DefaultQuery("userId", "")
	log.Println(userID)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	log.Printf("User %s connected via WebSocket", userID)

	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建 WebSocket 客户端并注册到 Hub
	client := websocketM.NewClient(h.wsHub, conn, userID, h.onlineService)
	h.wsHub.Register(client)

	log.Printf("User %s connected", userID)

	// 启动客户端的读写协程
	go client.WritePump()
	go client.ReadPump()
}

// SendMessage 发送消息
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Type       string `json:"type" binding:"required"`    // 消息类型
		Content    string `json:"content" binding:"required"` // 消息内容
		ReceiverID string `json:"receiver_id"`                // 接收者 ID
		GroupID    string `json:"group_id"`                   // 群组 ID
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为 ObjectID
	senderObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 构建消息对象
	message := &model.Message{
		Type:      model.MessageType(req.Type),
		Content:   req.Content,
		SenderID:  senderObjID,
		Status:    "sent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存消息
	if err := h.messageService.CreateMessage(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// getMessagesById 根据userID获取与当前用户的所有消息
func (h *ChatHandler) getMessagesById(c *gin.Context) {
	var requestBody struct {
		UserID  string `json:"userId"`  // 解析 JSON 请求体中的 userId
		OtherID string `json:"otherId"` // 解析
	}

	// 解析 JSON 数据
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userId := requestBody.UserID
	otherId := requestBody.OtherID

	log.Println("userId", userId, "otherId", otherId)

	messages, err := h.messageService.GetMessagesById(c.Request.Context(), userId, otherId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func (h *ChatHandler) getAllLastMessages(c *gin.Context) {
	var requestBody struct {
		UserID string `json:"userId"` // 解析 JSON 请求体中的 userId
	}

	// 解析 JSON 数据
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 确保 userId 不是空的
	if requestBody.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UserID is required"})
		return
	}

	// 输出 userId（测试）
	log.Printf("Received userId: %s", requestBody.UserID)

	userId := requestBody.UserID

	if userId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	messages, err := h.messageService.GetAllLastMessages(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
