package nicohttp

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"encoding/json"
)

const (
	uriLogSize string = "/logs/size"
	uriShutdown string = "/shutdown"
	uriAPI string = "/api"
	uriHealthz string = "/healthz"
	uriSuspend string = "/suspend"
	uriRestart string = "/restart"
	uriBuilder string = "/builder"
	
	applicationJSON = "application/json"
)

func getTarget(port uint32, uri string) string {
	return fmt.Sprintf("http://localhost:%d%s", port, uri)
}

func verifyContentType(resp *http.Response, t string) bool {
	if (strings.EqualFold(resp.Header.Get("Content-Type"), t)) {
		return true
	}
	return false
}

func verifyJSONIntKey(body []byte, field string, value int) (bool, error) {
	var m map[string]int
	err := json.Unmarshal(body, &m)
	if err != nil {
		return false, err
	}
	if m[field] == value {
		fmt.Println("Json field is equal")
		return true, nil
	}
	return false, fmt.Errorf("JSON field %s is %v and not %d", field, m[field], value )
}


func asyncStart(wg *sync.WaitGroup, srv *NicoServer) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		srv.Start()
	}()
}

