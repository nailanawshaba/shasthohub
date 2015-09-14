package bind

import (
	"fmt"

	keybase1 "github.com/keybase/client/protocol/go"
)

type secretUIWrapper struct {
	secretUI SecretUI
}

func unsupported() error {
	return fmt.Errorf("Unsupported")
}

func (w secretUIWrapper) GetSecret(pinentry keybase1.SecretEntryArg, terminal *keybase1.SecretEntryArg) (*keybase1.SecretEntryRes, error) {
	return nil, unsupported()
}

func (w secretUIWrapper) GetNewPassphrase(keybase1.GetNewPassphraseArg) (keybase1.GetNewPassphraseRes, error) {
	return keybase1.GetNewPassphraseRes{}, unsupported()
}

func (w secretUIWrapper) GetKeybasePassphrase(keybase1.GetKeybasePassphraseArg) (string, error) {
	return "", unsupported()
}

func (w secretUIWrapper) GetPaperKeyPassphrase(arg keybase1.GetPaperKeyPassphraseArg) (string, error) {
	a := GetPaperKeyPassphraseArg{SessionID: arg.SessionID, Username: arg.Username}
	return w.secretUI.GetPaperKeyPassphrase(&a)
}

type locksmithUIWrapper struct {
	locksmithUI LocksmithUI
}

func (w locksmithUIWrapper) PromptDeviceName(n int) (string, error) {
	return w.locksmithUI.PromptDeviceName(n)
}

func (w locksmithUIWrapper) SelectSigner(arg keybase1.SelectSignerArg) (keybase1.SelectSignerRes, error) {
	return keybase1.SelectSignerRes{
		Action: keybase1.SelectSignerAction_SIGN,
		Signer: &keybase1.DeviceSigner{
			Kind: keybase1.DeviceSignerKind_PAPER_BACKUP_KEY,
		},
	}, nil
}

func (w locksmithUIWrapper) DeviceSignAttemptErr(arg keybase1.DeviceSignAttemptErrArg) error {
	return unsupported()
}

func (w locksmithUIWrapper) DeviceNameTaken(keybase1.DeviceNameTakenArg) error { return unsupported() }

func (w locksmithUIWrapper) DisplaySecretWords(keybase1.DisplaySecretWordsArg) error {
	return unsupported()
}
func (w locksmithUIWrapper) KexStatus(keybase1.KexStatusArg) error { return unsupported() }

type logUIWrapper struct {
	logUI LogUI
}

func (w logUIWrapper) Debug(format string, args ...interface{}) {
	w.logUI.Log(fmt.Sprintf(format, args))
}

func (w logUIWrapper) Info(format string, args ...interface{}) {
	w.logUI.Log(fmt.Sprintf(format, args))
}

func (w logUIWrapper) Warning(format string, args ...interface{}) {
	w.logUI.Log(fmt.Sprintf(format, args))
}

func (w logUIWrapper) Notice(format string, args ...interface{}) {
	w.logUI.Log(fmt.Sprintf(format, args))
}

func (w logUIWrapper) Errorf(format string, args ...interface{}) {
	w.logUI.Log(fmt.Sprintf(format, args))
}

func (w logUIWrapper) Critical(format string, args ...interface{}) {
	w.logUI.Log(fmt.Sprintf(format, args))
}

type gpgUIWrapper struct{}

func (w gpgUIWrapper) WantToAddGPGKey(int) (bool, error) {
	return false, nil
}

func (w gpgUIWrapper) SelectKeyAndPushOption(keybase1.SelectKeyAndPushOptionArg) (keybase1.SelectKeyRes, error) {
	return keybase1.SelectKeyRes{}, unsupported()
}

func (w gpgUIWrapper) SelectKey(keybase1.SelectKeyArg) (string, error) {
	return "", unsupported()
}
