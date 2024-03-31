package internal

import (
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
}

func (c *Client) checkConnection() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("connection ws close: %v", err)
			}
			break
		}
	}
}
