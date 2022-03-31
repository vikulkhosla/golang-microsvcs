package nicohttp

import (
	"sync"
	"net/http"
	"log"
	"os"
	"fmt"
	"time"
	"flag"
	"strconv"

	"github.com/gorilla/mux"
)

var (
	builder *NicoBuilder
	mutex = &sync.Mutex{}
	handlerChain = [8]func(http.Handler) http.Handler{}
)

const (
	// ListenPortKey ...
	ListenPortKey string = "listenPort"
	// HandlerTimeoutKey ...
	HandlerTimeoutKey string = "handlerTimeout (secs)"
	// RateLimitKey ...
	RateLimitKey string = "rateLimit (per min)"
	// ShutdownWaitKey ...
	ShutdownWaitKey string = "shutdownWait (secs)"
	// AuthStrategyKey ...
	AuthStrategyKey string = "authStrategy"
	// LogFileDirKey ...
	LogFileDirKey string = "logFileDir"
	// LogSinkKey ...
	LogSinkKey string = "LogSink"
	// MemoryLoggerTypeKey  ...
	MemoryLoggerTypeKey string = "memoryLoggerType"
	// EnableMemoryLoggerKey  ...
	EnableMemoryLoggerKey string = "EnableMemoryLogger"
	// CustomPreMediatorKey  ...
	CustomPreMediatorKey string = "CustomPreMediator"
	// CustomPostMediatorKey  ...
	CustomPostMediatorKey string = "CustomPostMediator"
	// MemoryLoggerQoSKey ...
	MemoryLoggerQoSKey string = "MemoryLoggerQoS"
)

type  authNStrategy int
const (
	// JWTRSA Enforcement
	JWTRSA   authNStrategy = iota
	//BASIC - HTTP Basic Auth Enforcement
	BASIC
	// JWTHMAC Enforcement
	JWTHMAC
	// LDAP - LDAP enforcement
	LDAP
	// NOAUTH - No Enforcement
	NOAUTH 
 )

 type  memoryLoggerType int
const (
	// MemoryBound ...
	MemoryBound   memoryLoggerType = iota
	// EntryBound ...
	EntryBound
 )

type logSink int
const (
	//FILE sink
	FILE logSink = iota
	//STDOUT sink
	STDOUT
)


//TNBuilder - classic Builder Pattern implementation to allow customizing the build of a Nico Http Server
type NicoBuilder struct {
	server *NicoServer
	props map[string]interface{}
	baseFlags bool
	extendedFlags int
	extendedRequiredFlags int
	disabledMemoryLogs bool
}



// GetBuilder - returns a new Builder or existing Builder. 
func GetBuilder() (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	builder = &NicoBuilder{}
	builder.props = defaultProps()
	builder.server = &NicoServer{}
	builder.server.builder = builder
	builder.disabledMemoryLogs = false
	return builder
}


// Props - get build optionality exercised
func (b *NicoBuilder) Props() (map[string]interface{}) {
	return b.props
}

// WithBaseFlags - turn on base flags
func (b *NicoBuilder) WithBaseFlags() (*NicoBuilder) {
	b.baseFlags = true
   return b
}


// WithStringFlag - additional flag
func (b *NicoBuilder) WithStringFlag(arg, defaultVal, description string, required bool) (*NicoBuilder, *string) {
	 a := flag.String(arg, defaultVal, description)
	 if (required) {
	 	b.extendedRequiredFlags++
	 }
	 b.extendedFlags++
	return b,a
}


// WithIntFlag - additional flag
func (b *NicoBuilder) WithIntFlag(arg string, defaultVal int, description string, required bool) (*NicoBuilder, *int) {
	a := flag.Int(arg, defaultVal, description)
	if (required) {
		b.extendedRequiredFlags++
	}
	b.extendedFlags++
   return b,a
}


// WithDurationFlag - additional flag
func (b *NicoBuilder) WithDurationFlag(argptr *time.Duration, arg string, defaultVal time.Duration, description string, required bool) (*NicoBuilder) {
	flag.DurationVar(argptr, arg, defaultVal, description)
	if (required) {
		b.extendedRequiredFlags++
	}
	b.extendedFlags++
   return b
}


// WithBoolFlag - additional flag
func (b *NicoBuilder) WithBoolFlag(arg string, defaultVal bool, description string, required bool) (*NicoBuilder, *bool) {
	a := flag.Bool(arg, defaultVal, description)
	if (required) {
		b.extendedRequiredFlags++
	}
	b.extendedFlags++
   return b,a
}


// WithProperties - require custom HTTPServer to support memory based logs
// accessible through REST API
func (b *NicoBuilder) WithProperties(m map[string]interface{}) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()

	for allowedKey := range b.props {
		if v, ok := m[allowedKey]; ok {
			b.props[allowedKey] = v
		}
	}
	initDefaultHandlerChain()
	return b
}


// WithDefaults - require custom HTTPServer to support memory based logs
// accessible through REST API
func (b *NicoBuilder) WithDefaults() (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	initDefaultHandlerChain()
	return b
}


// WithMemoryLogger - require custom HTTPServer to support memory based logs
// accessible through REST API
func (b *NicoBuilder) WithMemoryLogger(lt memoryLoggerType, size int) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	handlerChain[memoryLoggerMediatorPos]= memoryPostLoggingMediator
	b.props[MemoryLoggerTypeKey] = lt
	b.props[MemoryLoggerQoSKey] = size
	return b
}


// WithNoMemoryLogger - require custom HTTPServer to not support memory based logs
func (b *NicoBuilder) WithNoMemoryLogger() (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	handlerChain[memoryLoggerMediatorPos]= noopHandler
	b.props[MemoryLoggerTypeKey] = "None"
	b.props[LogSinkKey] = "None"
	b.props[MemoryLoggerQoSKey] = 0
	b.disabledMemoryLogs = true
	return b
}


// WithTimeoutHandler - require custom HTTPServer to timeout request if upstream
// handlers not responsive. Return 503
func (b *NicoBuilder) WithTimeoutHandler(d time.Duration) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	handlerChain[timeoutHandlerPos] = timeoutMediator
	b.props[HandlerTimeoutKey] = d
	return b
}


// WithTracing - require custom HTTPServer to introduce a unique RequestID for all HTTP calls
func (b *NicoBuilder) WithTracing() (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	handlerChain[tracingMediatorPos] = tracingMediator
	return b
}


// WithAuthNMediator - require custom HTTP Server to support mediated authentication for
// all URI, based on the provided authentication scheme. Some authn schemes will require
// a config as a Json object
func (b *NicoBuilder) WithAuthNMediator(strategy authNStrategy, config string) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	b.props[AuthStrategyKey] = strategy
	switch (strategy) {
		case JWTRSA :
			handlerChain[authStrategyMediatorPos] = rsaJWTMediator
		case JWTHMAC :
			handlerChain[authStrategyMediatorPos] = hmacJWTMediator
		case LDAP :
			handlerChain[authStrategyMediatorPos] = ldapMediator
		case BASIC :
			handlerChain[authStrategyMediatorPos] = httpBasicAuthMediator
		case NOAUTH :
			handlerChain[authStrategyMediatorPos] = noAuthMediator
		default: 
			panic(fmt.Sprintf("Unsupported auth strategy %s\n", string(authNStrategy(strategy))))
	}
	return b
}


// WithCustomPreMediator - require custom HTTP Server to inject custom HTTP Handler as the first
// handler in the handler chain
func (b *NicoBuilder) WithCustomPreMediator(name string, f func(next http.Handler) http.Handler) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	handlerChain[customPreMediatorPos] = f
	b.props[CustomPreMediatorKey] = name
	return b
}


// WithCustomPostMediator - require custom HTTP Server to inject custom HTTP Handler as the last
// handler in the handler chain
func (b *NicoBuilder) WithCustomPostMediator(name string, f func(next http.Handler) http.Handler) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	handlerChain[customPostMediatorPos] = f
	b.props[CustomPostMediatorKey] = name
	return b
}

// WithLogSink - use specified log sink for batch writes on memory overflow
func (b *NicoBuilder) WithLogSink(sink logSink) (*NicoBuilder) {
	defer mutex.Unlock()
	mutex.Lock()
	b.props[LogSinkKey] = sink.String()
	return b
}


// Create the custom HTTPServer after all build optionality has been specified
func (b *NicoBuilder) Create(svcName string, port uint32) (*NicoServer, error) {
	fmt.Printf("Creating nicoHttp Server .......\n\n")
	defer mutex.Unlock()
	mutex.Lock()
	if b.baseFlags {
		requiredBaseArgs = 1 //svcName
		initBaseFlags()
	}
	if (b.extendedFlags > 0) || b.baseFlags {
		flag.Parse()
		flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })
		validateRequiredArgs()
		validateBaseArgs()
		//updateBuilderProperties()
	}

	b.server.httpRouter = mux.NewRouter()
	configureNonFuncRoutes(b)

	/* inject memory logger for regular log output, mux logging already intercepted */
	if (!b.disabledMemoryLogs) {
		mlw := newLogWriter (os.Stdout)
		log.SetOutput(mlw)
	}
	if flag.Lookup("listenPort") != nil {
		p, err := strconv.Atoi(flag.Lookup("listenPort").Value.String())
		if err != nil {
			panic(err)
		}
		port = uint32(p)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	s := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      rootHandler((b.server.httpRouter)),
	}

	initBuiltServer(svcName, port, b, s)
	return b.server, nil
}

/**************** Supporting Non public methods and constants **********************/

func noopHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func rootHandler(next http.Handler) http.Handler {
	return handlerChain[0](handlerChain[1](handlerChain[2](handlerChain[3](handlerChain[4](handlerChain[5](handlerChain[6](
		(next))))))))
}



func defaultProps() (map[string]interface{}) {
	m := make(map[string]interface{})
	m[ListenPortKey] = 8080
	m[HandlerTimeoutKey] = defaultHandlerTimeout / time.Second
	m[RateLimitKey] = defaultRateLimit
	m[ShutdownWaitKey] = defaultShutdownWait / time.Second
	m[AuthStrategyKey] = NOAUTH.String()
	m[LogFileDirKey] = defaultLogFileDir
	m[LogSinkKey] = STDOUT.String()
	m[MemoryLoggerTypeKey] = EntryBound.String()
	m[CustomPreMediatorKey] = "None"
	m[CustomPostMediatorKey] = "None"
	m[MemoryLoggerQoSKey] = defaultMemLogSize

	return m
}

const (
	customPostMediatorPos int = iota
	suspendMediatorPos
	memoryLoggerMediatorPos
	tracingMediatorPos
	authStrategyMediatorPos
	customAuthorizerPos
	timeoutHandlerPos
	customPreMediatorPos
)

func initDefaultHandlerChain() {
	handlerChain[customPostMediatorPos] = noopHandler /* custom post mediator */
	handlerChain[suspendMediatorPos]= suspendMediator
	handlerChain[memoryLoggerMediatorPos]= memoryPostLoggingMediator
	handlerChain[tracingMediatorPos] = tracingMediator
	handlerChain[authStrategyMediatorPos] = noAuthMediator
	handlerChain[customAuthorizerPos] = noopHandler /* custom authorizer */
//	handlerChain[timeoutHandlerPos] = timeoutMediator
	handlerChain[timeoutHandlerPos] = noopHandler

	handlerChain[customPreMediatorPos] = noopHandler /* custom pre mediator */
}

func initBuiltServer(svcName string, port uint32, b *NicoBuilder, s *http.Server) {
	b.props[ListenPortKey] = port
	b.server.server = s

	//init built server
	b.server.svcName = svcName
	b.server.port = port

	b.server.handlerTimeout = (b.props[HandlerTimeoutKey]).(time.Duration)
	b.server.shutdownWait = (b.props[ShutdownWaitKey]).(time.Duration)

	b.server.sink, _ = getLogSink((b.props[LogSinkKey]).(string))
	b.server.logQoS = (b.props[MemoryLoggerQoSKey]).(int)
}
