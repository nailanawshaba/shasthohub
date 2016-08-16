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

type SaltpackHandler struct {
	libkb.Contextified
	ui SaltpackUI
}

var _ keybase1.SaltpackInterface = (*SaltpackHandler)(nil)

type SaltpackRPCHandler struct {
	*BaseHandler
	*SaltpackHandler
}

// SaltpackUI resolves UI for Saltpack requests
type SaltpackUI interface {
	GetStreamUICli() *keybase1.StreamUiClient
	GetSaltpackUI(sessionID int) libkb.SaltpackUI
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
	NewRemoteIdentifyUI(sessionID int, g *libkb.GlobalContext) *RemoteIdentifyUI
}

type RemoteSaltpackUI struct {
	sessionID int
	cli       keybase1.SaltpackUiClient
}

func NewRemoteSaltpackUI(sessionID int, c *rpc.Client) *RemoteSaltpackUI {
	return &RemoteSaltpackUI{
		sessionID: sessionID,
		cli:       keybase1.SaltpackUiClient{Cli: c},
	}
}

func (r *RemoteSaltpackUI) SaltpackPromptForDecrypt(ctx context.Context, arg keybase1.SaltpackPromptForDecryptArg, usedDelegateUI bool) (err error) {
	arg.SessionID = r.sessionID
	arg.UsedDelegateUI = usedDelegateUI
	return r.cli.SaltpackPromptForDecrypt(ctx, arg)
}

func (r *RemoteSaltpackUI) SaltpackVerifySuccess(ctx context.Context, arg keybase1.SaltpackVerifySuccessArg) (err error) {
	arg.SessionID = r.sessionID
	return r.cli.SaltpackVerifySuccess(ctx, arg)
}

func NewSaltpackHandler(g *libkb.GlobalContext, ui SaltpackUI) *SaltpackHandler {
	return &SaltpackHandler{
		Contextified: libkb.NewContextified(g),
		ui:           ui,
	}
}

func NewSaltpackRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *SaltpackRPCHandler {
	handler := NewBaseHandler(xp)
	return &SaltpackRPCHandler{
		BaseHandler:     handler,
		SaltpackHandler: NewSaltpackHandler(g, handler),
	}
}

func (h *SaltpackHandler) SaltpackDecrypt(_ context.Context, arg keybase1.SaltpackDecryptArg) (info keybase1.SaltpackEncryptedMessageInfo, err error) {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := &engine.SaltpackDecryptArg{
		Sink:   snk,
		Source: src,
		Opts:   arg.Opts,
	}

	ctx := &engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SaltpackUI: h.ui.GetSaltpackUI(arg.SessionID),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewSaltpackDecrypt(earg, h.G())
	err = engine.RunEngine(eng, ctx)
	info = eng.MessageInfo()
	return info, err
}

func (h *SaltpackHandler) SaltpackEncrypt(_ context.Context, arg keybase1.SaltpackEncryptArg) error {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := &engine.SaltpackEncryptArg{
		Opts:   arg.Opts,
		Sink:   snk,
		Source: src,
	}

	ctx := &engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewSaltpackEncrypt(earg, h.G())
	return engine.RunEngine(eng, ctx)
}

func (h *SaltpackHandler) SaltpackSign(_ context.Context, arg keybase1.SaltpackSignArg) error {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := &engine.SaltpackSignArg{
		Opts:   arg.Opts,
		Sink:   snk,
		Source: src,
	}

	ctx := &engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewSaltpackSign(earg, h.G())
	return engine.RunEngine(eng, ctx)
}

func (h *SaltpackHandler) SaltpackVerify(_ context.Context, arg keybase1.SaltpackVerifyArg) error {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := &engine.SaltpackVerifyArg{
		Opts:   arg.Opts,
		Sink:   snk,
		Source: src,
	}

	ctx := &engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SaltpackUI: h.ui.GetSaltpackUI(arg.SessionID),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewSaltpackVerify(earg, h.G())
	return engine.RunEngine(eng, ctx)
}
