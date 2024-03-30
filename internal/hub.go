package internal

import (
	"encoding/base64"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/samaelpola/GoFM/internal/services"
	"io"
	"log"
	"time"
)

var (
	connectedUsers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "go_fm_connected_users",
		Help: "Number of users listening to each radio station",
	}, []string{"radio_station_name"})
)

func init() {
	for _, goFmStation := range services.GetListOfStation() {
		connectedUsers.WithLabelValues(goFmStation).Set(0)
	}
}

type Hub struct {
	name       string
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
}

func NewHub(name string) *Hub {
	return &Hub{
		name:       name,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			connectedUsers.WithLabelValues(h.name).Inc()
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				connectedUsers.WithLabelValues(h.name).Dec()
			}
		}
	}
}

func (h *Hub) sendChunkMusic(music any) {
	for client := range h.clients {
		go func(client *Client) {
			if err := client.conn.WriteJSON(music); err != nil {
				log.Println("error during send chunk: ", err)
			}
		}(client)
	}
}

func (h *Hub) goGenFm() {
	for {
		duration := 2 * time.Second
		musicType := services.GetCurrentMusicType()

		musics, err := services.GetMusics(musicType)
		if err != nil {
			log.Println("unable to get musics: ", err)
			return
		}

		for musicType == services.GetCurrentMusicType() {
			for _, music := range musics {
				if musicType != services.GetCurrentMusicType() {
					break
				}

				f, err := services.GetAudio(music.ID)
				if err != nil {
					log.Println(err)
					continue
				}

				trackDuration := services.GetTrackDuration(f)
				if _, err := f.Seek(0, io.SeekStart); err != nil {
					log.Println(err)
					continue
				}

				byteToRead := (int(f.Size()) * int(duration.Seconds())) / int(trackDuration.Seconds())
				b := make([]byte, byteToRead)

				for {
					if musicType != services.GetCurrentMusicType() {
						break
					}

					n, err := f.Read(b)
					if err != nil {
						if errors.Is(err, io.EOF) {
							break
						}

						log.Println("Error during read file: ", err)
						break
					}

					h.sendChunkMusic(struct {
						Id          int
						Name        string
						Title       string
						TypeOfMusic string
						Status      int
						Audio       string
					}{
						music.ID,
						music.Name,
						music.Title,
						music.Type,
						0,
						base64.StdEncoding.EncodeToString(b[:n]),
					})

					time.Sleep(duration - 1)
				}
			}
		}
	}
}

func (h *Hub) stream() {
	duration := 2 * time.Second

	if h.name == services.GOGEN {
		h.goGenFm()
	}

	musics, err := services.GetMusics(h.name)
	if err != nil {
		log.Println("unable to get musics: ", err)
		return
	}

	for {
		for _, music := range musics {
			f, err := services.GetAudio(music.ID)
			if err != nil {
				log.Printf("fail to get audio file with id %d: %s", music.ID, err)
				continue
			}

			trackDuration := services.GetTrackDuration(f)
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				log.Println(err)
				continue
			}

			byteToRead := (int(f.Size()) * int(duration.Seconds())) / int(trackDuration.Seconds())
			b := make([]byte, byteToRead)

			for {
				n, err := f.Read(b)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}

					log.Println("Error during read audio file: ", err)
					continue
				}

				h.sendChunkMusic(struct {
					Id          int
					Name        string
					Title       string
					TypeOfMusic string
					Status      int
					Audio       string
				}{
					music.ID,
					music.Name,
					music.Title,
					music.Type,
					0,
					base64.StdEncoding.EncodeToString(b[:n]),
				})

				time.Sleep(duration - 1)
			}
		}
	}
}
