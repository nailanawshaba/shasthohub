// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package client

import (
	"fmt"
	"os"
	"syscall"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/service"
)

// ForkServer forks a new background Keybase service, and waits until it's
// pingable. It will only do something useful on Unixes; it won't work on
// Windows (probably?). Returns an error if anything bad happens; otherwise,
// assume that the server was successfully started up.
func ForkServer(cl libkb.CommandLine, g *libkb.GlobalContext) error {
	srv := service.NewService(true /* isDaemon */, g)

	// If we try to get an exclusive lock and succeed, it means we
	// need to relaunch the daemon since it's dead
	g.Log.Debug("Getting flock")
	err := srv.GetExclusiveLockWithoutAutoUnlock()
	if err == nil {
		g.Log.Debug("Flocked! Server must have died")
		srv.ReleaseLock()
		err = spawnServer(cl)
		if err != nil {
			g.Log.Errorf("Error in spawning server process: %s", err)
			return err
		}
		err = pingLoop()
		if err != nil {
			g.Log.Errorf("Ping failure after server fork: %s", err)
			return err
		}
	} else {
		g.Log.Debug("The server is still up")
		err = nil
	}

	return err
}

func spawnServer(cl libkb.CommandLine) (err error) {

	var files []uintptr
	var cmd string
	var args []string
	var devnull, log *os.File
	var pid int

	defer func() {
		if err != nil {
			if devnull != nil {
				devnull.Close()
			}
			if log != nil {
				log.Close()
			}
		}
	}()

	if devnull, err = os.OpenFile("/dev/null", os.O_RDONLY, 0); err != nil {
		return
	}
	files = append(files, devnull.Fd())

	if G.Env.GetSplitLogOutput() {
		files = append(files, uintptr(1), uintptr(2))
	} else {
		if _, log, err = libkb.OpenLogFile(); err != nil {
			return
		}
		files = append(files, log.Fd(), log.Fd())
	}

	attr := syscall.ProcAttr{
		Env:   os.Environ(),
		Sys:   &syscall.SysProcAttr{Setsid: true},
		Files: files,
	}

	cmd, args, err = makeServerCommandLine(cl)
	if err != nil {
		return err
	}

	pid, err = syscall.ForkExec(cmd, args, &attr)
	if err != nil {
		err = fmt.Errorf("Error in ForkExec: %s", err)
	} else {
		G.Log.Info("Forking background server with pid=%d", pid)
	}
	return err
}
