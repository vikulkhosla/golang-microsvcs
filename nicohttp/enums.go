package nicohttp

import (
	"errors"
)


func (auth authNStrategy) String() string {
	return [...]string{"JWTRSA", "BASIC", "JWTHMAC", "LDAP", "NOAUTH"}[auth]
}


func getAuthStrategy(auth string) (authNStrategy, error) {
	s := map[string]int {"JWTRSA":0, "BASIC":1, "JWTHMAC":2, "LDAP":3, "NOAUTH":4}
	if val, ok := s[auth]; ok {
		return authNStrategy(val), nil
	}
	return -1, errors.New("invalid argument")
}


func (sinkID logSink) String() string {
	return [...]string{"FILE", "STDOUT"}[sinkID]
}


func getLogSink(sink string) (logSink, error) {
	sinks := map[string]int {"FILE":0, "STDOUT":1}
	if val, ok := sinks[sink]; ok {
		return logSink(val), nil
	}
	return -1, errors.New("invalid argument")
}

func (loggerType memoryLoggerType) String() string {
	return [...]string{"MemoryBound", "EntryBound"}[loggerType]
}


func getLoggerType(l string) (memoryLoggerType, error) {
	lt := map[string]int {"MemoryBound":0, "EntryBound":1}
	if val, ok := lt[l]; ok {
		return memoryLoggerType(val), nil
	}
	return -1, errors.New("invalid argument")
}
