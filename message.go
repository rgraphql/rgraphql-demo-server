package main

import (
	"encoding/base64"
	"encoding/json"

	pb "github.com/golang/protobuf/proto"
	proto "github.com/rgraphql/rgraphql/pkg/proto"
)

// Wrapper for a message on the stream.
type Message struct {
	// Socket bus binary data, encoded in base64
	Base64Payload string `json:"b,omitempty"`
	// If we want json-only encoding, it will come in as such
	ClientMessage *proto.RGQLClientMessage `json:"jc,omitempty"`
	ServerMessage *proto.RGQLServerMessage `json:"js,omitempty"`
}

func NewMessage(msg *proto.RGQLServerMessage, useJson bool) (*Message, error) {
	res := &Message{}
	if useJson {
		res.ServerMessage = msg
	} else {
		bin, err := pb.Marshal(msg)
		if err != nil {
			return nil, err
		}
		res.Base64Payload = base64.StdEncoding.EncodeToString(bin)
	}

	return res, nil
}

func ParseMessage(data []byte) (*Message, error) {
	sbm := &Message{}
	if err := json.Unmarshal(data, sbm); err != nil {
		return nil, err
	}
	if sbm.ClientMessage == nil {
		bin, err := base64.StdEncoding.DecodeString(sbm.Base64Payload)
		if err != nil {
			return nil, err
		}
		sbm.ClientMessage = &proto.RGQLClientMessage{}
		err = pb.Unmarshal(bin, sbm.ClientMessage)
		if err != nil {
			return nil, err
		}
	}
	return sbm, nil
}

func (m *Message) Marshal() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
