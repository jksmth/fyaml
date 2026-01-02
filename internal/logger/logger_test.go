package logger

import (
	"bytes"
	"testing"
)

func TestNoOpLogger_Debugf_DoesNothing(t *testing.T) {
	log := Nop()
	// Should not panic and produce no output
	log.Debugf("test message: %s", "value")
}

func TestNoOpLogger_Warnf_DoesNothing(t *testing.T) {
	log := Nop()
	// Should not panic and produce no output
	log.Warnf("test warning: %s", "value")
}

func TestStdLogger_Verbose_ShowsDebug(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, true) // verbose enabled

	log.Debugf("processing file: %s", "test.yml")

	got := buf.String()
	want := "[DEBUG] processing file: test.yml\n"
	if got != want {
		t.Errorf("Debugf() output = %q, want %q", got, want)
	}
}

func TestStdLogger_Quiet_HidesDebug(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, false) // verbose disabled

	log.Debugf("processing file: %s", "test.yml")

	got := buf.String()
	if got != "" {
		t.Errorf("Debugf() with verbose=false should produce no output, got %q", got)
	}
}

func TestStdLogger_Warnf_AlwaysShown_WhenVerbose(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, true) // verbose enabled

	log.Warnf("no files found in: %s", "empty-dir")

	got := buf.String()
	want := "[WARN] no files found in: empty-dir\n"
	if got != want {
		t.Errorf("Warnf() output = %q, want %q", got, want)
	}
}

func TestStdLogger_Warnf_AlwaysShown_WhenQuiet(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, false) // verbose disabled

	log.Warnf("no files found in: %s", "empty-dir")

	got := buf.String()
	want := "[WARN] no files found in: empty-dir\n"
	if got != want {
		t.Errorf("Warnf() output = %q, want %q", got, want)
	}
}

func TestStdLogger_MultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, true)

	log.Debugf("file 1: %s", "a.yml")
	log.Warnf("warning: %s", "something")
	log.Debugf("file 2: %s", "b.yml")

	got := buf.String()
	want := "[DEBUG] file 1: a.yml\n[WARN] warning: something\n[DEBUG] file 2: b.yml\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

