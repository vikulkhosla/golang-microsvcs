package nicohttp

import (
	"sync/atomic"
	"testing"
	"time"
	"net/http"
)


var builderFirstPort = uint32(50100)

func builderNextPort() uint32 {
    return atomic.AddUint32(&builderFirstPort, 1)
}

func TestDefaults(t *testing.T) {
	b := GetBuilder().WithDefaults()
	
	actual := b.Props()
	expected := map[string]interface{} {
		CustomPostMediatorKey : "None",
		CustomPreMediatorKey : "None",
		LogSinkKey : "STDOUT",
		MemoryLoggerQoSKey : 5000,
		AuthStrategyKey : "NOAUTH",
		HandlerTimeoutKey : time.Duration(60),
		ListenPortKey : 8080,
		LogFileDirKey : ".",
		MemoryLoggerTypeKey : "EntryBound",
		RateLimitKey : 100,
		ShutdownWaitKey : time.Duration(60),
	}
	for ek, ev := range expected {
		if av, ok := actual[ek]; !ok || ev != av {
			t.Fatalf("%s: Key = %s, expected value = %t, actual value = %d", t.Name(), ek, ev, av)
		}
	}
}


func TestWithCustomMediators(t *testing.T) {
	f := func (next http.Handler) http.Handler {
			return nil
	}
	b := GetBuilder().WithDefaults().WithCustomPreMediator("F1", f).WithCustomPostMediator("F1", f)
	
	actual := b.Props()
	expected := map[string]interface{} {
		CustomPostMediatorKey : "F1",
		CustomPreMediatorKey : "F1",
		LogSinkKey : "STDOUT",
		MemoryLoggerQoSKey : 5000,
		AuthStrategyKey : "NOAUTH",
		HandlerTimeoutKey : time.Duration(60),
		ListenPortKey : 8080,
		LogFileDirKey : ".",
		MemoryLoggerTypeKey : "EntryBound",
		RateLimitKey : 100,
		ShutdownWaitKey : time.Duration(60),
	}
	for ek, ev := range expected {
		if av, ok := actual[ek]; !ok || ev != av {
			t.Fatalf("%s: Key = %s, expected value = %t, actual value = %d", t.Name(), ek, ev, av)
		}
	}
}


func TestWithNoMemoryLogger(t *testing.T) {
	b := GetBuilder().WithDefaults().WithNoMemoryLogger()
	
	actual := b.Props()
	expected := map[string]interface{} {
		CustomPostMediatorKey : "None",
		CustomPreMediatorKey : "None",
		LogSinkKey : "None",
		MemoryLoggerQoSKey : 0,
		AuthStrategyKey : "NOAUTH",
		HandlerTimeoutKey : time.Duration(60),
		ListenPortKey : 8080,
		LogFileDirKey : ".",
		MemoryLoggerTypeKey : "None",
		RateLimitKey : 100,
		ShutdownWaitKey : time.Duration(60),
	}
	for ek, ev := range expected {
		if av, ok := actual[ek]; !ok || ev != av {
			t.Fatalf("%s: Key = %s, expected value = %t, actual value = %d", t.Name(), ek, ev, av)
		}
	}
}
