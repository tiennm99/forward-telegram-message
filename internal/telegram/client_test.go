package telegram

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tgerr"
)

func TestLoginFatal(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"no terminal answer", fmt.Errorf("callback: %w", mustAnswerErr(io.EOF)), true},
		{"no account", errNoAccount, true},
		{"wrong password", fmt.Errorf("sign in: %w", auth.ErrPasswordInvalid), true},
		{"wrong code", tgerr.New(400, "PHONE_CODE_INVALID"), true},
		{"flood wait", tgerr.New(420, "FLOOD_WAIT_60"), true},
		{"server error", tgerr.New(500, "INTERNAL"), false},
		{"network", errors.New("connection reset"), false},
	}
	for _, tt := range tests {
		if got := loginFatal(tt.err); got != tt.want {
			t.Errorf("%s: loginFatal = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func mustAnswerErr(err error) error {
	_, err = answer("", err)
	return err
}
