// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	keybase1 "github.com/keybase/client/go/protocol"
	"strings"
)

type AlgoType int

const (
	PGPFingerprintLen = 20
)

const (
	KIDPGPBase    AlgoType = 0x00
	KIDPGPRsa              = 0x1
	KIDPGPElgamal          = 0x10
	KIDPGPDsa              = 0x11
	KIDPGPEcdh             = 0x12
	KIDPGPEcdsa            = 0x13
	KIDNaclEddsa           = 0x20
	KIDNaclDH              = 0x21
)

type PGPFingerprint [PGPFingerprintLen]byte

type GenericKey interface {
	GetKID() keybase1.KID
	GetBinaryKID() keybase1.BinaryKID
	GetFingerprintP() *PGPFingerprint
	GetAlgoType() AlgoType

	// Sign to an ASCII signature (which includes the message
	// itself) and return it, along with a derived ID.
	SignToString(msg []byte) (sig string, id keybase1.SigID, err error)

	// Verify that the given signature is valid and extracts the
	// embedded message from it. Also returns the signature ID.
	VerifyStringAndExtract(sig string, debugLogger func(s string)) (msg []byte, id keybase1.SigID, err error)

	// Verify that the given signature is valid and that its
	// embedded message matches the given one. Also returns the
	// signature ID.
	VerifyString(sig string, msg []byte) (id keybase1.SigID, err error)

	// Encrypt to an ASCII armored encryption; optionally include a sender's
	// (private) key so that we can provably see who sent the message.
	EncryptToString(plaintext []byte, sender GenericKey) (ciphertext string, err error)

	// Decrypt the output of Encrypt above; provide the plaintext and also
	// the KID of the key that sent the message (if applicable).
	DecryptFromString(ciphertext string) (msg []byte, sender keybase1.KID, err error)

	VerboseDescription() string
	CheckSecretKey() error
	CanSign() bool
	CanEncrypt() bool
	CanDecrypt() bool
	HasSecretKey() bool
	Encode() (string, error) // encode public key to string

	// ExportPublicAndPrivate halves of this key. Pass the public bytes through, but encrypt
	// the private bytes via the given encryptor. That encryptor, will, via closures, capture
	// the
	ExportPublicAndPrivate(encryptor func(private []byte) (error, []byte)) (public []byte, err error)
}

// Any valid key matches the empty string.
func KeyMatchesQuery(key GenericKey, q string, exact bool) bool {
	if key.GetKID().Match(q, exact) {
		return true
	}
	return key.GetFingerprintP().Match(q, exact)
}

func GenericKeyEqual(k1, k2 GenericKey) bool {
	return k1.GetKID().Equal(k2.GetKID())
}

func (p *PGPFingerprint) Match(q string, exact bool) bool {
	if p == nil {
		return false
	}
	if exact {
		return strings.ToLower(p.String()) == strings.ToLower(q)
	}
	return strings.HasSuffix(strings.ToLower(p.String()), strings.ToLower(q))
}

func (p PGPFingerprint) Eq(p2 PGPFingerprint) bool {
	return FastByteArrayEq(p[:], p2[:])
}
