package bind

import (
	"testing"

	"github.com/ugorji/go/codec"
)

func makeRequest(messageID int, method string, args map[string]interface{}) []byte {
	mh := codec.MsgpackHandle{WriteExt: true}
	message := make([]interface{}, 4)
	message[0] = 0
	message[1] = messageID
	message[2] = method
	message[3] = []interface{}{args}
	b, _ := encodeToBytes(mh, message)
	return b
}

func makeResponse(messageID int, i interface{}) []byte {
	mh := codec.MsgpackHandle{WriteExt: true}
	message := make([]interface{}, 4)
	message[0] = 1
	message[1] = messageID
	message[2] = nil
	message[3] = i
	b, _ := encodeToBytes(mh, message)
	return b
}

type TestClient struct {
	t   *testing.T
	b   []byte
	svc *Service
}

func (tc *TestClient) Request(b []byte) {
	tc.t.Logf("Client got request: %v\n", b)
	tc.t.Logf("Sending reply: %v\n", tc.b)
	tc.svc.Send(tc.b)
}

func (tc *TestClient) Response(b []byte) {
	tc.t.Logf("Response: %v\n", b)
}

func (tc *TestClient) Log(s string) {
	tc.t.Logf(s)
}

func (tc *TestClient) Error(err error) {
	tc.t.Logf("Error: %v\n", err)
}

func TestService(t *testing.T) {
	args := make(map[string]interface{})
	args["sessionID"] = 2
	args["name"] = "testArg"

	breq := makeRequest(100, "keybase.1.test.test", args)

	tc := &TestClient{b: makeResponse(0, "testing"), t: t}
	svc := NewService(tc)
	tc.svc = svc
	svc.SendSync(breq)
}
