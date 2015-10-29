// +build windows

package client

import (
	"golang.org/x/sys/windows/svc"
)

func (s *CmdCtlStop) Run() (err error) {
	return controlService(svc.Stop, svc.Stopped)
}
