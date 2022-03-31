package main

import (
	"github.com/gorilla/mux"
)


// APIExtension ...
type APIExtension interface {
	ERoute(*mux.Router) (routeName string)
}