package forward

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunRejectsMissingConfigBeforeTelegram(t *testing.T) {
	for _, name := range []string{"PHONE", "API_ID", "API_HASH", "SOURCE_GROUP_ID", "TARGET_GROUP_ID"} {
		t.Setenv(name, "")
	}
	var out bytes.Buffer

	err := Run(context.Background(), strings.NewReader(""), &out)
	if err == nil || !strings.Contains(err.Error(), "missing PHONE") {
		t.Fatalf("Run error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("Run wrote %q before failing", out.String())
	}
}
