package libkb

import (
	"fmt"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	"net"
)

type SocketInfo interface {
	PrepSocket() error
	ToStringPair() (string, string)
}

type SocketInfoUnix struct {
	file string
}

type SocketInfoTCP struct {
	port int
}

func (s SocketInfoUnix) PrepSocket() error {
	return MakeParentDirs(s.file)
}

func (s SocketInfoUnix) ToStringPair() (string, string) {
	return "unix", s.file
}

func (s SocketInfoTCP) PrepSocket() error {
	return nil
}

func (s SocketInfoTCP) ToStringPair() (string, string) {
	return "tcp", fmt.Sprintf("127.0.0.1:%d", s.port)
}

func BindToSocket(info SocketInfo) (ret net.Listener, err error) {
	if err = info.PrepSocket(); err != nil {
		return
	}
	l, a := info.ToStringPair()
	G.Log.Info("Binding to %s:%s", l, a)
	ret, err = net.Listen(l, a)
	return ret, err
}

func DialSocket(info SocketInfo) (ret net.Conn, err error) {
	return net.Dial(info.ToStringPair())
}

type SocketWrapper struct {
	conn net.Conn
	xp   rpc.Transporter
	err  error
}

func (g *GlobalContext) MakeLoopbackServer() (l net.Listener, err error) {
	g.socketWrapperMu.Lock()
	g.LoopbackListener = NewLoopbackListener()
	l = g.LoopbackListener
	g.socketWrapperMu.Unlock()
	return
}

func (g *GlobalContext) BindToSocket() (net.Listener, error) {
	return BindToSocket(g.SocketInfo)
}

func (g *GlobalContext) GetSocket(clearError bool) (net.Conn, rpc.Transporter, error) {

	// Protect all global socket wrapper manipulation with a
	// lock to prevent race conditions.
	g.socketWrapperMu.Lock()
	defer g.socketWrapperMu.Unlock()

	needWrapper := false
	if g.SocketWrapper == nil {
		needWrapper = true
	} else if g.SocketWrapper.xp != nil && !g.SocketWrapper.xp.IsConnected() {
		// need reconnect
		G.Log.Info("rpc transport disconnected, reconnecting...")
		needWrapper = true
	}

	if needWrapper {
		sw := SocketWrapper{}
		if g.LoopbackListener != nil {
			sw.conn, sw.err = g.LoopbackListener.Dial()
		} else if g.SocketInfo == nil {
			sw.err = fmt.Errorf("Cannot get socket in standalone mode")
		} else {
			sw.conn, sw.err = DialSocket(g.SocketInfo)
		}
		if sw.err == nil {
			sw.xp = rpc.NewTransport(sw.conn, NewRPCLogFactory(), WrapError)
		}
		g.SocketWrapper = &sw
	}

	sw := g.SocketWrapper
	if sw.err != nil && clearError {
		g.SocketWrapper = nil
	}

	return sw.conn, sw.xp, sw.err
}
