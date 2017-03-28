package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/rgraphql/magellan"
	"github.com/rs/cors"
)

var ListenStr string = ":3001"

func main() {
	log.SetLevel(log.DebugLevel)

	schema, err := magellan.ParseSchema(schemaAst, &RootQueryResolver{}, nil)
	if err != nil {
		panic(err)
	}

	server := NewWebServer(schema)
	mux := http.NewServeMux()
	mux.Handle("/sock", server)

	log.WithField("listen", ListenStr).Info("Listening")
	handler := cors.New(cors.Options{
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"},
	}).Handler(mux)
	if err := http.ListenAndServe(ListenStr, handler); err != nil {
		panic(err)
	}
}
