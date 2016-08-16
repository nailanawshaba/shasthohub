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

type AccountHandler struct {
	libkb.Contextified
	ui AccountUI
}

type AccountRPCHandler struct {
	*BaseHandler
	*AccountHandler
}

var _ keybase1.AccountInterface = (*AccountHandler)(nil)

// AccountUI resolves UI for Account requests
type AccountUI interface {
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
}

func NewAccountHandler(g *libkb.GlobalContext, ui AccountUI) *AccountHandler {
	return &AccountHandler{
		Contextified: libkb.NewContextified(g),
	}
}

func NewAccountRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *AccountRPCHandler {
	handler := NewBaseHandler(xp)
	return &AccountRPCHandler{
		BaseHandler:    handler,
		AccountHandler: NewAccountHandler(g, handler),
	}
}

func (h *AccountHandler) PassphraseChange(_ context.Context, arg keybase1.PassphraseChangeArg) error {
	eng := engine.NewPassphraseChange(&arg, h.G())
	ctx := &engine.Context{
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	return engine.RunEngine(eng, ctx)
}

func (h *AccountHandler) PassphrasePrompt(_ context.Context, arg keybase1.PassphrasePromptArg) (keybase1.GetPassphraseRes, error) {
	ui := h.ui.GetSecretUI(arg.SessionID, h.G())
	if h.G().UIRouter != nil {
		delegateUI, err := h.G().UIRouter.GetSecretUI(arg.SessionID)
		if err != nil {
			return keybase1.GetPassphraseRes{}, err
		}
		if delegateUI != nil {
			ui = delegateUI
			h.G().Log.Debug("using delegate secret UI")
		}
	}

	return ui.GetPassphrase(arg.GuiArg, nil)
}
