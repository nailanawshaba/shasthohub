// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"golang.org/x/net/context"

	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
)

type LogHandler struct {
	logReg *logRegister
	libkb.Contextified
	ui LogUICn
}

// LogRPCHandler is the RPC handler for the log interface.
type LogRPCHandler struct {
	*BaseHandler
	*LogHandler
}

type LogUICn interface {
	GetLogUICli() *keybase1.LogUiClient
}

func NewLogHandler(logReg *logRegister, g *libkb.GlobalContext, ui LogUICn) *LogHandler {
	return &LogHandler{
		logReg:       logReg,
		Contextified: libkb.NewContextified(g),
		ui:           ui,
	}
}

// NewLogRPCHandler creates a LogHandler for the xp transport.
func NewLogRPCHandler(xp rpc.Transporter, logReg *logRegister, g *libkb.GlobalContext) *LogRPCHandler {
	handler := NewBaseHandler(xp)
	return &LogRPCHandler{
		BaseHandler: handler,
		LogHandler:  NewLogHandler(logReg, g, handler),
	}
}

func (h *LogHandler) RegisterLogger(_ context.Context, arg keybase1.RegisterLoggerArg) (err error) {
	h.G().Log.Debug("LogHandler::RegisterLogger: %+v", arg)
	defer h.G().Trace("LogHandler::RegisterLogger", func() error { return err })()

	if h.logReg == nil {
		// if not a daemon, h.logReg will be nil
		h.G().Log.Debug("- logRegister is nil, ignoring RegisterLogger request")
		return nil
	}

	ui := &LogUI{sessionID: arg.SessionID, cli: h.ui.GetLogUICli()}
	err = h.logReg.RegisterLogger(arg, ui)
	return err
}
