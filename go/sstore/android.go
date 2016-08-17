// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

// +build android

package sstore

import "github.com/keybase/client/go/libkb"

// TODO: Make this implementation use GetStoredSecretServiceName(), as
// otherwise tests will clobber each other's passwords. See
// https://keybase.atlassian.net/browse/CORE-1934 .

func NewSecretStorer(g *libkb.GlobalContext) SecretStorer {
	externalKeyStore := getGlobalExternalKeyStore()
	if externalKeyStore == nil {
		return nil
	}
	return secretStoreAccountName{externalKeyStore, g}
}
