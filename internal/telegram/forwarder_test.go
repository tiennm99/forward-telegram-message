package telegram

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestSourceMessage(t *testing.T) {
	const source = -1001234567890
	tests := []struct {
		name string
		msg  tg.MessageClass
		ok   bool
	}{
		{"source channel", &tg.Message{ID: 1, PeerID: &tg.PeerChannel{ChannelID: 1234567890}}, true},
		{"outgoing in source", &tg.Message{ID: 2, Out: true, PeerID: &tg.PeerChannel{ChannelID: 1234567890}}, true},
		{"other channel", &tg.Message{ID: 3, PeerID: &tg.PeerChannel{ChannelID: 99}}, false},
		{"basic group with same raw id", &tg.Message{ID: 4, PeerID: &tg.PeerChat{ChatID: 1234567890}}, false},
		{"service message", &tg.MessageService{ID: 5, PeerID: &tg.PeerChannel{ChannelID: 1234567890}}, false},
		{"empty message", &tg.MessageEmpty{ID: 6}, false},
	}
	for _, tt := range tests {
		msg, ok := sourceMessage(tt.msg, source)
		if ok != tt.ok || (ok && msg != tt.msg) {
			t.Errorf("%s: got %v, %v; want ok %v", tt.name, msg, ok, tt.ok)
		}
	}

	private := &tg.Message{ID: 7, PeerID: &tg.PeerUser{UserID: 555}}
	if _, ok := sourceMessage(private, 555); !ok {
		t.Error("private chat message from source user skipped")
	}
}

// fakeInvoker records forwarded message IDs and fails or blocks on demand.
type fakeInvoker struct {
	forwarded chan int
	fail      bool
	block     bool
}

func (f *fakeInvoker) Invoke(ctx context.Context, input bin.Encoder, _ bin.Decoder) error {
	if f.block {
		<-ctx.Done()
		return ctx.Err()
	}
	req := input.(*tg.MessagesForwardMessagesRequest)
	f.forwarded <- req.ID[0]
	if f.fail {
		return errors.New("forward failed")
	}
	return nil
}

func newTestForwarder(out *bytes.Buffer) *forwarder {
	f := &forwarder{sourceID: 555, queue: newQueue(), out: out}
	f.route.Store(&route{source: &tg.InputPeerUser{UserID: 555}, target: &tg.InputPeerChat{ChatID: 42}})
	return f
}

func TestHandleQueuesWithoutTelegram(t *testing.T) {
	f := newTestForwarder(&bytes.Buffer{})
	for i := 1; i <= 100; i++ {
		f.handle(&tg.Message{ID: i, PeerID: &tg.PeerUser{UserID: 555}})
	}
	f.handle(&tg.Message{ID: 999, PeerID: &tg.PeerUser{UserID: 1}})
	if got := len(f.queue.ids); got != 100 {
		t.Fatalf("queued %d messages, want 100", got)
	}

	unrouted := &forwarder{sourceID: 555, queue: newQueue()}
	unrouted.handle(&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 555}})
	if len(unrouted.queue.ids) != 0 {
		t.Fatal("queued a message before the route was resolved")
	}
}

func TestWorkForwardsInOrderAndSkipsFailures(t *testing.T) {
	var out bytes.Buffer
	f := newTestForwarder(&out)
	inv := &fakeInvoker{forwarded: make(chan int), fail: true}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		f.work(ctx, tg.NewClient(inv))
	}()

	for _, id := range []int{1, 2, 3} {
		f.queue.push(id)
	}
	for _, want := range []int{1, 2, 3} {
		if got := <-inv.forwarded; got != want {
			t.Fatalf("forwarded %d, want %d", got, want)
		}
	}
	cancel()
	<-done
	if !strings.Contains(out.String(), "Exception on forward_message: ") {
		t.Fatalf("failure not reported: %q", out.String())
	}
}

func TestWorkKeepsInterruptedForwardQueued(t *testing.T) {
	f := newTestForwarder(&bytes.Buffer{})
	f.queue.push(7)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	f.work(ctx, tg.NewClient(&fakeInvoker{block: true}))
	if len(f.queue.ids) != 1 || f.queue.ids[0] != 7 {
		t.Fatalf("queue = %v, want [7]", f.queue.ids)
	}
}
