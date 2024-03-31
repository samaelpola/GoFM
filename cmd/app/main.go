package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samaelpola/GoFM/internal"
	"log"
	"net/http"
	"time"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	var ws internal.WS
	ws.CreateAllGoFmStation()

	router.HandleFunc("/ws/{name}", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(mux.Vars(r)["name"], w, r)
	})

	cert, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
	if err != nil {
		log.Fatalf("Failed to load X509 key pair: %v", err)
	}

	srv := &http.Server{
		Handler:      router,
		Addr:         ":8082",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		TLSConfig:    &tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true},
	}

	go func() {
		fmt.Println("listening on 8082 ....")
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println("listening on 2112 ....")
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Fatal(err)
	}
}
