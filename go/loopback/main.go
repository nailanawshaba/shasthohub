package loopback

import (
	"github.com/keybase/client/go/libkb"
	"net"
	"sync"
)

var mu sync.Mutex
var started bool
var con net.Conn

//
// A VERY ROUGH SKETCH OF HOW WE CAN ONLY EXPORT Read and Write
// FROM GO AND GET ACCESS TO ALL GO LIBRARIES
//
//  There's still a fair amount of work to do here:
//    -- don't ignore the various errors
//    -- configure the global context with more care
//    -- test it ... just a bit...
//
// Some more references here on how to compile to an .so:
//
//   https://blog.filippo.io/building-python-modules-with-go-1-5/
//

type dummyCmd struct{}

func (d dummyCmd) GetUsage() libkb.Usage { return libkb.Usage{} }

func start() {
	if !started {
		g := libkb.NewGlobalContext()
		g.Init()
		g.ConfigureAll(libkb.NullConfiguration{}, dummyCmd{})
		g.MakeLoopbackServer()
		con, _, _ = g.GetSocket()
		started = true
	}
}

func Write(b []byte) {
	mu.Lock()
	start()
	mu.Unlock()

	con.Write(b)
}

func Read(arg []byte) {
	mu.Lock()
	start()
	mu.Unlock()
	con.Read(arg)
}
