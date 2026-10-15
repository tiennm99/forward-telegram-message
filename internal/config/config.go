// Package config reads the forwarder settings from the environment.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gotd/td/constant"
	"github.com/joho/godotenv"
)

// Config is what the forwarder needs to log in and route messages.
type Config struct {
	Phone    string
	APIID    int
	APIHash  string
	SourceID int64
	TargetID int64
}

// FromEnv loads .env when present, then reads the process environment.
func FromEnv() (Config, error) {
	_ = godotenv.Load()
	return Parse(os.Getenv)
}

// Parse reads every setting through getenv and reports all missing ones at
// once.
func Parse(getenv func(string) string) (Config, error) {
	value := func(name string) string { return strings.TrimSpace(getenv(name)) }

	var missing []string
	for _, name := range []string{"PHONE", "API_ID", "API_HASH", "SOURCE_GROUP_ID", "TARGET_GROUP_ID"} {
		if value(name) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing %s in environment or .env", strings.Join(missing, ", "))
	}

	cfg := Config{Phone: value("PHONE"), APIHash: value("API_HASH")}
	var err error
	if cfg.APIID, err = strconv.Atoi(value("API_ID")); err != nil || cfg.APIID <= 0 {
		return Config{}, fmt.Errorf("invalid API_ID %q: must be a positive integer", value("API_ID"))
	}
	if cfg.SourceID, err = parseChatID("SOURCE_GROUP_ID", value("SOURCE_GROUP_ID")); err != nil {
		return Config{}, err
	}
	if cfg.TargetID, err = parseChatID("TARGET_GROUP_ID", value("TARGET_GROUP_ID")); err != nil {
		return Config{}, err
	}
	if cfg.SourceID == cfg.TargetID {
		return Config{}, errors.New("SOURCE_GROUP_ID and TARGET_GROUP_ID must differ")
	}
	return cfg, nil
}

func parseChatID(name, raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || !validChatID(id) {
		return 0, fmt.Errorf("invalid %s %q: use a marked chat ID such as -1001234567890", name, raw)
	}
	return id, nil
}

// validChatID reports whether id is a marked chat ID: the other person's user
// ID for private chats, -<id> for basic groups and -100<id> for supergroups
// and channels.
func validChatID(id int64) bool {
	marked := constant.TDLibPeerID(id)
	return marked.IsUser() || marked.IsChat() || marked.IsChannel()
}
