package internal

import (
	"github.com/gorilla/websocket"
	"github.com/samaelpola/GoFM/internal/services"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WS struct {
	goFmStations []*Hub
}

func (ws *WS) CreateAllGoFmStation() {
	for _, goFmStation := range services.GetListOfStation() {
		hub := NewHub(goFmStation)
		ws.goFmStations = append(ws.goFmStations, hub)

		go hub.run()
		go hub.stream()
	}
}

func (ws *WS) getHubByName(name string) *Hub {
	for _, goFmStation := range ws.goFmStations {
		if goFmStation.name == name {
			return goFmStation
		}
	}

	return nil
}

func (ws *WS) ServeWs(goFmStation string, w http.ResponseWriter, r *http.Request) {
	hub := ws.getHubByName(goFmStation)
	if hub == nil {
		log.Printf("Error: No hub found for goFm station %s\n", goFmStation)
		http.Error(w, "No such room", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	client := &Client{hub: hub, conn: conn}
	client.hub.register <- client
	go client.checkConnection()
}
