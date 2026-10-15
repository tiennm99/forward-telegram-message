package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// ErrChatNotFound means a configured chat is not in the account's dialogs.
// Retrying cannot fix it, so callers should stop instead of reconnecting.
var ErrChatNotFound = errors.New("chat not found in this account's chats")

// dialogsPageSize is how many dialogs each lookup request fetches.
const dialogsPageSize = 100

// maxFloodWait caps how long one call may sleep on FLOOD_WAIT before giving up.
const maxFloodWait = 5 * time.Minute

// route holds the resolved source and target chats.
type route struct {
	source tg.InputPeerClass
	target tg.InputPeerClass
}

// resolveRoute finds the source and target chats in the account's dialogs.
//
// A marked chat ID alone cannot address a supergroup, channel, or user:
// Telegram also needs its access hash. The dialog list carries both, so the
// account must already be in both chats.
func resolveRoute(ctx context.Context, api *tg.Client, sourceID, targetID int64) (route, error) {
	var r route
	iter := dialogs.NewQueryBuilder(api).GetDialogs().BatchSize(dialogsPageSize).Iter()
	for (r.source == nil || r.target == nil) && iter.Next(ctx) {
		peer := iter.Value().Peer
		switch id, _ := inputPeerChatID(peer); id {
		case sourceID:
			r.source = peer
		case targetID:
			r.target = peer
		}
	}
	if err := iter.Err(); err != nil {
		return route{}, fmt.Errorf("list chats: %w", err)
	}
	if r.source == nil {
		return route{}, fmt.Errorf("source %d: %w", sourceID, ErrChatNotFound)
	}
	if r.target == nil {
		return route{}, fmt.Errorf("target %d: %w", targetID, ErrChatNotFound)
	}
	return r, nil
}

// forwarder queues new source messages from the updates loop and forwards
// them in order from a separate worker. Updates can arrive before the route
// is resolved, so route is read atomically.
type forwarder struct {
	sourceID int64
	route    atomic.Pointer[route]
	queue    *queue
	out      io.Writer
}

// handle queues m when it is a source message. It never calls Telegram:
// it runs inside the updates loop, which must keep draining while a forward
// waits on the network or a flood wait.
func (f *forwarder) handle(m tg.MessageClass) {
	if f.route.Load() == nil {
		return
	}
	if msg, ok := sourceMessage(m, f.sourceID); ok {
		f.queue.push(msg.ID)
	}
}

// work forwards queued messages until ctx ends. A failed forward is reported
// and skipped so later messages still go through; one cut short by ctx stays
// queued for the next connection.
func (f *forwarder) work(ctx context.Context, api *tg.Client) {
	for {
		id, err := f.queue.front(ctx)
		if err != nil {
			return
		}
		r := f.route.Load()
		req := &tg.MessagesForwardMessagesRequest{
			FromPeer: r.source,
			ToPeer:   r.target,
			ID:       []int{id},
			RandomID: []int64{rand.Int64()},
		}
		_, err = withFloodWait(ctx, func() (tg.UpdatesClass, error) {
			return api.MessagesForwardMessages(ctx, req)
		})
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			fmt.Fprintf(f.out, "Exception on forward_message: %v\n", err)
		}
		f.queue.pop()
	}
}

// sourceMessage returns m as a regular message when it belongs to the source
// chat. Service messages such as joins and pins are skipped.
func sourceMessage(m tg.MessageClass, sourceID int64) (*tg.Message, bool) {
	msg, ok := m.(*tg.Message)
	if !ok {
		return nil, false
	}
	if id, ok := peerChatID(msg.PeerID); !ok || id != sourceID {
		return nil, false
	}
	return msg, true
}

// withFloodWait retries call after Telegram's requested FLOOD_WAIT delay, so
// a burst of source messages does not drop forwards on rate limits.
func withFloodWait[T any](ctx context.Context, call func() (T, error)) (T, error) {
	for {
		result, err := call()
		wait, ok := tgerr.AsFloodWait(err)
		if !ok || wait > maxFloodWait {
			return result, err
		}
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(wait + time.Second):
		}
	}
}
