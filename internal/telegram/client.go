// Package telegram logs in to Telegram and forwards new messages.
package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"github.com/tiennm99/forward-telegram-message/internal/config"
	"github.com/tiennm99/forward-telegram-message/internal/prompt"
)

// ErrLogin means signing in failed in a way reconnecting cannot fix, such as
// no answer on the terminal or Telegram rejecting the phone, code, or
// password. Retrying would only request more login codes.
var ErrLogin = errors.New("telegram login failed")

var errNoAccount = fmt.Errorf("%w: this phone has no Telegram account; sign up in an official app first", ErrLogin)

// Listener forwards new source messages to the target across reconnects.
// It keeps the update state and the pending forwards between Listen calls,
// so messages sent while reconnecting are caught up.
type Listener struct {
	cfg  config.Config
	out  io.Writer
	fwd  *forwarder
	gaps *updates.Manager
}

// NewListener returns a Listener for cfg that reports on out.
func NewListener(cfg config.Config, out io.Writer) *Listener {
	fwd := &forwarder{sourceID: cfg.SourceID, queue: newQueue(), out: out}
	dispatcher := tg.NewUpdateDispatcher()
	dispatcher.OnNewMessage(func(_ context.Context, _ tg.Entities, u *tg.UpdateNewMessage) error {
		fwd.handle(u.Message)
		return nil
	})
	dispatcher.OnNewChannelMessage(func(_ context.Context, _ tg.Entities, u *tg.UpdateNewChannelMessage) error {
		fwd.handle(u.Message)
		return nil
	})
	return &Listener{
		cfg:  cfg,
		out:  out,
		fwd:  fwd,
		gaps: updates.New(updates.Config{Handler: dispatcher}),
	}
}

// Listen connects with the stored session, signs in only when that session is
// missing or expired, and forwards every new source message to the target
// until ctx ends or the connection fails.
func (l *Listener) Listen(ctx context.Context, storage session.Storage, ask *prompt.Prompter) error {
	defer l.gaps.Reset()
	client := telegram.NewClient(l.cfg.APIID, l.cfg.APIHash, telegram.Options{
		SessionStorage: storage,
		UpdateHandler:  l.gaps,
		Middlewares:    []telegram.Middleware{hook.UpdateHook(l.gaps.Handle)},
	})

	return client.Run(ctx, func(ctx context.Context) error {
		flow := auth.NewFlow(terminalAuth{phone: l.cfg.Phone, ask: ask}, auth.SendCodeOptions{})
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			if loginFatal(err) {
				return fmt.Errorf("%w: %w", ErrLogin, err)
			}
			return fmt.Errorf("telegram login: %w", err)
		}
		self, err := client.Self(ctx)
		if err != nil {
			return fmt.Errorf("read own account: %w", err)
		}
		r, err := resolveRoute(ctx, client.API(), l.cfg.SourceID, l.cfg.TargetID)
		if err != nil {
			return err
		}
		l.fwd.route.Store(&r)

		ctx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		go func() {
			defer close(done)
			l.fwd.work(ctx, client.API())
		}()
		defer func() {
			cancel()
			<-done
		}()

		return l.gaps.Run(ctx, client.API(), self.ID, updates.AuthOptions{
			IsBot: self.Bot,
			OnStart: func(context.Context) {
				fmt.Fprintln(l.out, "Bot is running.")
			},
		})
	})
}

// loginFatal reports whether a login error is one reconnecting cannot fix:
// a failed terminal prompt, a wrong password, or Telegram rejecting the
// request (4xx errors, flood waits included). Network and server errors stay
// retryable.
func loginFatal(err error) bool {
	if errors.Is(err, ErrLogin) || errors.Is(err, auth.ErrPasswordInvalid) {
		return true
	}
	rpc, ok := tgerr.As(err)
	return ok && rpc.Code < 500
}

// terminalAuth answers the login flow from the terminal.
type terminalAuth struct {
	phone string
	ask   *prompt.Prompter
}

func (a terminalAuth) Phone(context.Context) (string, error) { return a.phone, nil }

func (a terminalAuth) Password(context.Context) (string, error) {
	return answer(a.ask.Secret("Password"))
}

func (a terminalAuth) Code(context.Context, *tg.AuthSentCode) (string, error) {
	return answer(a.ask.Line("Enter code"))
}

func (a terminalAuth) AcceptTermsOfService(context.Context, tg.HelpTermsOfService) error {
	return errNoAccount
}

func (a terminalAuth) SignUp(context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errNoAccount
}

// answer marks a failed terminal prompt as a login failure.
func answer(s string, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrLogin, err)
	}
	return s, nil
}
