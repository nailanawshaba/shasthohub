package bind

import (
	"fmt"

	"github.com/maxtaco/go-framed-msgpack-rpc/rpc2"
)

type log struct {
	output func(string)
}

func (l log) log(format string, args ...interface{}) {
	l.output(fmt.Sprintf(format, args))
}

func (l log) TransportStart() {
	l.log("Start\n")
}

func (l log) TransportError(err error) {
	l.log("Error: %v\n", err)
}

func (l log) ServerCall(n int, s string, e error, i interface{}) {
	l.log("ServerCall: %#v\n", []interface{}{n, s, e, i})
}

func (l log) ServerReply(n int, s string, e error, i interface{}) {
	l.log("ServerReply: %#v\n", []interface{}{n, s, e, i})
}

func (l log) ClientCall(n int, s string, i interface{}) {
	l.log("ClientCall: %#v\n", []interface{}{n, s, i})
}

func (l log) ClientReply(n int, s string, e error, i interface{}) {
	l.log("ClientReply: %#v\n", []interface{}{n, s, e, i})
}

func (l log) StartProfiler(format string, args ...interface{}) rpc2.Profiler {
	return nil
}

func (l log) UnexpectedReply(n int) {
	l.log("Unexpected reply: %d\n", n)
}

func (l log) Warning(format string, args ...interface{}) {
	l.log("Warning: %s\n", fmt.Sprintf(format, args))
}
