// +build windows

// npipe_windows.go
package libkb

import (
	"net"

	"github.com/natefinch/npipe"
)

type SocketNamedPipe struct {
	Contextified
	pipename string
}

// user.Current() includes machine name on windows, but
// this is still only a local pipe because of the dot
// following the doulble backslashes.
// If the service ever runs under a different account than
// current user, this will have to be revisited.
func NewSocket(g *GlobalContext) (ret Socket, err error) {

	return SocketNamedPipe{
		Contextified: NewContextified(g),
		pipename:     `\\.\pipe\` + GetServiceName(),
	}, nil
}

func (s SocketNamedPipe) BindToSocket() (ret net.Listener, err error) {
	s.G().Log.Info("Binding to pipe:%s", s.pipename)
	return npipe.Listen(s.pipename)
}

func (s SocketNamedPipe) DialSocket() (ret net.Conn, err error) {
	return npipe.Dial(s.pipename)
}
