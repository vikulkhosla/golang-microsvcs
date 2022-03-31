package nicohttp

import (
	"testing"
	"strings"

)

func TestEnumLogSink1(t *testing.T) {
	sink := FILE
	if (!strings.EqualFold(sink.String(), "FILE")) {
		t.Fail()
	}
}

func TestEnumLogSink2(t *testing.T) {
	expected := FILE
	logSink, err := getLogSink("FILE")
	if err != nil || logSink != expected {
		t.Fail()
	}
}

func TestEnumAuthNStrategy1(t *testing.T) {
	a1 := NOAUTH
	if (!strings.EqualFold(a1.String(), "NOAUTH")) {
		t.Fail()
	}
}

func TestEnumAuthNStrategy2(t *testing.T) {
	expected := NOAUTH
	as, err := getAuthStrategy("NOAUTH")
	if err != nil || as != expected {
		t.Fail()
	}
}

func TestEnumMemoryLoggerType1(t *testing.T) {
	a1 := EntryBound
	if (!strings.EqualFold(a1.String(), "EntryBound")) {
		t.Fail()
	}
}

func TestEnumMemoryLoggerType2(t *testing.T) {
	expected := EntryBound
	as, err := getLoggerType("EntryBound")
	if err != nil || as != expected {
		t.Fail()
	}
}
