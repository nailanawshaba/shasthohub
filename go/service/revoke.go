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

type RevokeHandler struct {
	libkb.Contextified
	ui RevokeUI
}

var _ keybase1.RevokeInterface = (*RevokeHandler)(nil)

type RevokeRPCHandler struct {
	*BaseHandler
	*RevokeHandler
}

// RevokeUI resolves UI for revoke requests
type RevokeUI interface {
	GetLogUI(sessionID int) libkb.LogUI
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
}

func NewRevokeHandler(g *libkb.GlobalContext, ui RevokeUI) *RevokeHandler {
	return &RevokeHandler{
		Contextified: libkb.NewContextified(g),
		ui:           ui,
	}
}

func NewRevokeRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *RevokeRPCHandler {
	handler := NewBaseHandler(xp)
	return &RevokeRPCHandler{
		BaseHandler:   handler,
		RevokeHandler: NewRevokeHandler(g, handler),
	}
}

func (h *RevokeHandler) RevokeKey(_ context.Context, arg keybase1.RevokeKeyArg) error {
	sessionID := arg.SessionID
	ctx := engine.Context{
		LogUI:     h.ui.GetLogUI(sessionID),
		SecretUI:  h.ui.GetSecretUI(sessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewRevokeKeyEngine(arg.KeyID, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *RevokeHandler) RevokeDevice(_ context.Context, arg keybase1.RevokeDeviceArg) error {
	sessionID := arg.SessionID
	ctx := engine.Context{
		LogUI:     h.ui.GetLogUI(sessionID),
		SecretUI:  h.ui.GetSecretUI(sessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewRevokeDeviceEngine(engine.RevokeDeviceEngineArgs{ID: arg.DeviceID, Force: arg.Force}, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *RevokeHandler) RevokeSigs(_ context.Context, arg keybase1.RevokeSigsArg) error {
	ctx := engine.Context{
		LogUI:     h.ui.GetLogUI(arg.SessionID),
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewRevokeSigsEngine(arg.SigIDQueries, h.G())
	return engine.RunEngine(eng, &ctx)
}
