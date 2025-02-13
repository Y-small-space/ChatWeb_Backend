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
	wsHub               *websocketM.Hub               // WebSocket Hub
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
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建 WebSocket 客户端并注册到 Hub
	client := websocketM.NewClient(h.wsHub, conn, userID, h.onlineService)
	h.wsHub.Register(client)

	// 设置用户为在线状态
	// if err := h.onlineService.SetUserOnline(c.Request.Context(), userID); err != nil {
	// 	log.Printf("Failed to set user online: %v", err)
	// }

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

	// 设置接收方（个人或群组）
	if req.GroupID != "" {
		groupObjID, err := primitive.ObjectIDFromHex(req.GroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
			return
		}
		message.GroupID = groupObjID
	} else if req.ReceiverID != "" {
		receiverObjID, err := primitive.ObjectIDFromHex(req.ReceiverID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receiver ID"})
			return
		}
		message.ReceiverID = receiverObjID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either receiver_id or group_id is required"})
		return
	}

	// 保存消息
	if err := h.messageService.CreateMessage(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回消息
	c.JSON(http.StatusOK, gin.H{"message": message})
}

// GetMessages 获取消息
func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var query struct {
		ReceiverID string `form:"receiver_id"` // 接收者 ID
		GroupID    string `form:"group_id"`    // 群组 ID
		Limit      int    `form:"limit,default=20"`
		Offset     int    `form:"offset,default=0"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var messages []*model.Message
	var err error

	// 根据查询参数选择不同的消息获取方法
	if query.GroupID != "" {
		messages, err = h.messageService.GetGroupMessages(c.Request.Context(), query.GroupID, query.Limit, query.Offset)
	} else if query.ReceiverID != "" {
		messages, err = h.messageService.GetUserMessages(c.Request.Context(), userID, query.ReceiverID, query.Limit, query.Offset)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either receiver_id or group_id is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回消息列表
	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
