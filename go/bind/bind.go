// To build on iOS: `gomobile bind -target=ios`

package bind

import (
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/protocol"
	"github.com/keybase/client/go/service"
	"github.com/maxtaco/go-framed-msgpack-rpc/rpc2"
	"github.com/ugorji/go/codec"
)

type Service struct {
	mh         *codec.MsgpackHandle
	transport  *transportEncoder
	dispatcher rpc2.Dispatcher
}

type Client interface {
	Request([]byte)
	Response([]byte)
	Log(string)
}

func Init(homeDir string) {
	libkb.G.Init()
	usage := libkb.Usage{
		Config:    true,
		API:       true,
		KbKeyring: true,
	}
	config := libkb.AppConfig{HomeDir: homeDir, RunMode: libkb.DevelRunMode, Debug: true, LocalRPCDebug: "Acsvip"}
	libkb.G.Configure(config, usage)
}

func NewService(client Client) *Service {
	mh := codec.MsgpackHandle{WriteExt: true}
	lg := log{output: func(s string) { client.Log(s) }}

	transport := newTransportEncoder(mh, client)
	dispatcher := rpc2.NewDispatch(transport, lg, nil)
	transport.dispatcher = dispatcher

	// Add protocols
	protocols := []rpc2.Protocol{
		keybase1.LoginProtocol(service.NewLoginHandler(transport)),
		keybase1.TestProtocol(service.NewTestHandler(transport)),
	}
	for _, proto := range protocols {
		if err := dispatcher.RegisterProtocol(proto); err != nil {
			panic(err)
			return nil
		}
	}

	return &Service{mh: &mh, transport: transport, dispatcher: dispatcher}
}

func (svc *Service) send(b []byte, ch chan []byte) {
	nb := int(b[0])
	nFields := (nb - 0x90)
	b = b[1:]

	transportBytes := newTransportBytes(*svc.mh, b, svc.dispatcher, svc.transport.client, ch)
	msg := rpc2.NewMessage(transportBytes, nFields)
	err := svc.dispatcher.Dispatch(&msg)
	if err != nil {
		berr, err := encodeToBytes(transportBytes.mh, libkb.WrapError(err))
		if err != nil {
			panic(err)
		}
		ch <- berr
	}
}

func (s *Service) Send(b []byte) {
	s.send(b, nil)
}

func (s *Service) SendSync(b []byte) []byte {
	ch := make(chan []byte)
	s.send(b, ch)
	return <-ch
}
