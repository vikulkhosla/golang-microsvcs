package nicohttp

import (
	"net/http"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
	"context"

	"github.com/gorilla/mux"
)

const (
	defaultHandlerTimeout time.Duration = 300 * time.Second
	defaultRateLimit int = 500
	defaultShutdownWait time.Duration = 120 * time.Second
	defaultLogFileDir string = "."
)

//NicoServer - constructed HTTP Server with required optionality
type NicoServer struct {
	svcName string
	port	uint32
	server *http.Server
	builder *NicoBuilder
	handlerTimeout time.Duration
	shutdownWait time.Duration
	logQoS int
	maxLogEntries int

	healthy    		int32
	suspended    	int32
	httpRouter 		*mux.Router
	interruptChannel chan os.Signal
	startTime time.Time
	suspendTime time.Time
	suspendDuration time.Duration
	logChanReceivers sync.WaitGroup

	nextLogID      int
	evictedLogSize int
	memLogSize     int
	memLog         []memoryLogEntry
	logChan        chan string
	logChanState	uint32
	logCmdChan		chan string
	snapshotID     int

	sink logSink
}

// Builder - access the builder used
func (h *NicoServer) Builder() (*NicoBuilder) {
	return h.builder
}

// Start - the TN Http server
func (h *NicoServer) Start() {

	if (!h.builder.disabledMemoryLogs) {
		h.logChan = make(chan string)
		h.logCmdChan = make (chan string)
		h.logChanState = 1
		h.logChanReceivers.Add(1)
		h.memLogSize = h.logQoS
		h.memLog = make([]memoryLogEntry, h.memLogSize, h.memLogSize)
		go func() {
			defer h.logChanReceivers.Done()
			entryBoundMemoryLogger(h)
		}()
	}

	go func() {
		if err := h.server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	h.startTime = time.Now()
	atomic.StoreInt32(&h.healthy, 1)
	atomic.StoreInt32(&h.suspended, 0)

	fmt.Printf("Service %s started at port %d, start time: %s\n", h.svcName, h.port, h.startTime )
	h.interruptChannel = make(chan os.Signal, 1)
	signal.Notify(h.interruptChannel, os.Interrupt)

	<-h.interruptChannel
	log.Printf("Service %s shutting down\n", h.svcName)

	if (!h.builder.disabledMemoryLogs) {
		h.logChanState = 0
		close(h.logChan)
		h.logChanReceivers.Wait()
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.shutdownWait * time.Second)
	defer cancel()
	h.server.Shutdown(ctx)
//	os.Exit(0)
}

// Stop - the TN Http server
func (h *NicoServer) Stop() {
	log.Printf("Service %s being stopped.\n", h.svcName)

	if (!h.builder.disabledMemoryLogs) {
		h.logChanState = 0
		close(h.logChan)
		h.logChanReceivers.Wait()
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.shutdownWait * time.Second)
	defer cancel()
	h.server.Shutdown(ctx)
	log.SetOutput(os.Stdout)
}


// Mux - get the Http mux
func (h *NicoServer) Mux() (*mux.Router) {
	return h.httpRouter
}


// Service - returns the name of the Service used in Builder.Create() call
func (h *NicoServer) Service() (string) {
	return h.svcName
}


// Port - returns the port number used in Builder.Create() call
func (h *NicoServer) Port() (uint32) {
	return h.port
}
