// +build !windows

package client

import (
	"github.com/keybase/cli"
	"github.com/keybase/client/go/libcmdline"
	"golang.org/x/net/context"
)

func (s *CmdCtlStop) Run() (err error) {
	cli, err := GetCtlClient(s.G())
	if err != nil {
		return err
	}
	return cli.Stop(context.TODO(), 0)
}
