package bind

import (
	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
)

type LoginWithPassphraseArg struct {
	SessionID   int
	Username    string
	Passphrase  string
	StoreSecret bool
}

type GetPaperKeyPassphraseArg struct {
	SessionID int
	Username  string
}

func NewLoginWithPassphraseArg() *LoginWithPassphraseArg {
	return &LoginWithPassphraseArg{}
}

func NewGetPaperKeyPassphraseArg() *GetPaperKeyPassphraseArg {
	return &GetPaperKeyPassphraseArg{}
}

type SecretUI interface {
	GetPaperKeyPassphrase(*GetPaperKeyPassphraseArg) (string, error)
}

type LocksmithUI interface {
	PromptDeviceName(int) (string, error)
}

type LogUI interface {
	Log(string)
}

func Init() {
	libkb.G.Init()
	usage := libkb.Usage{
		Config:    true,
		API:       true,
		KbKeyring: true,
	}
	libkb.G.ConfigureUsage(usage)
}

func LoginWithPassphrase(arg *LoginWithPassphraseArg, locksmith LocksmithUI, secret SecretUI, log LogUI) error {
	ctx := &engine.Context{
		LocksmithUI: locksmithUIWrapper{locksmith},
		SecretUI:    secretUIWrapper{secret},
		LogUI:       logUIWrapper{log},
		GPGUI:       gpgUIWrapper{},
	}

	loginEngine := engine.NewLoginWithPassphraseEngine(arg.Username, arg.Passphrase, arg.StoreSecret, libkb.G)
	return engine.RunEngine(loginEngine, ctx)
}
