package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StreamLogs(c *gin.Context) {
	containerName := c.Param("name")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// Mock streaming logs
	for i := 0; i < 10; i++ {
		msg := []byte("Mock log line " + time.Now().Format(time.RFC3339) + " from " + containerName + "\n")
		if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
