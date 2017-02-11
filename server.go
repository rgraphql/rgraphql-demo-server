package main

import (
	"context"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/googollee/go-socket.io"
	"github.com/rgraphql/magellan"
	proto "github.com/rgraphql/rgraphql/pkg/proto"
)

type WebServer struct {
	WSServer      *socketio.Server
	Clients       map[string]*WSClient
	ClientsMtx    sync.RWMutex
	GraphQLServer *magellan.Server
	Context       context.Context
	ContextCancel context.CancelFunc
}

func NewWebServer() (*WebServer, error) {
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	gsrv, err := magellan.ParseSchema(schemaAst, &RootQueryResolver{})
	if err != nil {
		return nil, err
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	ws := &WebServer{
		WSServer:      server,
		Clients:       make(map[string]*WSClient),
		GraphQLServer: gsrv,
		Context:       ctx,
		ContextCancel: ctxCancel,
	}
	ws.initHandlers()

	return ws, nil
}

func (ws *WebServer) initHandlers() {
	s := ws.WSServer
	s.On("connection", func(so socketio.Socket) {
		id := so.Id()

		ws.ClientsMtx.Lock()
		defer ws.ClientsMtx.Unlock()

		if _, ok := ws.Clients[id]; ok {
			return
		}

		ctx, ctxCancel := context.WithCancel(ws.Context)
		sendCh := make(chan *proto.RGQLServerMessage, 10)

		client, err := ws.GraphQLServer.BuildClient(ctx, sendCh)
		if err != nil {
			log.WithError(err).Warn("Unable to build GraphQL client")
			ctxCancel()
			return
		}

		nc := &WSClient{
			ctx:       ctx,
			ctxCancel: ctxCancel,
			server:    ws,
			id:        id,
			sock:      so,
			client:    client,
			sendCh:    sendCh,
		}
		go nc.sendWorker()

		log.WithField("id", id).Debug("Client connected")
		ws.Clients[id] = nc
		nc.initHandlers(ws.clientDisposed)
	})
}

func (ws *WebServer) clientDisposed(id string) {
	ws.ClientsMtx.Lock()
	cli, ok := ws.Clients[id]
	if ok {
		cli.ctxCancel()
		delete(ws.Clients, id)
	}
	log.WithField("id", id).Debug("Client disconnected")
	ws.ClientsMtx.Unlock()
}
