package api

import (
	"chatweb/internal/model"
	"chatweb/internal/service"
	"chatweb/pkg/event"
	"chatweb/pkg/websocket"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatHandler struct {
	messageService      *service.MessageService
	notificationService *service.NotificationService
	groupService        *service.GroupService
	onlineService       *service.OnlineService
	wsHub               *websocket.Hub
	upgrader            websocket.Upgrader
}

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
		wsHub:               websocket.NewHub(eventBus),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该更严格
			},
		},
	}
}

func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	client := websocket.NewClient(h.wsHub, conn, userID, h.onlineService)
	h.wsHub.Register(client)

	// 设置用户在线状态
	if err := h.onlineService.SetUserOnline(c.Request.Context(), userID); err != nil {
		log.Printf("Failed to set user online: %v", err)
	}

	// 启动客户端的读写goroutines
	go client.WritePump()
	go client.ReadPump()
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Type       string `json:"type" binding:"required"`
		Content    string `json:"content" binding:"required"`
		ReceiverID string `json:"receiver_id"`
		GroupID    string `json:"group_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	message := &model.Message{
		Type:      model.MessageType(req.Type),
		Content:   req.Content,
		SenderID:  senderObjID,
		Status:    "sent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

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

	if err := h.messageService.CreateMessage(c.Request.Context(), message); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var query struct {
		ReceiverID string `form:"receiver_id"`
		GroupID    string `form:"group_id"`
		Limit      int    `form:"limit,default=20"`
		Offset     int    `form:"offset,default=0"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var messages []*model.Message
	var err error

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

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
