package debug

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestDebugPrint(t *testing.T) {
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })

	viper.Set("debug", false)
	DebugPrint("hidden %d", 1)
	if got := buf.String(); got != "" {
		t.Fatalf("expected no output when debug=false, got %q", got)
	}

	viper.Set("debug", true)
	DebugPrint("value=%d", 42)
	if got := buf.String(); !strings.Contains(got, "[DEBUG] value=42") {
		t.Fatalf("expected debug output, got %q", got)
	}
}
