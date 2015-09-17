package bind

import (
	"bytes"
	"io/ioutil"

	"github.com/maxtaco/go-framed-msgpack-rpc/rpc2"
	"github.com/ugorji/go/codec"
)

type transportBytes struct {
	mh         codec.MsgpackHandle
	dispatcher rpc2.Dispatcher
	dec        *codec.Decoder
	client     Client
	ch         chan []byte
}

func newTransportBytes(mh codec.MsgpackHandle, data []byte, dispatcher rpc2.Dispatcher, client Client, ch chan []byte) *transportBytes {
	return &transportBytes{
		mh:         mh,
		dec:        codec.NewDecoderBytes(data, &mh),
		dispatcher: dispatcher,
		client:     client,
		ch:         ch,
	}
}

func (t transportBytes) RawWrite(b []byte) error {
	panic("Unsupported RawWrite")
}

func (t transportBytes) ReadByte() (byte, error) {
	panic("Unsupported ReadByte")
}

func (t transportBytes) Decode(i interface{}) error {
	return t.dec.Decode(i)
}

func (t transportBytes) Encode(i interface{}) error {
	b, err := encodeToBytes(t.mh, i)
	if err != nil {
		return err
	}
	t.client.Response(b)
	if t.ch != nil {
		t.ch <- b
	}
	return nil
}

func (t transportBytes) GetDispatcher() (rpc2.Dispatcher, error) {
	return t.dispatcher, nil
}

func (t transportBytes) ReadLock()   {}
func (t transportBytes) ReadUnlock() {}

type transportEncoder struct {
	mh         codec.MsgpackHandle
	dispatcher rpc2.Dispatcher
	client     Client
}

func newTransportEncoder(mh codec.MsgpackHandle, client Client) *transportEncoder {
	return &transportEncoder{
		mh:     mh,
		client: client,
	}
}

func (t transportEncoder) RawWrite(b []byte) error {
	panic("Unsupported RawWrite")
}

func (t transportEncoder) ReadByte() (byte, error) {
	panic("Unsupported ReadByte")
}

func (t transportEncoder) Decode(i interface{}) error {
	panic("Unsupported Decode")
	return nil
}

func (t transportEncoder) Encode(i interface{}) error {
	b, err := encodeToBytes(t.mh, i)
	if err != nil {
		return err
	}
	go t.client.Request(b)
	return nil
}

func (t transportEncoder) GetDispatcher() (rpc2.Dispatcher, error) {
	return t.dispatcher, nil
}

func (t transportEncoder) ReadLock()   {}
func (t transportEncoder) ReadUnlock() {}

func encodeToBytes(mh codec.MsgpackHandle, i interface{}) (v []byte, err error) {
	buf := new(bytes.Buffer)
	enc := codec.NewEncoder(buf, &mh)
	if err = enc.Encode(i); err != nil {
		return
	}
	v, _ = ioutil.ReadAll(buf)
	return
}
