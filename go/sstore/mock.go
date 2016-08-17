// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sstore

import (
	"errors"

	"github.com/keybase/client/go/libkb"
)

// Used by tests that want to mock out the secret store.
type TestSecretStore struct {
	context            SecretStoreContext
	secretStoreNoneMap map[libkb.NormalizedUsername][]byte
	libkb.Contextified
}

func NewTestSecretStorer(c SecretStoreContext, g *libkb.GlobalContext) SecretStorer {
	ret := TestSecretStore{context: c, secretStoreNoneMap: map[libkb.NormalizedUsername][]byte{}, Contextified: libkb.NewContextified(g)}
	return ret
}

func (t TestSecretStore) GetUsersWithStoredSecrets() (ret []string, err error) {
	for name := range t.secretStoreNoneMap {
		ret = append(ret, string(name))
	}
	return
}

func (t TestSecretStore) GetTerminalPrompt() string {
	return "Store your key in the local secret store?"
}

func (t TestSecretStore) GetApprovalPrompt() string {
	return "Store your key in the local secret store?"
}

func (t TestSecretStore) GetAllUserNames() (libkb.NormalizedUsername, []libkb.NormalizedUsername, error) {
	return t.context.GetAllUserNames()
}

func (t TestSecretStore) RetrieveSecret(accountName libkb.NormalizedUsername) (ret []byte, err error) {
	ret, ok := t.secretStoreNoneMap[accountName]

	t.G().Log.Debug("| TestSecretStore::RetrieveSecret(%d)", len(ret))

	if !ok {
		return nil, errors.New("No secret to retrieve")
	}

	return
}

func (t TestSecretStore) StoreSecret(accountName libkb.NormalizedUsername, secret []byte) error {
	t.G().Log.Debug("| TestSecretStore::StoreSecret(%d)", len(secret))

	t.secretStoreNoneMap[accountName] = secret
	return nil
}

func (t TestSecretStore) ClearSecret(accountName libkb.NormalizedUsername) error {
	t.G().Log.Debug("| TestSecretStore::ClearSecret()")

	delete(t.secretStoreNoneMap, accountName)
	return nil
}
