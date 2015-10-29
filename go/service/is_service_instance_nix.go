// +build !windows

package service

import (
	"errors"

	"github.com/keybase/client/go/libcmdline"
	"github.com/keybase/client/go/libkb"
)

func CheckRunAsService(cl *libcmdline.CommandLine, cmd libkb.Command) (error, bool) {

	if cl.IsService() {
		return cmd.Run()
	}
	return nil, false
}
