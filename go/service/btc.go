// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	"golang.org/x/net/context"
)

type BTCHandler struct {
	libkb.Contextified
	ui BTCUI
}

type BTCRPCHandler struct {
	*BaseHandler
	*BTCHandler
}

// BTCUI resolves UI for login requests
type BTCUI interface {
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
	GetLogUI(sessionID int) libkb.LogUI
}

func NewBTCHandler(g *libkb.GlobalContext, ui BTCUI) *BTCHandler {
	return &BTCHandler{
		Contextified: libkb.NewContextified(g),
	}
}

func NewBTCRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *BTCRPCHandler {
	handler := NewBaseHandler(xp)
	return &BTCRPCHandler{
		BaseHandler: handler,
		BTCHandler:  NewBTCHandler(g, handler),
	}
}

// BTC creates a BTCEngine and runs it.
func (h *BTCHandler) RegisterBTC(_ context.Context, arg keybase1.RegisterBTCArg) error {
	ctx := engine.Context{
		LogUI:     h.ui.GetLogUI(arg.SessionID),
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewBTCEngine(arg.Address, arg.Force, h.G())
	return engine.RunEngine(eng, &ctx)
}
