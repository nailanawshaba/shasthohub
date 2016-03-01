// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !darwin,!android

package libkb

var NewTestSecretStoreFunc func(context SecretStoreContext, accountName NormalizedUsername) SecretStore

func NewSecretStore(c SecretStoreContext, username NormalizedUsername) SecretStore {
	if NewTestSecretStoreFunc != nil {
		return NewTestSecretStoreFunc(c, username)
	}
	return nil
}

func HasSecretStore() bool {
	return NewTestSecretStoreFunc != nil
}

func GetUsersWithStoredSecrets(c SecretStoreContext) ([]string, error) {
	if NewTestSecretStoreFunc != nil {
		return GetTestUsersWithStoredSecrets(c)
	}
	return nil, nil
}

func GetTerminalPrompt() string {
	// TODO: Come up with specific prompts for other platforms.
	return "Store your key in the local secret store?"
}
