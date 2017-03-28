package main

import (
	"context"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/rgraphql/magellan"
	proto "github.com/rgraphql/rgraphql/pkg/proto"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type WSClient struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	id        uint32
	server    *WebServer
	sock      *websocket.Conn
	sendQueue chan *proto.RGQLServerMessage
	gqlClient *magellan.ClientInstance

	disposeOnce sync.Once
	disposeCb   func(id uint32)
}

func (c *WSClient) initHandlers(disposed func(id uint32)) {
	c.disposeCb = disposed
}

func (c *WSClient) handleMessage(msg *proto.RGQLClientMessage) error {
	c.gqlClient.HandleMessage(msg)
	return nil
}

func (c *WSClient) handleMessageRaw(b []byte) {
	le := log.WithField("client", c.id)
	res := &proto.RGQLClientMessage{}
	err := pb.Unmarshal(b, res)
	if err != nil {
		le.WithError(err).Warn("Unable to deserialize")
		return
	}
	if err := c.handleMessage(res); err != nil {
		le.WithError(err).Error("Error handling message")
	}
}

func (c *WSClient) marshalMessage(msg *proto.RGQLServerMessage) (data []byte, err error) {
	defer func() {
		if err != nil {
			log.WithError(err).Error("Unable to serialize sb message (dropping)")
		}
	}()

	mj, err := pb.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return mj, nil
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.sock.WriteMessage(websocket.CloseMessage, []byte{})
		ticker.Stop()
		c.sock.Close()
	}()

	for {
		select {
		case msg, ok := <-c.sendQueue:
			c.sock.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}

			msgb, err := c.marshalMessage(msg)
			if err != nil {
				continue
			}

			w, err := c.sock.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(msgb)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.sock.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.sock.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *WSClient) readPump() {
	defer func() {
		c.sock.Close()
		c.disconnected()
	}()
	// c.sock.SetReadLimit(maxMessageSize)
	c.sock.SetReadDeadline(time.Now().Add(pongWait))
	c.sock.SetPongHandler(func(string) error { c.sock.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.sock.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.WithError(err).Warn("Unexpected close")
			}
			break
		}
		c.handleMessageRaw(message)
	}
}

func (c *WSClient) disconnected() {
	c.disposed()
}

func (c *WSClient) disposed() {
	c.disposeOnce.Do(func() {
		c.disposeCb(c.id)
	})
}
