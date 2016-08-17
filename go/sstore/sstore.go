package sstore

import (
	"sync"

	"github.com/keybase/client/go/libkb"
)

type SecretStorer interface {
	RetrieveSecret(username libkb.NormalizedUsername) ([]byte, error)
	StoreSecret(username libkb.NormalizedUsername, secret []byte) error
	ClearSecret(username libkb.NormalizedUsername) error
	GetUsersWithStoredSecrets() ([]string, error)
	GetApprovalPrompt() string
	GetTerminalPrompt() string
}

type SecretStorage struct {
	storer SecretStorer
	sync.Mutex
}

func NewSecretStorage(g *libkb.GlobalContext) *SecretStorage {
	return &SecretStorage{storer: NewSecretStorer(g)}
}

func (s *SecretStorage) RetrieveSecret(username libkb.NormalizedUsername) ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	return s.storer.RetrieveSecret(username)
}

func (s *SecretStorage) StoreSecret(username libkb.NormalizedUsername, secret []byte) error {
	s.Lock()
	defer s.Unlock()
	return s.storer.StoreSecret(username, secret)
}

func (s *SecretStorage) ClearSecret(username libkb.NormalizedUsername) error {
	s.Lock()
	defer s.Unlock()
	return s.storer.ClearSecret(username)
}

func (s *SecretStorage) GetUsersWithStoredSecrets() ([]string, error) {
	s.Lock()
	defer s.Unlock()
	return s.storer.GetUsersWithStoredSecrets()
}

func (s *SecretStorage) GetApprovalPrompt() string {
	s.Lock()
	defer s.Unlock()
	return s.storer.GetApprovalPrompt()
}

func (s *SecretStorage) GetTerminalPrompt() string {
	s.Lock()
	defer s.Unlock()
	return s.storer.GetTerminalPrompt()
}
