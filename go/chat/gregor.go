package chat

import (
	"context"
	"fmt"

	"github.com/keybase/client/go/gregor"
	"github.com/keybase/client/go/protocol/chat1"
	"github.com/keybase/client/go/protocol/gregor1"
	"github.com/keybase/client/go/protocol/keybase1"
	"github.com/keybase/go-framed-msgpack-rpc/rpc"
)

type GregorConnection struct {
	gregor.Handler
}

func NewGregorConnection(conn *gregor.Handler) *GregorConnection {
	return &GregorConnection{
		Handler: conn,
	}
}

// OnConnect is called by the rpc library to indicate we have connected to
// gregord
func (g *GregorConnection) OnConnect(ctx context.Context, conn *rpc.Connection,
	cli rpc.GenericClient, srv *rpc.Server) error {
	g.Lock()
	defer g.Unlock()

	// If we get a random OnConnect on some other connection that is not g.conn, then
	// just reject it.
	if conn != g.conn {
		return gregor.ErrDuplicateConnection
	}

	timeoutCli := WrapGenericClientWithTimeout(cli, GregorRequestTimeout, ErrChatServerTimeout)
	chatCli := chat1.RemoteClient{Cli: NewRemoteClient(g.G(), cli)}

	g.chatLog.Debug(ctx, "connected")
	if err := srv.Register(gregor1.OutgoingProtocol(g)); err != nil {
		return fmt.Errorf("error registering protocol: %s", err.Error())
	}

	// Grab authentication and sync params
	gcli, err := g.getGregorCli()
	if err != nil {
		return fmt.Errorf("failed to get gregor client: %s", err.Error())
	}
	uid, token, err := g.authParams(ctx)
	if err != nil {
		return err
	}
	iboxVers := g.inboxParams(ctx, uid)
	latestCtime := g.notificationParams(ctx, gcli)

	// Run SyncAll to both authenticate, and grab all the data we will need to run the
	// various resync procedures for chat and notifications
	var identBreaks []keybase1.TLFIdentifyFailure
	ctx = Context(ctx, g.G(), keybase1.TLFIdentifyBehavior_CHAT_GUI, &identBreaks,
		NewIdentifyNotifier(g.G()))
	syncAllRes, err := chatCli.SyncAll(ctx, chat1.SyncAllArg{
		Uid:       uid,
		DeviceID:  gcli.Device.(gregor1.DeviceID),
		Session:   token,
		InboxVers: iboxVers,
		Ctime:     latestCtime,
		Fresh:     g.firstConnect,
	})
	if err != nil {
		return fmt.Errorf("error running SyncAll: %s", err.Error())
	}

	// Use the client parameter instead of conn.GetClient(), since we can get stuck
	// in a recursive loop if we keep retrying on reconnect.
	if err := g.auth(ctx, timeoutCli, &syncAllRes.Auth); err != nil {
		return fmt.Errorf("error authenticating: %s", err.Error())
	}

	// Sync chat data using a Syncer object
	if err := g.G().Syncer.Connected(ctx, chatCli, uid, &syncAllRes.Chat); err != nil {
		return fmt.Errorf("error running chat sync: %s", err.Error())
	}

	// Sync down events since we have been dead
	replayedMsgs, consumedMsgs, err := g.serverSync(ctx, gregor1.IncomingClient{Cli: timeoutCli}, gcli,
		&syncAllRes.Notification)
	if err != nil {
		g.chatLog.Debug(ctx, "sync failure: %s", err.Error())
	} else {
		g.chatLog.Debug(ctx, "sync success: replayed: %d consumed: %d", len(replayedMsgs),
			len(consumedMsgs))
	}

	// Sync badge state in the background
	if g.badger != nil {
		if err := g.badger.Resync(ctx, g.GetClient, gcli, &syncAllRes.Badge); err != nil {
			g.chatLog.Debug(ctx, "badger failure: %s", err.Error())
		}
	}

	// Call out to reachability module if we have one
	if g.reachability != nil {
		g.reachability.setReachability(keybase1.Reachability{
			Reachable: keybase1.Reachable_YES,
		})
	}

	// Broadcast reconnect oobm. Spawn this off into a goroutine so that we don't delay
	// reconnection any longer than we have to.
	go func(m gregor1.Message) {
		g.BroadcastMessage(context.Background(), m)
	}(g.makeReconnectOobm())

	// No longer first connect if we are now connected
	g.firstConnect = false

	return nil
}

func (g *GregorConnecion) GetClient() chat1.RemoteInterface {
	if g.IsShutdown() || g.cli == nil {
		g.chatLog.Debug(context.Background(), "GetClient: shutdown, using errorClient for chat1.RemoteClient")
		return chat1.RemoteClient{Cli: OfflineClient{}}
	}
	return chat1.RemoteClient{Cli: NewRemoteClient(g.G(), g.cli)}
}

func (g *gregorHandler) HandleOutOfBandMessage(ctx context.Context, obm gregor.OutOfBandMessage) error {
	if obm.System() == nil {
		return g.Handler.HandleOutOfBandMessage(ctx, obm)
	}

	switch obm.System().String() {
	case "chat.activity":
		return g.G().PushHandler.Activity(ctx, obm)
	case "chat.tlffinalize":
		return g.G().PushHandler.TlfFinalize(ctx, obm)
	case "chat.tlfresolve":
		return g.G().PushHandler.TlfResolve(ctx, obm)
	case "chat.typing":
		return g.G().PushHandler.Typing(ctx, obm)
	case "chat.membershipUpdate":
		return g.G().PushHandler.MembershipUpdate(ctx, obm)
	}

	return g.Handler.HandleOutOfBandMessage(ctx, obm)
}
