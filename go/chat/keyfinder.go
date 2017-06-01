package chat

import (
	"fmt"
	"sync"

	"github.com/keybase/client/go/chat/globals"
	"github.com/keybase/client/go/chat/types"
	"github.com/keybase/client/go/protocol/chat1"
	context "golang.org/x/net/context"
)

// KeyFinder remembers results from previous calls to CryptKeys().
type KeyFinder interface {
	Find(ctx context.Context, name string, membersType chat1.ConversationMembersType, public bool) (types.NameInfo, error)
}

type KeyFinderImpl struct {
	globals.Contextified
	sync.Mutex
	keys map[string]types.NameInfo
}

// NewKeyFinder creates a KeyFinder.
func NewKeyFinder(g *globals.Context) KeyFinder {
	return &KeyFinderImpl{
		Contextified: globals.NewContextified(g),
		keys:         make(map[string]types.NameInfo),
	}
}

func (k *KeyFinderImpl) cacheKey(name string, membersType chat1.ConversationMembersType, public bool) string {
	return fmt.Sprintf("%s|%v|%v", name, membersType, public)
}

func (k *KeyFinderImpl) createNameInfoSource(ctx context.Context,
	membersType chat1.ConversationMembersType) types.NameInfoSource {
	switch membersType {
	case chat1.ConversationMembersType_KBFS:
		return NewKBFSNameInfoSource(g)
	case chat1.ConversationMembersType_TEAM:
		return NewTeamsNameInfoSource(g)
	}
	t.Debug(ctx, "createNameInfoSource: unknown members type, using KBFS: %v", membersType)
	return NewKBFSNameInfoSource(g)
}

// Find finds keybase1.TLFCryptKeys for tlfName, checking for existing
// results.
func (k *KeyFinderImpl) Find(ctx context.Context, name string,
	membersType chat1.ConversationMembersType, public bool) (types.NameInfo, error) {

	ckey := k.cacheKey(tlfName, membersType, tlfPublic)
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
	nameSource := k.createNameInfoSource(membersType)
	nameInfo, err := nameSource.Lookup(ctx, name, vis)
	if err != nil {
		return NameInfo{}, err
	}

	k.Lock()
	k.keys[ckey] = nameInfo
	k.Unlock()

	return nameInfo, nil
}
