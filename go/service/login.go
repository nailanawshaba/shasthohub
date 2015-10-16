package service

import (
	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
)

type LoginHandler struct {
	libkb.Contextified
	*CancelHandler
	identifyUI  libkb.IdentifyUI
	locksmithUI libkb.LocksmithUI
}

func NewLoginHandler(xp rpc.Transporter, g *libkb.GlobalContext) *LoginHandler {
	return &LoginHandler{
		CancelHandler: NewCancelHandler(xp),
		Contextified:  libkb.NewContextified(g),
	}
}

func (h *LoginHandler) GetConfiguredAccounts(sessionID int) ([]keybase1.ConfiguredAccount, error) {
	return libkb.GetConfiguredAccounts(h.G())
}

func (h *LoginHandler) Logout(sessionID int) error {
	return h.G().Logout()
}

func (h *LoginHandler) Reset(sessionID int) error {
	eng := engine.NewResetEngine(h.G())
	ctx := engine.Context{}
	return engine.RunEngine(eng, &ctx)
}

func (h *LoginHandler) LoginWithPrompt(arg keybase1.LoginWithPromptArg) error {
	ctx := &engine.Context{
		LogUI:       h.getLogUI(arg.SessionID),
		LocksmithUI: h.getLocksmithUI(arg.SessionID),
		SecretUI:    h.getSecretUI(arg.SessionID),
		LoginUI:     h.getLoginUI(arg.SessionID),
		GPGUI:       h.getGPGUI(arg.SessionID),
	}
	eng := engine.NewLoginWithPromptEngine(arg.Username, h.G())

	return h.loginWithEngine(eng, ctx, arg.SessionID)
}

func (h *LoginHandler) LoginWithStoredSecret(arg keybase1.LoginWithStoredSecretArg) error {
	ctx := &engine.Context{
		LogUI:       h.getLogUI(arg.SessionID),
		LocksmithUI: h.getLocksmithUI(arg.SessionID),
		SecretUI:    h.getSecretUI(arg.SessionID),
		LoginUI:     h.getLoginUI(arg.SessionID),
		GPGUI:       h.getGPGUI(arg.SessionID),
	}
	loginEngine := engine.NewLoginWithStoredSecretEngine(arg.Username, h.G())
	return h.loginWithEngine(loginEngine, ctx, arg.SessionID)
}

func (h *LoginHandler) LoginWithPassphrase(arg keybase1.LoginWithPassphraseArg) error {
	ctx := &engine.Context{
		LogUI:       h.getLogUI(arg.SessionID),
		LocksmithUI: h.getLocksmithUI(arg.SessionID),
		SecretUI:    h.getSecretUI(arg.SessionID),
		LoginUI:     h.getLoginUI(arg.SessionID),
		GPGUI:       h.getGPGUI(arg.SessionID),
	}

	loginEngine := engine.NewLoginWithPassphraseEngine(arg.Username, arg.Passphrase, arg.StoreSecret, h.G())
	return h.loginWithEngine(loginEngine, ctx, arg.SessionID)
}

func (h *LoginHandler) ClearStoredSecret(arg keybase1.ClearStoredSecretArg) error {
	return libkb.ClearStoredSecret(libkb.NewNormalizedUsername(arg.Username))
}

func (h *LoginHandler) loginWithEngine(eng *engine.LoginEngine, ctx *engine.Context, sessionID int) error {
	h.setCanceler(sessionID, eng)
	defer h.removeCanceler(sessionID)
	err := engine.RunEngine(eng, ctx)
	if err != nil {
		if _, ok := err.(libkb.CanceledError); ok {
			h.G().Log.Debug("logging out due to login cancel")
			h.G().Logout()
		}
	}
	return err
}

func (h *LoginHandler) CancelLogin(sessionID int) error {
	c := h.canceler(sessionID)
	if c == nil {
		h.G().Log.Debug("CancelLogin called and there's no login engine for sessionID %d", sessionID)
		return libkb.LoginSessionNotFound{SessionID: sessionID}
	}
	return c.Cancel()
}

func (h *LoginHandler) PaperKey(sessionID int) error {
	ctx := &engine.Context{
		LogUI:    h.getLogUI(sessionID),
		LoginUI:  h.getLoginUI(sessionID),
		SecretUI: h.getSecretUI(sessionID),
	}
	eng := engine.NewPaperKey(h.G())
	return engine.RunEngine(eng, ctx)
}

func (h *LoginHandler) Unlock(sessionID int) error {
	ctx := &engine.Context{
		LogUI:    h.getLogUI(sessionID),
		SecretUI: h.getSecretUI(sessionID),
	}
	eng := engine.NewUnlock(h.G())
	return engine.RunEngine(eng, ctx)
}

type RemoteLocksmithUI struct {
	sessionID int
	uicli     keybase1.LocksmithUiClient
}

func NewRemoteLocksmithUI(sessionID int, c *rpc.Client) *RemoteLocksmithUI {
	return &RemoteLocksmithUI{
		sessionID: sessionID,
		uicli:     keybase1.LocksmithUiClient{Cli: c},
	}
}

func (r *RemoteLocksmithUI) PromptDeviceName(dummy int) (string, error) {
	return r.uicli.PromptDeviceName(r.sessionID)
}

func (r *RemoteLocksmithUI) DeviceNameTaken(arg keybase1.DeviceNameTakenArg) error {
	arg.SessionID = r.sessionID
	return r.uicli.DeviceNameTaken(arg)
}

func (r *RemoteLocksmithUI) SelectSigner(arg keybase1.SelectSignerArg) (keybase1.SelectSignerRes, error) {
	arg.SessionID = r.sessionID
	return r.uicli.SelectSigner(arg)
}

func (r *RemoteLocksmithUI) DeviceSignAttemptErr(arg keybase1.DeviceSignAttemptErrArg) error {
	arg.SessionID = r.sessionID
	return r.uicli.DeviceSignAttemptErr(arg)
}

func (r *RemoteLocksmithUI) DisplaySecretWords(arg keybase1.DisplaySecretWordsArg) error {
	arg.SessionID = r.sessionID
	return r.uicli.DisplaySecretWords(arg)
}

func (r *RemoteLocksmithUI) KexStatus(arg keybase1.KexStatusArg) error {
	arg.SessionID = r.sessionID
	return r.uicli.KexStatus(arg)
}

func (r *RemoteLocksmithUI) DisplayProvisionSuccess(arg keybase1.DisplayProvisionSuccessArg) error {
	arg.SessionID = r.sessionID
	return r.uicli.DisplayProvisionSuccess(arg)
}
