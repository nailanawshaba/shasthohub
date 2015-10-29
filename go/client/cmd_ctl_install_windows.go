// +build windows

package client

import (
	"github.com/keybase/cli"
	"github.com/keybase/client/go/libcmdline"
	"github.com/keybase/client/go/libkb"
)

func NewCmdCtlInstall(cl *libcmdline.CommandLine, g *libkb.GlobalContext) cli.Command {
	return cli.Command{
		Name:  "install",
		Usage: "Install the background keybase service",
		Action: func(c *cli.Context) {
			cl.ChooseCommand(&CmdCtlInstall{libkb.NewContextified(g)}, "install", c)
			cl.SetForkCmd(libcmdline.NoFork)
		},
	}
}

type CmdCtlInstall struct {
	libkb.Contextified
}

func (s *CmdCtlInstall) ParseArgv(ctx *cli.Context) error {
	return nil
}

func (s *CmdCtlInstall) Run() (err error) {
	return installService()
}

func (s *CmdCtlInstall) GetUsage() libkb.Usage {
	return libkb.Usage{}
}
