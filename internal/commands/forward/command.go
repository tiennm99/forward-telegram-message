// Package forward runs the forwarder and reconnects when it fails.
package forward

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gotd/td/session"

	"github.com/tiennm99/forward-telegram-message/internal/config"
	"github.com/tiennm99/forward-telegram-message/internal/prompt"
	"github.com/tiennm99/forward-telegram-message/internal/telegram"
)

// sessionFile holds the Telegram login session. It grants full account
// access, so it is written with 0600 permissions and ignored by Git.
const sessionFile = "session.json"

// restartDelay is how long to wait before reconnecting after a failure.
const restartDelay = 5 * time.Second

// Run forwards messages until ctx ends. Connection failures restart the
// forwarder; a bad config, a failed login, or an unknown chat stops it.
func Run(ctx context.Context, in io.Reader, out io.Writer) error {
	cfg, err := config.FromEnv()
	if err != nil {
		return err
	}
	storage := &session.FileStorage{Path: sessionFile}
	ask := prompt.New(in, out)
	listener := telegram.NewListener(cfg, out)

	for {
		fmt.Fprintln(out, "Starting...")
		err := listener.Listen(ctx, storage, ask)
		if ctx.Err() != nil {
			return nil
		}
		switch {
		case errors.Is(err, telegram.ErrChatNotFound):
			return fmt.Errorf("%w; the account must be a member of both chats", err)
		case errors.Is(err, telegram.ErrLogin):
			return err
		}
		fmt.Fprintf(out, "Exception on main: %v. Restarting...\n", err)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(restartDelay):
		}
	}
}
