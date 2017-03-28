package main

import (
	"context"
	"net/http"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"github.com/rgraphql/magellan"
	proto "github.com/rgraphql/rgraphql/pkg/proto"
)

type WebServer struct {
	clients    map[uint32]*WSClient
	clientsMtx sync.RWMutex
	clientsCtr uint32

	gqlServer *magellan.Server
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

func NewWebServer(
	gqlServer *magellan.Server,
) *WebServer {
	return &WebServer{
		clients:   make(map[uint32]*WSClient),
		gqlServer: gqlServer,
	}
}

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("Unable to upgrade WS connection")
		return
	}

	sendQueue := make(chan *proto.RGQLServerMessage, 20)

	ws.clientsMtx.Lock()
	ws.clientsCtr++
	id := ws.clientsCtr

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	gc, err := ws.gqlServer.BuildClient(ctx, sendQueue, &RootQueryResolver{}, nil)
	if err != nil {
		log.WithError(err).Error("Unable to build client")
		return
	}
	client := &WSClient{
		id:        id,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		server:    ws,
		sock:      conn,
		sendQueue: sendQueue,
		gqlClient: gc,
	}
	ws.clients[id] = client
	ws.clientsMtx.Unlock()

	client.initHandlers(ws.clientDisposed)
	go client.writePump()
	client.readPump()
}

func (ws *WebServer) clientDisposed(id uint32) {
	ws.clientsMtx.Lock()
	delete(ws.clients, id)
	log.WithField("id", id).Debug("Client disconnected")
	ws.clientsMtx.Unlock()
}
