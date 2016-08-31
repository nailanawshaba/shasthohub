package systests

import (
	"testing"

	"github.com/keybase/client/go/libkb"
)

type chatTester struct {
	t  *testing.T
	tc *libkb.TestContext
}

func newChatTester(t *testing.T) *chatTester {
	return &chatTester{
		t: t,
	}
}

func (ct *chatTester) setup() {
	ct.tc = setupTest(ct.t, "chat")
}

func (ct *chatTester) cleanup() {
	ct.tc.Cleanup()
}

func TestChatBasic(t *testing.T) {
	tester := newChatTester(t)
	tester.setup()
	defer tester.cleanup()

}
