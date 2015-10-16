package service

import (
	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
)

// FavoriteHandler implements the keybase1.Favorite protocol
type FavoriteHandler struct {
	*BaseHandler
	libkb.Contextified
}

// NewFavoriteHandler creates a FavoriteHandler with the xp
// protocol.
func NewFavoriteHandler(xp rpc.Transporter, g *libkb.GlobalContext) *FavoriteHandler {
	return &FavoriteHandler{
		BaseHandler:  NewBaseHandler(xp),
		Contextified: libkb.NewContextified(g),
	}
}

// FavoriteAdd handles the favoriteAdd RPC.
func (h *FavoriteHandler) FavoriteAdd(arg keybase1.FavoriteAddArg) error {
	eng := engine.NewFavoriteAdd(&arg, h.G())
	ctx := &engine.Context{}
	return engine.RunEngine(eng, ctx)
}

// FavoriteDelete handles the favoriteDelete RPC.
func (h *FavoriteHandler) FavoriteDelete(arg keybase1.FavoriteDeleteArg) error {
	eng := engine.NewFavoriteDelete(&arg, h.G())
	ctx := &engine.Context{}
	return engine.RunEngine(eng, ctx)
}

// FavoriteList handles the favoriteList RPC.
func (h *FavoriteHandler) FavoriteList(sessionID int) ([]keybase1.Folder, error) {
	eng := engine.NewFavoriteList(h.G())
	ctx := &engine.Context{}
	if err := engine.RunEngine(eng, ctx); err != nil {
		return nil, err
	}
	return eng.Favorites(), nil
}
