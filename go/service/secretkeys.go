// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"errors"

	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	"golang.org/x/net/context"
)

type SecretKeysRPCHandler struct {
	*BaseHandler
	libkb.Contextified
}

func NewSecretKeysRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *SecretKeysRPCHandler {
	return &SecretKeysRPCHandler{
		BaseHandler:  NewBaseHandler(xp),
		Contextified: libkb.NewContextified(g),
	}
}

func (h *SecretKeysRPCHandler) GetSecretKeys(_ context.Context, sessionID int) (keybase1.SecretKeys, error) {
	if h.G().Env.GetRunMode() == libkb.ProductionRunMode {
		return keybase1.SecretKeys{}, errors.New("GetSecretKeys is a devel-only RPC")
	}
	ctx := engine.Context{
		LogUI:     h.GetLogUI(sessionID),
		SecretUI:  h.GetSecretUI(sessionID, h.G()),
		SessionID: sessionID,
	}
	eng := engine.NewSecretKeysEngine(h.G())
	err := engine.RunEngine(eng, &ctx)
	if err != nil {
		return keybase1.SecretKeys{}, err
	}
	return eng.Result(), nil
}
