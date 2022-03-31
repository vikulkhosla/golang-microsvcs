package nicohttp

import (
	"sync/atomic"
	"testing"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"reflect"
	"strings"
)


var loggerFirstPort = uint32(50000)

func getLoggerPort() uint32 {
	return atomic.AddUint32(&loggerFirstPort, 1)
}



func TestGetLogSize0(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Get(getTarget(p, uriLogSize))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if !verifyContentType(resp, applicationJSON) {
		t.Fatalf("%s: Content-Type is not %s", t.Name(), applicationJSON)
	}
	if b, err := verifyJSONIntKey(body, "current", 0); !b || err != nil {
		t.Fatalf("%s: %s", t.Name(), err)
	}
}


func TestGetLogSize1(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	http.Get(getTarget(p, uriAPI))
	resp, err := http.Get(getTarget(p, uriLogSize))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	fmt.Println(string(body))
	if !verifyContentType(resp, applicationJSON) {
		t.Fatalf("%s: Content-Type is not %s", t.Name(), applicationJSON)
	}
	if b, err := verifyJSONIntKey(body, "current", 1); !b || err != nil {
		t.Fatalf("%s: response JSON field %s is not %d", t.Name(), "current", 0)
	}
}


func TestGenerateAPI1(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Get(getTarget(p, uriAPI))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if !verifyContentType(resp, applicationJSON) {
		t.Fatalf("%s: Content-Type is not %s", t.Name(), applicationJSON)
	}
	var m map[string][]string
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	expected := []string{"GET   /healthz", "POST   /suspend", "GET   /suspend", "POST   /restart", 
						"POST   /shutdown", "GET   /api", "GET   /uptime", "GET   /builder",
						"GET   /logs/head/{entries}", "GET   /logs/tail/{entries}", "GET   /logs/size",
						"POST   /dumplog",
					}
	actual, ok := m["base-service"]
	if !ok {
		t.Fatalf("%s: JSON response has no base-service field", t.Name())
	}
	if (len(actual) != len(expected)) {
		t.Fatalf("%s: Expected number of API entries = %d, actual = %d", t.Name(), len(expected), len(actual))
	}
	fmt.Println(actual)
	
	expectedMap := make(map[string][]string)
	actualMap := make(map[string][]string)

	for _, s := range expected {
		tokens := strings.Fields(s)
		if v, ok := expectedMap[tokens[0]]; !ok {
			expectedMap[tokens[0]] = []string{tokens[1]}
		} else {
			expectedMap[tokens[0]] = append(v, tokens[1])
		}
	}
	for _, s := range actual {
		tokens := strings.Fields(s)
		if v, ok := actualMap[tokens[0]]; !ok {
			actualMap[tokens[0]] = []string{tokens[1]}
		} else {
			actualMap[tokens[0]] = append(v, tokens[1])
		}
	}
	for ek, ev := range expectedMap {
		if av, ok := actualMap[ek]; !ok || !reflect.DeepEqual(ev,av) {
			t.Fatalf("%s: Key = %s, expected value = %v, actual value = %v", t.Name(), ek, ev, av)
		}
	}

}


func TestHealthz1(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Get(getTarget(p, uriHealthz))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s: Response is not %d", t.Name(), http.StatusOK)
	}
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
}

func TestHealthzAfterSuspend(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Post(getTarget(p, uriSuspend), "", nil)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	resp, err = http.Get(getTarget(p, uriHealthz))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("%s: Response is not %d", t.Name(), http.StatusServiceUnavailable)
	}
}


func TestHealthzFirdtSuspendThenRestart(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Post(getTarget(p, uriSuspend), "", nil)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	resp, err = http.Post(getTarget(p, uriRestart), "", nil)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	resp, err = http.Get(getTarget(p, uriHealthz))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s: Response is not %d", t.Name(), http.StatusOK)
	}
}


func TestGetBuilder1(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Get(getTarget(p, uriBuilder))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if !verifyContentType(resp, applicationJSON) {
		t.Fatalf("%s: Content-Type is not %s", t.Name(), applicationJSON)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	var actual map[string]interface{}
	err = json.Unmarshal(body, &actual)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	expected := map[string]interface{} {
		CustomPostMediatorKey : "None",
		CustomPreMediatorKey : "None",
		LogSinkKey : "STDOUT",
		MemoryLoggerQoSKey : float64(5000),
		AuthStrategyKey : "NOAUTH",
		HandlerTimeoutKey : float64(60),
		ListenPortKey : float64(p),
		LogFileDirKey : ".",
		MemoryLoggerTypeKey : "EntryBound",
		RateLimitKey : float64(100),
		ShutdownWaitKey : float64(60),
	}
	for ek, ev := range expected {
		if av, ok := actual[ek]; !ok || ev != av {
			t.Fatalf("%s: Key = %s, expected value = %t, actual value = %d", t.Name(), ek, ev, av)
		}
	}
}

func TestGenerateAPINoMemoryLogging(t *testing.T) {
	p := getLoggerPort()
	srv, err := GetBuilder().WithDefaults().WithNoMemoryLogger().Create(t.Name(), p)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer srv.Stop()
	go srv.Start()
	resp, err := http.Get(getTarget(p, uriAPI))
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	if !verifyContentType(resp, applicationJSON) {
		t.Fatalf("%s: Content-Type is not %s", t.Name(), applicationJSON)
	}
	var m map[string][]string
	err = json.Unmarshal(body, &m)
	if err != nil {
		t.Fatalf("%s: %t", t.Name(), err)
	}
	expected := []string{"GET   /healthz", "POST   /suspend", "GET   /suspend", "POST   /restart", 
						"POST   /shutdown", "GET   /api", "GET   /uptime", "GET   /builder",
					}
	actual, ok := m["base-service"]
	if !ok {
		t.Fatalf("%s: JSON response has no base-service field", t.Name())
	}
	if (len(actual) != len(expected)) {
		t.Fatalf("%s: Expected number of API entries = %d, actual = %d", t.Name(), len(expected), len(actual))
	}
	fmt.Println(actual)
	
	expectedMap := make(map[string][]string)
	actualMap := make(map[string][]string)

	for _, s := range expected {
		tokens := strings.Fields(s)
		if v, ok := expectedMap[tokens[0]]; !ok {
			expectedMap[tokens[0]] = []string{tokens[1]}
		} else {
			expectedMap[tokens[0]] = append(v, tokens[1])
		}
	}
	for _, s := range actual {
		tokens := strings.Fields(s)
		if v, ok := actualMap[tokens[0]]; !ok {
			actualMap[tokens[0]] = []string{tokens[1]}
		} else {
			actualMap[tokens[0]] = append(v, tokens[1])
		}
	}
	for ek, ev := range expectedMap {
		if av, ok := actualMap[ek]; !ok || !reflect.DeepEqual(ev,av) {
			t.Fatalf("%s: Key = %s, expected value = %v, actual value = %v", t.Name(), ek, ev, av)
		}
	}

}
