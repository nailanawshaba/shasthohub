// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build !darwin,!android

package libkb

// These can be set to mocks by a test
var NewTestSecretStoreFunc func(context SecretStoreContext, accountName NormalizedUsername) SecretStore
var GetUsersWithStoredSecretsFunc func(c SecretStoreContext) ([]string, error)

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
	if GetUsersWithStoredSecretsFunc != nil {
		return GetUsersWithStoredSecretsFunc(c)
	}
	return nil, nil
}

func GetTerminalPrompt() string {
	// TODO: Come up with specific prompts for other platforms.
	return "Store your key in the local secret store?"
}
