package main

import (
	"sync"
	"net/http"

	"github.com/gorilla/mux"
)

type routerSwapper struct {
	mu sync.Mutex
   root *mux.Router
}

func (rs *routerSwapper) Swap(newRouter *mux.Router) {
	rs.mu.Lock()
	rs.root = newRouter
	rs.mu.Unlock()
}

func (rs *routerSwapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   rs.mu.Lock()
   root := rs.root
  rs.mu.Unlock()
  root.ServeHTTP(w, r)
}