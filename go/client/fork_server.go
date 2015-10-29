package client

import (
	"os"
	"os/exec"
	"time"

	"github.com/keybase/cli"
	"github.com/keybase/client/go/libkb"
)

// GetExtraFlags gets the extra fork-related flags for this platform
func GetExtraFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "no-auto-fork, F",
			Usage: "Disable auto-fork of background service.",
		},
		cli.BoolFlag{
			Name:  "auto-fork",
			Usage: "Enable auto-fork of background service.",
		},
	}
}

func pingLoop() error {
	var err error
	for i := 0; i < 10; i++ {
		_, _, err = G.GetSocket(true)
		if err == nil {
			G.Log.Debug("Connected (%d)", i)
			return nil
		}
		G.Log.Debug("Failed to connect to socket (%d): %s", i, err)
		err = nil
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}

func makeServerCommandLine(cl libkb.CommandLine) (arg0 string, args []string, err error) {
	// ForkExec requires an absolute path to the binary. LookPath() gets this
	// for us, or correctly leaves arg0 alone if it's already a path.
	arg0, err = exec.LookPath(os.Args[0])
	if err != nil {
		return
	}

	// Fixme: This isn't ideal, it would be better to specify when the args
	// are defined if they should be reexported to the server, and if so, then
	// we should automate the reconstruction of the argument vector.  Let's do
	// this when we yank out keybase/cli
	bools := []string{
		"debug",
		"api-dump-unsafe",
		"plain-logging",
	}

	strings := []string{
		"home",
		"server",
		"config",
		"session",
		"proxy",
		"username",
		"gpg-home",
		"gpg",
		"secret-keyring",
		"pid-file",
		"socket-file",
		"gpg-options",
		"local-rpc-debug-unsafe",
		"run-mode",
		"timers",
	}
	args = append(args, arg0)

	for _, b := range bools {
		if isSet, isTrue := cl.GetBool(b, true); isSet && isTrue {
			args = append(args, "--"+b)
		}
	}

	for _, s := range strings {
		if v := cl.GetGString(s); len(v) > 0 {
			args = append(args, "--"+s, v)
		}
	}

	args = append(args, "service")

	var chdir string
	chdir, err = G.Env.GetServiceSpawnDir()
	if err != nil {
		return
	}

	G.Log.Debug("| Setting run directory for keybase service to %s", chdir)
	args = append(args, "--chdir", chdir)

	G.Log.Debug("| Made server args: %s %v", arg0, args)

	return
}
