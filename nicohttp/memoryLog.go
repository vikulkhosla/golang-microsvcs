package nicohttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
	"strings"
)

type memoryLogEntry struct {
	ID int
	TS int64
	LE string
}

const (
	defaultMemLogSize int = 5000
	dumpLogCmd string = "_DUMPLOG_"
	defaultLogChannelSleep time.Duration = 50 * time.Millisecond
)


func entryBoundMemoryLogger(server *NicoServer) {
	fmt.Printf("Starting entry bound memory logger ..... max entries = %d\n", server.memLogSize)

	quitLogger:

	for {
		select {
			case msg, open := <-server.logChan :
				if open {
					le := memoryLogEntry{ID: 0, TS: time.Now().UnixNano(), LE: msg}
					appendLogEntry(server, &le)
					continue
				}
				break quitLogger
			case msg, open := <-server.logCmdChan :
				if open {
					if strings.EqualFold(msg, dumpLogCmd) {
						server.snapshotID++
						n, err := dumpMemoryLog(server)
						server.evictedLogSize += server.nextLogID
						s := fmt.Sprintf("API Driven memory log dump: snapshotID=%d, entries=%d, bytesWritten=%d, error=%s", 
													server.snapshotID, server.nextLogID, n, err)
						server.nextLogID = 0
						e := memoryLogEntry{ID: server.nextLogID, TS: time.Now().UnixNano(), LE: s}
						server.memLog = make([]memoryLogEntry, server.memLogSize, server.memLogSize)
						server.memLog[server.nextLogID] = e
						server.nextLogID++					
					}
					continue
				}
				break quitLogger
			default:
				if len(server.logChan) == 0 && len(server.logCmdChan) == 0 {
					time.Sleep(defaultLogChannelSleep)
				}
		}
	}
	fmt.Printf("Closing memory log channel for service %s ..... \n", server.svcName)
}


func appendLogEntry(server *NicoServer, le *memoryLogEntry) {
	if server.nextLogID == cap(server.memLog) {
		server.snapshotID++
		n, err := dumpMemoryLog(server)
		server.evictedLogSize += server.memLogSize
		s := fmt.Sprintf("Dumped memory log snapshot to disk: snapshotID=%d, entries=%d, bytesWritten=%d, error=%s", 
													server.snapshotID, server.nextLogID, n, err)
		server.nextLogID = 0
		e := memoryLogEntry{ID: server.nextLogID, TS: time.Now().UnixNano(), LE: s}
		server.memLog = make([]memoryLogEntry, server.memLogSize, server.memLogSize)
		server.memLog[server.nextLogID] = e
		server.nextLogID++
	}
	le.ID = server.nextLogID
	server.memLog[server.nextLogID] = *le
	server.nextLogID++
}

func logHead(size int, server *NicoServer) []memoryLogEntry {
	if size > server.nextLogID {
		return server.memLog[0:server.nextLogID]
	}
	return server.memLog[0:size]
}

func logTail(size int, server *NicoServer) []memoryLogEntry {
	if size >= server.nextLogID {
		return logHead(size, server)
	}
	return server.memLog[server.nextLogID-size : server.nextLogID]
}

func logSize(server *NicoServer) (maxsize int, current int, evicted int) {
	return len(server.memLog), server.nextLogID, server.evictedLogSize
}

type logWriter struct {
	existing io.Writer
}

func newLogWriter(e io.Writer) io.Writer {
	l := logWriter{existing: e}
	return &l
}

func (lw logWriter) Write(p []byte) (n int, err error) {
	n, e := lw.existing.Write(p)
	if (builder.server.logChanState == 1) {
		builder.server.logChan <- string(p)
	}
	return n, e
}

func dumpMemoryLog(server *NicoServer) (bytesWritten int, err error) {
	fmt.Println("Dumping memory log ......")
	server.snapshotID++
	switch (server.sink) {
		case FILE:
			var dir string
			if flagset["logFileDir"] {
				dir = *argLogFileDir
			} else {
				dir = "."
			}
			f := fmt.Sprintf("%s/%s.log.%d", dir, server.svcName, server.snapshotID)
			return dumpMemoryLogToFile(f, server)
		case STDOUT:
			return dumpMemoryLogToStdout(server)
		default:
			fmt.Println("Unsupported log sink ......")
			return 0, errors.New("Unsupported log sink")
	}
}


func dumpMemoryLogToFile(file string, server *NicoServer) (bytesWritten int, err error) {
	fmt.Println("Dumping memory log ...... to FILE Sink")
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	js, err := json.MarshalIndent(server.memLog[0:server.nextLogID], "", "\t")

	if err != nil {
		return 0, err
	}
	n, err := f.Write(js)
	return n, err
}


func dumpMemoryLogToStdout(server *NicoServer) (bytesWritten int, err error) {
	fmt.Println("Dumping memory log ...... to STDOUT Sink")

	js, err := json.MarshalIndent(server.memLog[0:server.nextLogID], "", "\t")

	if err != nil {
		return 0, err
	}
	n, err := os.Stdout.Write(js)
	return n, err
}
