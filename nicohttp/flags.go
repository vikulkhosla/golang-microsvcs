package nicohttp

import (
	"flag"
	"time"
	"os"
	"fmt"
	"strings"
)

// Required Base args
var (
	requiredBaseArgs int
	argSvcName      *string
)

// Optional Base args
var (
	argListenPort     *int
	argHandlerTimeout time.Duration
	argRateLimit      *int
	argShutdownWait   time.Duration
	argLogFileDir   *string
	argAuthStrategy   *string
	argLogSink		*string
	argMemoryLogsEnabled *bool
	argMemoryLogType *string
)

var (
	flagset    = make(map[string]bool)
)


func initBaseFlags() {
	argSvcName = flag.String("serviceName", "", "[REQUIRED] name of the micro service")
	argListenPort = flag.Int("listenPort", 8080, "[OPTIONAL] HTTP Server listen port")
	flag.DurationVar(&argHandlerTimeout, "handlerTimeout",60*time.Second, "[OPTIONAL] handlerTimeout in seconds")
	argRateLimit = flag.Int("rateLimit", 60 , "[OPTIONAL] rate limit - requests per minute")
	flag.DurationVar(&argShutdownWait, "shutdownTimeout", 60*time.Second, "[OPTIONAL] graceful shutdown timeout in seconds")
	argAuthStrategy = flag.String("authStrategy", "NONE", "[OPTIONAL] JWT for JWT verification, NONE for no authentication")
	argLogFileDir = flag.String("logFileDir", ".", "[OPTIONAL] Directory where log file will be written. Log file is <service-name>.log")
	argLogSink = flag.String("logSink", ".", "[OPTIONAL] Log Sink can be File or Stdout. Default is File")
	argMemoryLogsEnabled = flag.Bool("memoryLogEnabled", true, "[OPTIONAL] Enable memory logs. Default is true")
	argMemoryLogType = flag.String("memoryLogType", ".", "[OPTIONAL] Either EntryBound or MemoryBound. Default is EntryBound")
}

func validateRequiredArgs() {
	if flag.NFlag() < (builder.extendedRequiredFlags + requiredBaseArgs) {
		fmt.Printf("Required args = %d, provided args = %d\n\n", (builder.extendedRequiredFlags + requiredBaseArgs), flag.NFlag())
		flag.Usage()
		os.Exit(1)
	}

	if *argSvcName == "" {
		fmt.Printf("Service name is required\n\n")
		flag.Usage()
		os.Exit(1)
	}
}

func validateBaseArgs() {
	if flagset["logFileDir"] {
		if strings.HasSuffix(*argLogFileDir, "/") {
			panic(fmt.Sprintf("logFileDir %s cannot end with a trailing /", *argLogFileDir))
		}
	}
	if _, err := os.Stat(*argLogFileDir); err != nil {
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("Log Directory %s does not exist", *argLogFileDir))
		}
	}
	if flagset["logSink"] {
		if _, err := getLogSink(*argLogSink); err != nil {
			panic(fmt.Sprintf("Invalid log sink: %s", *argLogSink))
		}
	}
}
