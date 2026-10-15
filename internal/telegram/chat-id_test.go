package telegram

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestPeerChatID(t *testing.T) {
	tests := []struct {
		peer tg.PeerClass
		want int64
	}{
		{&tg.PeerUser{UserID: 777000}, 777000},
		{&tg.PeerChat{ChatID: 42}, -42},
		{&tg.PeerChannel{ChannelID: 1234567890}, -1001234567890},
	}
	for _, tt := range tests {
		got, ok := peerChatID(tt.peer)
		if !ok || got != tt.want {
			t.Errorf("peerChatID(%v) = %d, %v; want %d", tt.peer, got, ok, tt.want)
		}
	}
	if _, ok := peerChatID(nil); ok {
		t.Error("peerChatID(nil) ok")
	}
}

func TestInputPeerChatID(t *testing.T) {
	tests := []struct {
		peer tg.InputPeerClass
		want int64
		ok   bool
	}{
		{&tg.InputPeerUser{UserID: 5, AccessHash: 1}, 5, true},
		{&tg.InputPeerChat{ChatID: 42}, -42, true},
		{&tg.InputPeerChannel{ChannelID: 1, AccessHash: 1}, -1000000000001, true},
		{&tg.InputPeerEmpty{}, 0, false},
	}
	for _, tt := range tests {
		got, ok := inputPeerChatID(tt.peer)
		if got != tt.want || ok != tt.ok {
			t.Errorf("inputPeerChatID(%v) = %d, %v; want %d, %v", tt.peer, got, ok, tt.want, tt.ok)
		}
	}
}
