package warnings

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmit_DefaultShowsWarningAndHint(t *testing.T) {
	ResetForTest()
	Init(nil, false)
	var buf bytes.Buffer
	Emitf(&buf, KeyTLSInsecure, "tls off")
	s := buf.String()
	if !strings.Contains(s, "tls off") {
		t.Errorf("missing message: %q", s)
	}
	if !strings.Contains(s, "tip:") || !strings.Contains(s, string(KeyTLSInsecure)) {
		t.Errorf("missing first-time hint: %q", s)
	}
}

func TestEmit_HintShownOnlyOnce(t *testing.T) {
	ResetForTest()
	Init(nil, false)
	var buf bytes.Buffer
	Emitf(&buf, KeyTLSInsecure, "a")
	Emitf(&buf, KeyTLSInsecure, "b")
	s := buf.String()
	if strings.Count(s, "tip:") != 1 {
		t.Errorf("hint should appear once, got %q", s)
	}
}

func TestEmit_SuppressedByConfig(t *testing.T) {
	ResetForTest()
	Init([]string{string(KeyTLSInsecure)}, false)
	var buf bytes.Buffer
	Emitf(&buf, KeyTLSInsecure, "silenced")
	if buf.Len() != 0 {
		t.Errorf("expected no output when suppressed, got %q", buf.String())
	}
	// Other keys still emit
	Emitf(&buf, KeyFlatLayout, "visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Errorf("non-suppressed key should emit: %q", buf.String())
	}
}

func TestShowAllOverridesSuppression(t *testing.T) {
	ResetForTest()
	Init([]string{string(KeyTLSInsecure), string(KeyFlatLayout)}, true)
	var buf bytes.Buffer
	Emitf(&buf, KeyTLSInsecure, "forced")
	if !strings.Contains(buf.String(), "forced") {
		t.Errorf("--show-warnings should override suppress: %q", buf.String())
	}
	if Suppressed(KeyTLSInsecure) {
		t.Error("Suppressed() should return false when showAll=true")
	}
}

func TestSuppressed_UnknownKey_False(t *testing.T) {
	ResetForTest()
	Init(nil, false)
	if Suppressed(Key("never_configured")) {
		t.Error("unknown key should not be suppressed")
	}
}
