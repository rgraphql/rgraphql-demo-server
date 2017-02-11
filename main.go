package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rs/cors"
)

var ListenStr string = ":3001"

func main() {
	log.SetLevel(log.DebugLevel)

	server, err := NewWebServer()
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", server.WSServer)

	log.WithField("listen", ListenStr).Info("Listening")
	handler := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
	}).Handler(mux)
	if err := http.ListenAndServe(ListenStr, handler); err != nil {
		panic(err)
	}
}
