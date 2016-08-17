// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sstore

import (
	"sync"

	"github.com/keybase/client/go/libkb"
	"github.com/keybase/go-codec/codec"
)

// TODO: Make this implementation use GetStoredSecretServiceName(), as
// otherwise tests will clobber each other's passwords. See
// https://keybase.atlassian.net/browse/CORE-1934 .

// ExternalKeyStore is the interface for the actual (external) keystore.
type ExternalKeyStore interface {
	RetrieveSecret(serviceName string, key string) ([]byte, error)
	StoreSecret(serviceName string, key string, secret []byte) error
	ClearSecret(serviceName string, key string) error
	GetUsersWithStoredSecretsMsgPack(serviceName string) ([]byte, error)
	SetupKeyStore(serviceName string, key string) error
}

// externalKeyStore is the reference to some external key store
var externalKeyStore ExternalKeyStore

var externalKeyStoreMu sync.Mutex

func (s secretStoreAccountName) serviceName() string {
	return s.context.GetStoredSecretServiceName()
}

// SetGlobalExternalKeyStore is called by Android to register Android's KeyStore with Go
func SetGlobalExternalKeyStore(s ExternalKeyStore) {
	externalKeyStoreMu.Lock()
	defer externalKeyStoreMu.Unlock()
	externalKeyStore = s
}

func getGlobalExternalKeyStore() ExternalKeyStore {
	externalKeyStoreMu.Lock()
	defer externalKeyStoreMu.Unlock()
	return externalKeyStore
}

type secretStoreAccountName struct {
	externalKeyStore ExternalKeyStore
	context          SecretStoreContext
}

func (s secretStoreAccountName) StoreSecret(username libkb.NormalizedUsername, secret []byte) (err error) {
	s.externalKeyStore.SetupKeyStore(s.serviceName(), string(username))
	return s.externalKeyStore.StoreSecret(s.serviceName(), string(username), secret)
}

func (s secretStoreAccountName) RetrieveSecret(username libkb.NormalizedUsername) ([]byte, error) {
	s.externalKeyStore.SetupKeyStore(s.serviceName(), string(username))
	return s.externalKeyStore.RetrieveSecret(s.serviceName(), string(username))
}

func (s secretStoreAccountName) ClearSecret(username libkb.NormalizedUsername) (err error) {
	return s.externalKeyStore.ClearSecret(s.serviceName(), string(username))
}

func (s secretStoreAccountName) GetUsersWithStoredSecrets() ([]string, error) {
	if s.externalKeyStore == nil {
		return nil, nil
	}
	usersMsgPack, err := s.externalKeyStore.GetUsersWithStoredSecretsMsgPack(s.serviceName())
	if err != nil {
		return nil, err
	}
	var users []string
	ch := codecHandle()
	err = libkb.MsgpackDecodeAll(usersMsgPack, ch, &users)
	return users, err
}

func (s secretStoreAccountName) GetTerminalPrompt() string {
	return "Store secret in Android's KeyStore?"
}

func (s secretStoreAccountName) GetApprovalPrompt() string {
	return "Store secret in Android's KeyStore?"
}

func codecHandle() *codec.MsgpackHandle {
	var mh codec.MsgpackHandle
	mh.WriteExt = true
	return &mh
}
