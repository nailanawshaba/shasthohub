// +build windows

package client

import (
	"github.com/keybase/cli"
	"github.com/keybase/client/go/libcmdline"
	"github.com/keybase/client/go/libkb"
)

func NewCmdCtlRemove(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	return cli.Command{
		Name:  "remove",
		Usage: "Remove the background keybase service",
		Action: func(c *cli.Context) {
			cl.ChooseCommand(&CmdCtlRemove{libkb.NewContextified(g)}, "remove", c)
			cl.SetForkCmd(libcmdline.NoFork)
		},
	}
}

type CmdCtlRemove struct {
	libkb.Contextified
}

func (s *CmdCtlRemove) ParseArgv(ctx *cli.Context) error {
	return nil
}

func (s *CmdCtlRemove) Run() (err error) {
	return removeService()
}

func (s *CmdCtlRemove) GetUsage() libkb.Usage {
	return libkb.Usage{}
}
