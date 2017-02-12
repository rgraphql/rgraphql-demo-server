package main

import (
	"context"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/googollee/go-socket.io"
	"github.com/rgraphql/magellan"
	pb "github.com/rgraphql/rgraphql/pkg/proto"
)

type WSClient struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	server      *WebServer
	disposeOnce sync.Once
	disposeCb   func(id string)

	id         string
	sock       socketio.Socket
	client     *magellan.ClientInstance
	sendCh     <-chan *pb.RGQLServerMessage
	setUseJson chan<- bool
}

func (c *WSClient) initHandlers(disposed func(id string)) {
	c.disposeCb = disposed
	s := c.sock
	s.On("sb", c.handleMessageRaw)
	s.On("disconnection", c.disconnected)
}

func (c *WSClient) handleMessage(msg *pb.RGQLClientMessage, wasJson bool) {
	c.setUseJson <- wasJson
	c.client.HandleMessage(msg)
}

func (c *WSClient) handleMessageRaw(b string) {
	le := log.WithField("client", c.sock.Id())
	msg, err := ParseMessage([]byte(b))
	if err != nil {
		le.WithError(err).Warn("Unable to parse message")
		return
	}
	c.handleMessage(msg.ClientMessage, len(msg.Base64Payload) == 0)
}

func (c *WSClient) sendWorker() {
	done := c.ctx.Done()
	suj := make(chan bool, 1)
	useJson := false
	c.setUseJson = suj
	for {
		select {
		case msg := <-c.sendCh:
			emsg, err := NewMessage(msg, useJson)
			if err != nil {
				log.WithError(err).Warn("Unable to build message")
				continue
			}
			data, err := emsg.Marshal()
			if err != nil {
				log.WithError(err).Warn("Unable to marshal message")
				continue
			}
			err = c.sock.Emit("sb", data)
			if err != nil {
				log.WithError(err).Warn("Unable to transmit message")
				continue
			}
		case uj := <-suj:
			useJson = uj
		case <-done:
			return
		}
	}
}

func (c *WSClient) disconnected() {
	c.disposed()
}

func (c *WSClient) disposed() {
	c.disposeOnce.Do(func() {
		c.ctxCancel()
		c.disposeCb(c.id)
	})
}
