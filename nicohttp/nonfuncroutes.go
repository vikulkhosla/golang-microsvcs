package nicohttp

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
	"strconv"
	"fmt"
	"syscall"

	"github.com/gorilla/mux"
)

func configureNonFuncRoutes(b *NicoBuilder) {
	r := b.server.httpRouter
	r.HandleFunc("/healthz", healthz).Methods("GET")
	r.HandleFunc("/suspend", suspend).Methods("POST")
	r.HandleFunc("/suspend", suspendStatus).Methods("GET")
	r.HandleFunc("/restart", restart).Methods("POST")
	r.HandleFunc("/shutdown", shutdown).Methods("POST")
	r.HandleFunc("/api", api).Methods("GET")
	r.HandleFunc("/uptime",getUpTime).Methods("GET")
	r.HandleFunc("/builder",getBuilder).Methods("GET")
	if (!b.disabledMemoryLogs) {
		r.HandleFunc("/logs/head/{entries}", getHead).Methods("GET")
		r.HandleFunc("/logs/tail/{entries}", getTail).Methods("GET")
		r.HandleFunc("/logs/size",getLogSize).Methods("GET")
		r.HandleFunc("/dumplog",dumpLog).Methods("POST")
	}
}


func healthz(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&builder.server.suspended) == 1 {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	if atomic.LoadInt32(&builder.server.healthy) == 1 {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}


func api(w http.ResponseWriter, r *http.Request) {

	apiInherited, apiService, err := generateAPI(builder.server.httpRouter)
	if err == nil {
		m := map[string][]string {"base-service": apiInherited, builder.server.svcName: apiService}
		jsFinal, err3 := json.MarshalIndent(m, "", "\t")
		if (err3 != nil) {
			http.Error(w, err3.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsFinal)
		return

	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}


func restart(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&builder.server.suspended) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer 	atomic.StoreInt32(&builder.server.suspended, 0)
	defer func() {
		dur := time.Since(builder.server.suspendTime)
		builder.server.suspendDuration += dur
	}()
	w.WriteHeader(http.StatusNoContent)
	log.Printf("API Driven restart for service: %s successful \n", builder.server.svcName)
}


func getBuilder(w http.ResponseWriter, r *http.Request) {
	js, err := json.MarshalIndent(builder.props, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func suspendStatus(w http.ResponseWriter, r *http.Request) {
	var map1 map[string]bool

	if atomic.LoadInt32(&builder.server.suspended) == 1 {
		map1 = map[string]bool{"suspended": true}
	} else {
		map1 = map[string]bool{"suspended": false}
	}
	js, err := json.MarshalIndent(map1, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}



func suspend(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&builder.server.suspended) == 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	atomic.StoreInt32(&builder.server.suspended, 1)
	builder.server.suspendTime = time.Now()
	w.WriteHeader(http.StatusNoContent)
	log.Printf("API Driven suspension for service: %s successful \n", builder.server.svcName)
}


func shutdown(w http.ResponseWriter, r *http.Request) {

	log.Printf("API driven shutdown triggered for service: %s: \n", builder.server.svcName)
	time.Sleep(1 * time.Second)
	defer func() {
		w.WriteHeader(http.StatusNoContent)
	}()
	log.Printf("API driven shutdown triggered for service: %s: \n", builder.server.svcName)
	builder.server.server.SetKeepAlivesEnabled(false)
	builder.server.interruptChannel <- syscall.SIGINT
}


func getHead(w http.ResponseWriter, r *http.Request) {

	entries := mux.Vars(r)["entries"]
	nume, err := strconv.Atoi(entries)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return	
	}
	plog := logHead(nume, builder.server) 
	js, err := json.MarshalIndent(plog, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func getTail(w http.ResponseWriter, r *http.Request) {

	entries := mux.Vars(r)["entries"]
	nume, err := strconv.Atoi(entries)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return	
	}
	plog := logTail(nume, builder.server) 
	js, err := json.MarshalIndent(plog, "", "\t")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func getLogSize(w http.ResponseWriter, r *http.Request) {

	max, current, evicted := logSize(builder.server)
	map1 := map[string]int {"max": max, "current": current, "evicted": evicted}
	js, err := json.MarshalIndent(map1, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func getUpTime(w http.ResponseWriter, r *http.Request) {

	t := time.Since(builder.server.startTime)
	h, m, s := decomposeDuration(t)
	up := fmt.Sprintf("H: %d, M: %d, S: %d", h, m, s)

	sut := builder.server.suspendDuration
	if (atomic.LoadInt32(&builder.server.suspended) == 1) {
		sut2 := time.Since(builder.server.suspendTime)
		sut += sut2
	}
	h, m, s = decomposeDuration(sut)
	su := fmt.Sprintf("H: %d, M: %d, S: %d", h, m, s)

	map1 := map[string]string{"uptime": up, "suspended:": su}

	js, err := json.MarshalIndent(map1, "", "\t")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func dumpLog(w http.ResponseWriter, r *http.Request) {
	builder.server.logCmdChan <- dumpLogCmd
	w.WriteHeader(http.StatusNoContent)
}


func decomposeDuration(d time.Duration) (hours, minutes, seconds time.Duration) {
	hours = d/time.Hour
	minutes = (d - (hours * time.Hour))/time.Minute
	seconds = (d - (hours * time.Hour) - (minutes * time.Minute))/time.Second
	return
}
