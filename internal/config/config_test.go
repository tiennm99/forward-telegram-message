package config

import (
	"strings"
	"testing"
)

func env(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func validEnv() map[string]string {
	return map[string]string{
		"PHONE":           "+84912345678",
		"API_ID":          "123456",
		"API_HASH":        "0123456789abcdef",
		"SOURCE_GROUP_ID": "-1001234567890",
		"TARGET_GROUP_ID": " -42 ",
	}
}

func TestParse(t *testing.T) {
	got, err := Parse(env(validEnv()))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := Config{
		Phone:    "+84912345678",
		APIID:    123456,
		APIHash:  "0123456789abcdef",
		SourceID: -1001234567890,
		TargetID: -42,
	}
	if got != want {
		t.Fatalf("Parse = %+v, want %+v", got, want)
	}
}

func TestParseReportsAllMissing(t *testing.T) {
	_, err := Parse(env(map[string]string{"PHONE": "+1"}))
	if err == nil || !strings.Contains(err.Error(), "API_ID, API_HASH, SOURCE_GROUP_ID, TARGET_GROUP_ID") {
		t.Fatalf("Parse error = %v", err)
	}
}

func TestParseRejectsBadValues(t *testing.T) {
	tests := []struct {
		name, value, want string
	}{
		{"API_ID", "abc", "invalid API_ID"},
		{"API_ID", "-5", "invalid API_ID"},
		{"SOURCE_GROUP_ID", "group", "invalid SOURCE_GROUP_ID"},
		{"SOURCE_GROUP_ID", "-1000000000000", "invalid SOURCE_GROUP_ID"},
		{"TARGET_GROUP_ID", "-1001234567890", "must differ"},
	}
	for _, tt := range tests {
		values := validEnv()
		values[tt.name] = tt.value
		if _, err := Parse(env(values)); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s=%q: error = %v, want %q", tt.name, tt.value, err, tt.want)
		}
	}
}

func TestValidChatID(t *testing.T) {
	for _, id := range []int64{777000, -42, -999999999999, -1000000000001, -1001234567890} {
		if !validChatID(id) {
			t.Errorf("validChatID(%d) = false", id)
		}
	}
	for _, id := range []int64{0, -1000000000000} {
		if validChatID(id) {
			t.Errorf("validChatID(%d) = true", id)
		}
	}
}
