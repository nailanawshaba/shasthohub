// +build windows

package service

import (
	"errors"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"

	"github.com/keybase/client/go/libcmdline"
	"github.com/keybase/client/go/libkb"
)

// Keep this around to simplify things
var G = libkb.G

var cmd libcmdline.Command

func IsServiceInstance() bool {
	isIntSess, err := svc.IsAnInteractiveSession()
	return err == nil && isIntSess
}

// In case this instance was started non-interactively, we still want it to
// be an official Windows service
func CheckRunAsService(cl *libcmdline.CommandLine, cmd libkb.Command) (error, bool) {
	isIntSess, _ := svc.IsAnInteractiveSession()

	if !isIntSess {
		RunWinService(false)
		return nil, true
	}
	return nil, false
}

type winservice struct{}

func (m *winservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	// do process startup
	ch := make(chan error)
	g := G
	// If we are in service debug mode, we came in through
	// keybase.main() and G is already initialized
	if g.Env == nil {
		g.Init()
		defer g.Shutdown()
	}

	s := NewService(true, g)

	if s == nil {
		ch <- errors.New("Error creating service object\n")
		return
	}

	go func() {
		ch <- s.Run()
	}()

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case <-ch:
			// stop service
			s.Stop()
			break loop
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	g.Shutdown()
	return
}

func RunWinService(isDebug bool) {
	var err error

	//	elog.Info(1, fmt.Sprintf("starting keybase service"))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run("keybase", &winservice{})
	if err != nil {
		//		elog.Error(1, fmt.Sprintf("keyhbase service failed: %v", err))
		return
	}
	//	elog.Info(1, fmt.Sprintf("keybase service stopped"))
}
