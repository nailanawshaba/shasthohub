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

type RemotePgpUI struct {
	sessionID int
	cli       keybase1.PGPUiClient
}

func NewRemotePgpUI(sessionID int, c *rpc.Client) *RemotePgpUI {
	return &RemotePgpUI{
		sessionID: sessionID,
		cli:       keybase1.PGPUiClient{Cli: c},
	}
}

func (u *RemotePgpUI) OutputSignatureSuccess(ctx context.Context, arg keybase1.OutputSignatureSuccessArg) error {
	return u.cli.OutputSignatureSuccess(ctx, arg)
}

type PGPHandler struct {
	libkb.Contextified
	ui PGPUI
}

var _ keybase1.PGPInterface = (*PGPHandler)(nil)

// PGPUI resolves UI for PGP requests
type PGPUI interface {
	NewRemoteIdentifyUI(sessionID int, g *libkb.GlobalContext) *RemoteIdentifyUI
	NewRemoteSkipPromptIdentifyUI(sessionID int, g *libkb.GlobalContext) *RemoteIdentifyUI
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
	GetStreamUICli() *keybase1.StreamUiClient
	GetLogUI(sessionID int) libkb.LogUI
	GetPgpUI(sessionID int) libkb.PgpUI
	GetGPGUI(sessionID int) libkb.GPGUI
	GetLoginUI(sessionID int) libkb.LoginUI
}

type PGPRPCHandler struct {
	*BaseHandler
	*PGPHandler
}

func NewPGPHandler(g *libkb.GlobalContext, ui PGPUI) *PGPHandler {
	return &PGPHandler{
		Contextified: libkb.NewContextified(g),
	}
}

var _ keybase1.PGPInterface = (*PGPHandler)(nil)

func NewPGPRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *PGPRPCHandler {
	handler := NewBaseHandler(xp)
	return &PGPRPCHandler{
		BaseHandler: handler,
		PGPHandler:  NewPGPHandler(g, handler),
	}
}

func (h *PGPHandler) PGPSign(_ context.Context, arg keybase1.PGPSignArg) (err error) {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := engine.PGPSignArg{Sink: snk, Source: src, Opts: arg.Opts}
	ctx := engine.Context{
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewPGPSignEngine(&earg, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *PGPHandler) PGPPull(_ context.Context, arg keybase1.PGPPullArg) error {
	earg := engine.PGPPullEngineArg{
		UserAsserts: arg.UserAsserts,
	}
	ctx := engine.Context{
		LogUI:      h.ui.GetLogUI(arg.SessionID),
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewPGPPullEngine(&earg, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *PGPHandler) PGPEncrypt(_ context.Context, arg keybase1.PGPEncryptArg) error {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := &engine.PGPEncryptArg{
		Recips:       arg.Opts.Recipients,
		Sink:         snk,
		Source:       src,
		NoSign:       arg.Opts.NoSign,
		NoSelf:       arg.Opts.NoSelf,
		BinaryOutput: arg.Opts.BinaryOut,
		KeyQuery:     arg.Opts.KeyQuery,
	}
	ctx := &engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewPGPEncrypt(earg, h.G())
	return engine.RunEngine(eng, ctx)
}

func (h *PGPHandler) PGPDecrypt(_ context.Context, arg keybase1.PGPDecryptArg) (keybase1.PGPSigVerification, error) {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	snk := libkb.NewRemoteStreamBuffered(arg.Sink, cli, arg.SessionID)
	earg := &engine.PGPDecryptArg{
		Sink:         snk,
		Source:       src,
		AssertSigned: arg.Opts.AssertSigned,
		SignedBy:     arg.Opts.SignedBy,
	}
	ctx := &engine.Context{
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		IdentifyUI: h.ui.NewRemoteSkipPromptIdentifyUI(arg.SessionID, h.G()),
		LogUI:      h.ui.GetLogUI(arg.SessionID),
		PgpUI:      h.ui.GetPgpUI(arg.SessionID),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewPGPDecrypt(earg, h.G())
	err := engine.RunEngine(eng, ctx)
	if err != nil {
		return keybase1.PGPSigVerification{}, err
	}

	return sigVer(h.G(), eng.SignatureStatus(), eng.Owner()), nil
}

func (h *PGPHandler) PGPVerify(_ context.Context, arg keybase1.PGPVerifyArg) (keybase1.PGPSigVerification, error) {
	cli := h.ui.GetStreamUICli()
	src := libkb.NewRemoteStreamBuffered(arg.Source, cli, arg.SessionID)
	earg := &engine.PGPVerifyArg{
		Source:    src,
		Signature: arg.Opts.Signature,
		SignedBy:  arg.Opts.SignedBy,
	}
	ctx := &engine.Context{
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		LogUI:      h.ui.GetLogUI(arg.SessionID),
		PgpUI:      h.ui.GetPgpUI(arg.SessionID),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewPGPVerify(earg, h.G())
	err := engine.RunEngine(eng, ctx)
	if err != nil {
		return keybase1.PGPSigVerification{}, err
	}

	return sigVer(h.G(), eng.SignatureStatus(), eng.Owner()), nil
}

func sigVer(g *libkb.GlobalContext, ss *libkb.SignatureStatus, owner *libkb.User) keybase1.PGPSigVerification {
	var res keybase1.PGPSigVerification
	if ss.IsSigned {
		res.IsSigned = ss.IsSigned
		res.Verified = ss.Verified
		if owner != nil {
			signer := owner.Export()
			if signer != nil {
				res.Signer = *signer
			}
		}
		if ss.Entity != nil {
			bundle := libkb.NewPGPKeyBundle(g, ss.Entity)
			res.SignKey = bundle.Export()
		}
	}
	return res
}

func (h *PGPHandler) PGPImport(_ context.Context, arg keybase1.PGPImportArg) error {
	ctx := &engine.Context{
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		LogUI:     h.ui.GetLogUI(arg.SessionID),
		SessionID: arg.SessionID,
	}
	eng, err := engine.NewPGPKeyImportEngineFromBytes(arg.Key, arg.PushSecret, h.G())
	if err != nil {
		return err
	}
	err = engine.RunEngine(eng, ctx)
	return err
}

type exporter interface {
	engine.Engine
	Results() []keybase1.KeyInfo
}

func (h *PGPHandler) export(sessionID int, ex exporter) ([]keybase1.KeyInfo, error) {
	ctx := &engine.Context{
		SecretUI:  h.ui.GetSecretUI(sessionID, h.G()),
		LogUI:     h.ui.GetLogUI(sessionID),
		SessionID: sessionID,
	}
	if err := engine.RunEngine(ex, ctx); err != nil {
		return nil, err
	}
	return ex.Results(), nil
}

func (h *PGPHandler) PGPExport(_ context.Context, arg keybase1.PGPExportArg) (ret []keybase1.KeyInfo, err error) {
	return h.export(arg.SessionID, engine.NewPGPKeyExportEngine(arg, h.G()))
}

func (h *PGPHandler) PGPExportByKID(_ context.Context, arg keybase1.PGPExportByKIDArg) (ret []keybase1.KeyInfo, err error) {
	return h.export(arg.SessionID, engine.NewPGPKeyExportByKIDEngine(arg, h.G()))
}

func (h *PGPHandler) PGPExportByFingerprint(_ context.Context, arg keybase1.PGPExportByFingerprintArg) (ret []keybase1.KeyInfo, err error) {
	return h.export(arg.SessionID, engine.NewPGPKeyExportByFingerprintEngine(arg, h.G()))
}
func (h *PGPHandler) PGPKeyGen(_ context.Context, arg keybase1.PGPKeyGenArg) error {
	ctx := &engine.Context{
		LogUI:     h.ui.GetLogUI(arg.SessionID),
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	earg := engine.ImportPGPKeyImportEngineArg(arg)
	eng := engine.NewPGPKeyImportEngine(earg)
	return engine.RunEngine(eng, ctx)
}

func (h *PGPHandler) PGPDeletePrimary(_ context.Context, sessionID int) (err error) {
	return libkb.DeletePrimary()
}

func (h *PGPHandler) PGPSelect(_ context.Context, sarg keybase1.PGPSelectArg) error {
	arg := engine.GPGImportKeyArg{
		Query:      sarg.FingerprintQuery,
		AllowMulti: sarg.AllowMulti,
		SkipImport: sarg.SkipImport,
		OnlyImport: sarg.OnlyImport,
	}
	gpg := engine.NewGPGImportKeyEngine(&arg, h.G())
	ctx := &engine.Context{
		GPGUI:     h.ui.GetGPGUI(sarg.SessionID),
		SecretUI:  h.ui.GetSecretUI(sarg.SessionID, h.G()),
		LogUI:     h.ui.GetLogUI(sarg.SessionID),
		LoginUI:   h.ui.GetLoginUI(sarg.SessionID),
		SessionID: sarg.SessionID,
	}
	return engine.RunEngine(gpg, ctx)
}

func (h *PGPHandler) PGPUpdate(_ context.Context, arg keybase1.PGPUpdateArg) error {
	ctx := engine.Context{
		LogUI:     h.ui.GetLogUI(arg.SessionID),
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewPGPUpdateEngine(arg.Fingerprints, arg.All, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *PGPHandler) PGPPurge(ctx context.Context, arg keybase1.PGPPurgeArg) (keybase1.PGPPurgeRes, error) {
	ectx := &engine.Context{
		LogUI:      h.ui.GetLogUI(arg.SessionID),
		SessionID:  arg.SessionID,
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		NetContext: ctx,
	}
	eng := engine.NewPGPPurge(h.G(), arg)
	var res keybase1.PGPPurgeRes
	if err := engine.RunEngine(eng, ectx); err != nil {
		return res, err
	}
	res.Filenames = eng.KeyFiles()
	return res, nil
}
