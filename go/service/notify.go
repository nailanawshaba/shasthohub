// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	context "golang.org/x/net/context"
)

// NotifyCtlRPCHandler is the RPC handler for notify control messages
type NotifyCtlRPCHandler struct {
	libkb.Contextified
	*BaseHandler
	id libkb.ConnectionID
}

// NewNotifyCtlRPCHandler creates a new handler for setting up notification
// channels
func NewNotifyCtlRPCHandler(xp rpc.Transporter, id libkb.ConnectionID, g *libkb.GlobalContext) *NotifyCtlRPCHandler {
	return &NotifyCtlRPCHandler{
		Contextified: libkb.NewContextified(g),
		BaseHandler:  NewBaseHandler(xp),
		id:           id,
	}
}

func (h *NotifyCtlRPCHandler) SetNotifications(_ context.Context, n keybase1.NotificationChannels) error {
	h.G().NotifyRouter.SetChannels(h.id, n)
	return nil
}
