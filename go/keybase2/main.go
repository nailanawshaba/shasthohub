package main

import (
	"os"

	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/libkb"
	"gopkg.in/alecthomas/kingpin.v2"
)

var G = libkb.G

func main() {
	g := G
	g.Init()

	err := mainInner(g)
	e2 := g.Shutdown()
	if err == nil {
		err = e2
	}
	if err != nil {
		g.Log.Error(err.Error())
		os.Exit(2)
	}
}

func mainInner(g *libkb.GlobalContext) error {
	app := kingpin.New("keybase", "Keybase command line client.")

	client.RegisterCmdID(app)
	client.RegisterCmdLogin(app)
	client.RegisterCmdListTrackers(app)

	_, err := app.Parse(os.Args[1:])
	return err
}
