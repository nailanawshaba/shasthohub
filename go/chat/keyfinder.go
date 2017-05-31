package chat

import (
	"fmt"
	"sync"

	"github.com/keybase/client/go/chat/types"
	"github.com/keybase/client/go/protocol/chat1"
	context "golang.org/x/net/context"
)

// KeyFinder remembers results from previous calls to CryptKeys().
type KeyFinder interface {
	Find(ctx context.Context, tlf types.CryptKeysSource, name string, public bool) (types.CryptKeysRes, error)
}

type KeyFinderImpl struct {
	sync.Mutex
	keys map[string]types.CryptKeysRes
}

// NewKeyFinder creates a KeyFinder.
func NewKeyFinder() KeyFinder {
	return &KeyFinderImpl{
		keys: make(map[string]types.CryptKeysRes),
	}
}

func (k *KeyFinderImpl) cacheKey(name string, public bool) string {
	return fmt.Sprintf("%s|%v", name, public)
}

// Find finds keybase1.TLFCryptKeys for tlfName, checking for existing
// results.
func (k *KeyFinderImpl) Find(ctx context.Context, cks types.CryptKeysSource, name string,
	public bool) (types.CryptKeysRes, error) {

	ckey := k.cacheKey(tlfName, tlfPublic)
	k.Lock()
	existing, ok := k.keys[ckey]
	k.Unlock()
	if ok {
		return existing, nil
	}

	vis := chat1.TLFVisibility_PRIVATE
	if public {
		vis = chat1.TLFVisibility_PUBLIC
	}
	res, err := cks.CryptKeys(ctx, name, vis)
	if err != nil {
		return types.CryptKeysRes{}, err
	}

	k.Lock()
	k.keys[ckey] = res
	k.Unlock()

	return res, nil
}
