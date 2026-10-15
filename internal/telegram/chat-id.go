package telegram

import (
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
)

// Chat IDs use the usual "marked" form: the other person's user ID for private
// chats, -<id> for basic groups and -100<id> for supergroups and channels.

// peerChatID returns the marked ID of a message's chat.
func peerChatID(p tg.PeerClass) (int64, bool) {
	var id constant.TDLibPeerID
	switch p := p.(type) {
	case *tg.PeerUser:
		id.User(p.UserID)
	case *tg.PeerChat:
		id.Chat(p.ChatID)
	case *tg.PeerChannel:
		id.Channel(p.ChannelID)
	default:
		return 0, false
	}
	return int64(id), true
}

// inputPeerChatID returns the marked ID of a dialog's input peer.
func inputPeerChatID(p tg.InputPeerClass) (int64, bool) {
	var id constant.TDLibPeerID
	switch p := p.(type) {
	case *tg.InputPeerUser:
		id.User(p.UserID)
	case *tg.InputPeerChat:
		id.Chat(p.ChatID)
	case *tg.InputPeerChannel:
		id.Channel(p.ChannelID)
	default:
		return 0, false
	}
	return int64(id), true
}
