package client

import (
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/protocol/go"
	"github.com/maxtaco/go-framed-msgpack-rpc/rpc2"
)

// Keep this around to simplify things
var G = libkb.G
var GlobUI *UI

func InitUI() {
	GlobUI = &UI{}
	G.SetUI(GlobUI)
}

type Usage interface {
	GetUsage() libkb.Usage
}

func InitClient(u Usage) error {
	InitUI()
	err := G.ConfigureForUsage(u.GetUsage())
	if err != nil {
		return err
	}
	return RegisterGlobalLogUI(G)
}

func RegisterGlobalLogUI(g *libkb.GlobalContext) error {
	protocols := []rpc2.Protocol{NewLogUIProtocol()}
	if err := RegisterProtocols(protocols); err != nil {
		return err
	}
	// Send our current debugging state, so that the server can avoid
	// sending us verbose logs when we don't want to read them.
	logLevel := keybase1.LogLevel_INFO
	if g.Env.GetDebug() {
		logLevel = keybase1.LogLevel_DEBUG
	}
	ctlClient, err := GetCtlClient()
	if err != nil {
		return err
	}
	ctlClient.SetLogLevel(logLevel)
	return nil
}
