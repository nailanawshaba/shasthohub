package all

import (
	crypto "github.com/keybase/client/go/crypto"
	nacl "github.com/keybase/client/go/crypto/nacl"
	pgp "github.com/keybase/client/go/crypto/pgp"
	keybase1 "github.com/keybase/client/go/protocol"
)

func ParseGenericKey(bundle string) (crypto.GenericKey, *crypto.Warnings, error) {
	if pgp.IsPGPBundle(bundle) {
		// PGP key
		return pgp.ReadOneKeyFromString(bundle)
	}
	// NaCl key
	key, err := nacl.ImportKeypairFromKID(keybase1.KIDFromString(bundle))
	return key, &Warnings{}, err
}

func CanEncrypt(key crypto.GenericKey) bool {
	switch key.(type) {
	case nacl.NaclDHKeyPair:
		return true
	case *pgp.PGPKeyBundle:
		return true
	default:
		return false
	}
}
