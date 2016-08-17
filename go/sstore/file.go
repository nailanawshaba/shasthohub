// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package sstore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/keybase/client/go/libkb"
)

var ErrSecretForUserNotFound = libkb.NotFoundError{Msg: "No secret found for user"}

type SecretStoreFile struct {
	dir string
}

func NewSecretStoreFile(dir string) *SecretStoreFile {
	return &SecretStoreFile{dir: dir}
}

func (s *SecretStoreFile) RetrieveSecret(username libkb.NormalizedUsername) ([]byte, error) {
	secret, err := ioutil.ReadFile(s.userpath(username))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSecretForUserNotFound
		}

		return nil, err
	}

	return secret, nil
}

func (s *SecretStoreFile) StoreSecret(username libkb.NormalizedUsername, secret []byte) error {
	f, err := ioutil.TempFile(s.dir, username.String())
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		// os.Fchmod not supported on windows
		if err := f.Chmod(0600); err != nil {
			return err
		}
	}
	if _, err := f.Write(secret); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	final := s.userpath(username)
	if err := os.Rename(f.Name(), final); err != nil {
		return err
	}
	return os.Chmod(final, 0600)
}

func (s *SecretStoreFile) ClearSecret(username libkb.NormalizedUsername) error {
	if err := os.Remove(s.userpath(username)); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return nil
}

func (s *SecretStoreFile) GetUsersWithStoredSecrets() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(s.dir, "*.ss"))
	if err != nil {
		return nil, err
	}
	users := make([]string, len(files))
	for i, f := range files {
		users[i] = stripExt(filepath.Base(f))
	}
	return users, nil
}

func (s *SecretStoreFile) GetApprovalPrompt() string {
	return "Remember login key"
}

func (s *SecretStoreFile) GetTerminalPrompt() string {
	return "Remember your login key?"
}

func (s *SecretStoreFile) userpath(username libkb.NormalizedUsername) string {
	return filepath.Join(s.dir, fmt.Sprintf("%s.ss", username))
}

func stripExt(path string) string {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[:i]
		}
	}
	return ""
}
