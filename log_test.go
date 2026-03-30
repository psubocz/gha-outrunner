package outrunner

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestSimpleHandlerOutput(t *testing.T) {
	var buf bytes.Buffer
	h := NewSimpleHandler(&buf, slog.LevelInfo)
	logger := slog.New(h)

	ts := time.Date(2026, 3, 30, 14, 5, 9, 0, time.UTC)
	r := slog.NewRecord(ts, slog.LevelInfo, "hello world", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "2026-03-30 14:05:09") {
		t.Errorf("expected timestamp, got %q", got)
	}
	if !strings.Contains(got, "INFO") {
		t.Errorf("expected INFO level, got %q", got)
	}
	if !strings.Contains(got, "hello world") {
		t.Errorf("expected message, got %q", got)
	}

	// Verify it ends with a newline
	if !strings.HasSuffix(got, "\n") {
		t.Errorf("expected trailing newline, got %q", got)
	}

	// Now test with attrs
	buf.Reset()
	logger.Info("loaded", slog.Int("runners", 3), slog.String("file", "config.yml"))
	got = buf.String()
	if !strings.Contains(got, "runners=3") {
		t.Errorf("expected runners=3, got %q", got)
	}
	if !strings.Contains(got, "file=config.yml") {
		t.Errorf("expected file=config.yml, got %q", got)
	}
}

func TestSimpleHandlerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	h := NewSimpleHandler(&buf, slog.LevelWarn)

	if h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("DEBUG should be disabled at WARN level")
	}
	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("INFO should be disabled at WARN level")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("WARN should be enabled at WARN level")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Error("ERROR should be enabled at WARN level")
	}
}

func TestSimpleHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	h := NewSimpleHandler(&buf, slog.LevelInfo)
	logger := slog.New(h).WithGroup("scaler")

	logger.Info("started", slog.String("name", "test-runner"))
	got := buf.String()
	if !strings.Contains(got, "scaler.name=test-runner") {
		t.Errorf("expected grouped attr, got %q", got)
	}
}

func TestSimpleHandlerNestedGroups(t *testing.T) {
	var buf bytes.Buffer
	h := NewSimpleHandler(&buf, slog.LevelInfo)
	logger := slog.New(h).WithGroup("docker").WithGroup("container")

	logger.Info("created", slog.String("id", "abc123"))
	got := buf.String()
	if !strings.Contains(got, "docker.container.id=abc123") {
		t.Errorf("expected nested groups, got %q", got)
	}
}

func TestSimpleHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := NewSimpleHandler(&buf, slog.LevelInfo)
	logger := slog.New(h).With(slog.String("scaleSet", "linux"))

	logger.Info("ready")
	got := buf.String()
	if !strings.Contains(got, "scaleSet=linux") {
		t.Errorf("expected pre-set attr, got %q", got)
	}

	// Pre-set attrs should appear on every message
	buf.Reset()
	logger.Info("second")
	got = buf.String()
	if !strings.Contains(got, "scaleSet=linux") {
		t.Errorf("expected pre-set attr on second message, got %q", got)
	}
}

func TestSimpleHandlerEmptyAttr(t *testing.T) {
	var buf bytes.Buffer
	h := NewSimpleHandler(&buf, slog.LevelInfo)

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
	r.AddAttrs(slog.Attr{}) // empty attr should be skipped
	r.AddAttrs(slog.String("key", "val"))

	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if !strings.Contains(got, "key=val") {
		t.Errorf("expected key=val, got %q", got)
	}
	// Should not have extra spaces from empty attr
	if strings.Contains(got, "  ") {
		t.Errorf("unexpected double space (empty attr not skipped), got %q", got)
	}
}
