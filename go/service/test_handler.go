// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package service

import (
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	"golang.org/x/net/context"
)

type TestRPCHandler struct {
	*BaseHandler
	libkb.Contextified
}

func NewTestRPCHandler(xp rpc.Transporter, g *libkb.GlobalContext) *TestRPCHandler {
	return &TestRPCHandler{
		BaseHandler:  NewBaseHandler(xp),
		Contextified: libkb.NewContextified(g),
	}
}

func (t TestRPCHandler) Test(ctx context.Context, arg keybase1.TestArg) (test keybase1.Test, err error) {
	client := t.rpcClient()
	cbArg := keybase1.TestCallbackArg{Name: arg.Name, SessionID: arg.SessionID}
	var cbReply string
	err = client.Call(ctx, "keybase.1.test.testCallback", []interface{}{cbArg}, &cbReply)
	if err != nil {
		return
	}

	test.Reply = cbReply
	return
}

func (t TestRPCHandler) TestCallback(_ context.Context, arg keybase1.TestCallbackArg) (s string, err error) {
	return
}

func (t TestRPCHandler) Panic(_ context.Context, message string) error {
	t.G().Log.Info("Received panic() RPC")
	go func() {
		panic(message)
	}()
	return nil
}
