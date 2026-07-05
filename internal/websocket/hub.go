package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/user/network-monitoring/internal/auth"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow CORS upgrade
	},
}

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	OrgID          uuid.UUID
	Send           chan []byte
}

type Hub struct {
	// Registered clients grouped by Organization UUID
	Rooms      map[uuid.UUID]map[*Client]bool
	Broadcast  chan BroadcastEvent
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

type BroadcastEvent struct {
	OrgID   uuid.UUID       `json:"-"`
	Type    string          `json:"type"` // "device_status", "ping_result", "alert", "stats"
	Payload interface{}     `json:"payload"`
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[uuid.UUID]map[*Client]bool),
		Broadcast:  make(chan BroadcastEvent, 1000),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.OrgID] == nil {
				h.Rooms[client.OrgID] = make(map[*Client]bool)
			}
			h.Rooms[client.OrgID][client] = true
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if room, exists := h.Rooms[client.OrgID]; exists {
				if _, ok := room[client]; ok {
					delete(room, client)
					close(client.Send)
					if len(room) == 0 {
						delete(h.Rooms, client.OrgID)
					}
				}
			}
			h.mu.Unlock()

		case event := <-h.Broadcast:
			h.mu.RLock()
			room, exists := h.Rooms[event.OrgID]
			if exists {
				msgBytes, err := json.Marshal(event)
				if err == nil {
					for client := range room {
						select {
						case client.Send <- msgBytes:
						default:
							// Handle slow consumers by unregistering them
							go func(c *Client) {
								h.Unregister <- c
								c.Conn.Close()
							}(client)
						}
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}

func HandleWS(hub *Hub, c *gin.Context) {
	// Verify JWT token passed via query parameter (e.g. ?token=...)
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token query parameter is required"})
		return
	}

	claims, err := auth.ValidateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		Hub:   hub,
		Conn:  conn,
		OrgID: claims.OrganizationID,
		Send:  make(chan []byte, 256),
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
