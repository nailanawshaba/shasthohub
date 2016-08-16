// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"time"

	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	"golang.org/x/net/context"
)

// TrackHandler implements keybase1.TrackInterface
type TrackHandler struct {
	libkb.Contextified

	lastCheckTime time.Time
	ui            TrackUI
}

type TrackRPCHandler struct {
	*BaseHandler
	*TrackHandler
}

var _ keybase1.TrackInterface = (*TrackHandler)(nil)

// TrackUI resolves UI for track requests
type TrackUI interface {
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
	NewRemoteIdentifyUI(sessionID int, g *libkb.GlobalContext) *RemoteIdentifyUI
}

// NewTrackHandler creates a TrackHandler
func NewTrackHandler(g *libkb.GlobalContext, ui TrackUI) *TrackHandler {
	return &TrackHandler{
		Contextified: libkb.NewContextified(g),
		ui:           ui,
	}
}

// NewTrackRPCHandler creates a TrackHandler for the xp transport.
func NewTrackRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *TrackRPCHandler {
	handler := NewBaseHandler(xp)
	return &TrackRPCHandler{
		BaseHandler:  handler,
		TrackHandler: NewTrackHandler(g, handler),
	}
}

// Track creates a TrackEngine and runs it.
func (h *TrackHandler) Track(_ context.Context, arg keybase1.TrackArg) error {
	earg := engine.TrackEngineArg{
		UserAssertion:    arg.UserAssertion,
		Options:          arg.Options,
		ForceRemoteCheck: arg.ForceRemoteCheck,
	}
	ctx := engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewTrackEngine(&earg, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *TrackHandler) TrackWithToken(_ context.Context, arg keybase1.TrackWithTokenArg) error {
	earg := engine.TrackTokenArg{
		Token:   arg.TrackToken,
		Options: arg.Options,
	}
	ctx := engine.Context{
		IdentifyUI: h.ui.NewRemoteIdentifyUI(arg.SessionID, h.G()),
		SecretUI:   h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID:  arg.SessionID,
	}
	eng := engine.NewTrackToken(&earg, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *TrackHandler) DismissWithToken(_ context.Context, arg keybase1.DismissWithTokenArg) error {
	outcome, err := h.G().TrackCache.Get(arg.TrackToken)
	if err != nil {
		h.G().Log.Error("Failed to get track token", err)
		return err
	}
	if outcome.ResponsibleGregorItem == nil {
		h.G().Log.Debug("No responsible gregor item found for track token %s", arg.TrackToken)
		return nil
	}

	return h.G().GregorDismisser.DismissItem(outcome.ResponsibleGregorItem.Metadata().MsgID())
}

// Untrack creates an UntrackEngine and runs it.
func (h *TrackHandler) Untrack(_ context.Context, arg keybase1.UntrackArg) error {
	earg := engine.UntrackEngineArg{
		Username: arg.Username,
	}
	ctx := engine.Context{
		SecretUI:  h.ui.GetSecretUI(arg.SessionID, h.G()),
		SessionID: arg.SessionID,
	}
	eng := engine.NewUntrackEngine(&earg, h.G())
	return engine.RunEngine(eng, &ctx)
}

func (h *TrackHandler) CheckTracking(_ context.Context, sessionID int) error {
	if !h.G().RateLimits.GetPermission(libkb.CheckTrackingRateLimit, libkb.TrackingRateLimitSeconds*time.Second) {
		h.G().Log.Debug("Skipping CheckTracking due to rate limit.")
		return nil
	}
	return libkb.CheckTracking(h.G())
}

func (h *TrackHandler) FakeTrackingChanged(_ context.Context, arg keybase1.FakeTrackingChangedArg) error {
	user, err := libkb.LoadUser(libkb.LoadUserArg{
		Name: arg.Username,
	})
	if err != nil {
		return err
	}
	h.G().NotifyRouter.HandleTrackingChanged(user.GetUID(), user.GetName())
	return nil
}
