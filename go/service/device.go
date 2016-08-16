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

// DeviceHandler implements keybase1.DeviceInterface
type DeviceHandler struct {
	libkb.Contextified
	ui DeviceUI
}

var _ keybase1.DeviceInterface = (*DeviceHandler)(nil)

// DeviceRPCHandler is the RPC handler for the device interface
type DeviceRPCHandler struct {
	*BaseHandler
	*DeviceHandler
}

// DeviceUI resolves UI for login requests
type DeviceUI interface {
	GetSecretUI(sessionID int, g *libkb.GlobalContext) libkb.SecretUI
	GetLogUI(sessionID int) libkb.LogUI
	GetProvisionUI(sessionID int) libkb.ProvisionUI
}

// NewDeviceHandler creates a DeviceHandler
func NewDeviceHandler(g *libkb.GlobalContext, ui DeviceUI) *DeviceHandler {
	return &DeviceHandler{
		Contextified: libkb.NewContextified(g),
		ui:           ui,
	}
}

// NewDeviceRPCHandler creates a DeviceHandler for the xp transport.
func NewDeviceRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *DeviceRPCHandler {
	handler := NewBaseHandler(xp)
	return &DeviceRPCHandler{
		BaseHandler:   handler,
		DeviceHandler: NewDeviceHandler(g, handler),
	}
}

// DeviceList returns a list of all the devices for a user.
func (h *DeviceHandler) DeviceList(_ context.Context, sessionID int) ([]keybase1.Device, error) {
	ctx := &engine.Context{
		LogUI:     h.ui.GetLogUI(sessionID),
		SessionID: sessionID,
	}
	eng := engine.NewDevList(h.G())
	if err := engine.RunEngine(eng, ctx); err != nil {
		return nil, err
	}
	return eng.List(), nil
}

// DeviceHistoryList returns a list of all the devices for a user,
// with detailed history and provisioner, revoker information.
func (h *DeviceHandler) DeviceHistoryList(nctx context.Context, sessionID int) ([]keybase1.DeviceDetail, error) {
	ctx := &engine.Context{
		LogUI:      h.ui.GetLogUI(sessionID),
		NetContext: nctx,
		SessionID:  sessionID,
	}
	eng := engine.NewDeviceHistorySelf(h.G())
	if err := engine.RunEngine(eng, ctx); err != nil {
		return nil, err
	}
	return eng.Devices(), nil
}

// DeviceAdd starts the kex2 device provisioning on the
// provisioner (device X/C1)
func (h *DeviceHandler) DeviceAdd(c context.Context, sessionID int) error {
	ctx := &engine.Context{
		ProvisionUI: h.ui.GetProvisionUI(sessionID),
		SecretUI:    h.ui.GetSecretUI(sessionID, h.G()),
		SessionID:   sessionID,
		NetContext:  c,
	}
	eng := engine.NewDeviceAdd(h.G())
	return engine.RunEngine(eng, ctx)
}

// CheckDeviceNameFormat verifies that the device name has a valid
// format.
func (h *DeviceHandler) CheckDeviceNameFormat(_ context.Context, arg keybase1.CheckDeviceNameFormatArg) (bool, error) {
	ok := libkb.CheckDeviceName.F(arg.Name)
	if ok {
		return ok, nil
	}
	return false, errors.New(libkb.CheckDeviceName.Hint)
}
