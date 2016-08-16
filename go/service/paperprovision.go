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

type PaperProvisionHandler struct {
	libkb.Contextified
	ui PaperProvisionUICn
}

type PaperProvisionRPCHandler struct {
	*BaseHandler
	*PaperProvisionHandler
}

// PaperProvisionUICn resolves UI for login requests
type PaperProvisionUICn interface {
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
	GetLogUI(sessionID int) libkb.LogUI
	GetLoginUI(sessionID int) libkb.LoginUI
	GetProvisionUI(sessionID int) libkb.ProvisionUI
	GetGPGUI(sessionID int) libkb.GPGUI
}

func NewPaperProvisionHandler(g *libkb.GlobalContext, ui PaperProvisionUICn) *PaperProvisionHandler {
	return &PaperProvisionHandler{
		Contextified: libkb.NewContextified(g),
		ui:           ui,
	}
}

func NewPaperProvisionRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *PaperProvisionRPCHandler {
	handler := NewBaseHandler(xp)
	return &PaperProvisionRPCHandler{
		BaseHandler:           handler,
		PaperProvisionHandler: NewPaperProvisionHandler(g, handler),
	}
}

func (h *PaperProvisionHandler) PaperProvision(ctx context.Context, arg keybase1.PaperProvisionArg) error {

	ectx := engine.Context{
		NetContext:  ctx,
		LogUI:       h.ui.GetLogUI(arg.SessionID),
		SecretUI:    h.ui.GetSecretUI(arg.SessionID, h.G()),
		LoginUI:     h.ui.GetLoginUI(arg.SessionID),
		ProvisionUI: h.ui.GetProvisionUI(arg.SessionID),
		SessionID:   arg.SessionID,
	}
	eng := engine.NewPaperProvisionEngine(h.G(), arg.Username, arg.DeviceName, arg.PaperKey)
	err := engine.RunEngine(eng, &ectx)
	if err != nil {
		return err
	}
	return nil
}
