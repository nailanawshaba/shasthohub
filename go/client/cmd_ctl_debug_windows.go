// +build windows

package client

import (
	"github.com/keybase/cli"
	"github.com/keybase/client/go/libcmdline"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/service"
)

func NewCmdCtlDebug(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	return cli.Command{
		Name:  "debug",
		Usage: "Debug the background keybase service",
		Action: func(c *cli.Context) {
			cl.ChooseCommand(&CmdCtlDebug{libkb.NewContextified(g)}, "debug", c)
			cl.SetForkCmd(libcmdline.NoFork)
			cl.SetService()
		},
	}
}

type CmdCtlDebug struct {
	libkb.Contextified
}

func (s *CmdCtlDebug) ParseArgv(ctx *cli.Context) error {
	return nil
}

func (s *CmdCtlDebug) Run() (err error) {
	service.RunWinService(true)
	return nil
}

func (s *CmdCtlDebug) GetUsage() libkb.Usage {
	return libkb.Usage{}
}
