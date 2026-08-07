package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	mu      sync.Mutex
	file    *os.File
	logPath string
	logger  = log.New(os.Stderr, "", log.LstdFlags)
)

// Init starts an owner-only rolling application log in dataDir. The previous
// log is retained as scrobbler.log.1 when it grows beyond two megabytes.
func Init(dataDir string) error {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return err
	}
	logPath = filepath.Join(dataDir, "scrobbler.log")
	if info, err := os.Stat(logPath); err == nil && info.Size() > 2*1024*1024 {
		_ = os.Remove(logPath + ".1")
		_ = os.Rename(logPath, logPath+".1")
	}
	opened, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	file = opened
	logger = log.New(opened, "", log.LstdFlags|log.Lmicroseconds)
	logger.Printf("application log started")
	return nil
}

func Path() string {
	mu.Lock()
	defer mu.Unlock()
	return logPath
}

func Printf(format string, args ...any) {
	mu.Lock()
	defer mu.Unlock()
	logger.Printf(format, args...)
}

func Event(name string, fields map[string]any) {
	mu.Lock()
	defer mu.Unlock()
	logger.Printf("event=%s time=%s fields=%s", name, time.Now().UTC().Format(time.RFC3339), formatFields(fields))
}

func formatFields(fields map[string]any) string {
	result := ""
	for key, value := range fields {
		if result != "" {
			result += " "
		}
		result += fmt.Sprintf("%s=%v", key, value)
	}
	return result
}

func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		_ = file.Close()
		file = nil
	}
}
