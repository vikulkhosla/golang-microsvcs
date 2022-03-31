package main

import (
	"net/http"
	"github.com/gorilla/mux"
)

type extender string

func (e extender) ERoute(r *mux.Router) string {
	r.HandleFunc("/bar", api1Handler1).Methods("Get").Name("GetBar")
	return "GetBar"
}

func api1Handler1(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("bar"))
}

// APIExtension ...
var APIExtension extender
