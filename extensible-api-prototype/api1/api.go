package main

import (
	"net/http"
	"github.com/gorilla/mux"
)

type extender string

func (e extender) ERoute(r *mux.Router) string {
	r.HandleFunc("/foo", api1Handler2).Methods("Get").Name("GetFoo")
	return "GetFoo"
}

func api1Handler2(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("foo"))
}

// APIExtension ...
var APIExtension extender

