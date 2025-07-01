package utils

import "log"

type Log struct {
	Level int
}

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelWarning
	LogLevelError
)

func (l *Log) Debug(v any) {
	if l.Level <= LogLevelDebug {
		log.Printf("[DEBUG] %v\n", v)
	}
}

func (l *Log) Info(v any) {
	if l.Level <= LogLevelInfo {
		log.Printf("[INFO] %v\n", v)
	}
}

func (l *Log) Warning(v any) {
	if l.Level <= LogLevelWarning {
		log.Printf("[WARNING] %v\n", v)
	}
}

func (l *Log) Error(v any) {
	if l.Level <= LogLevelError {
		log.Printf("[ERROR] %v\n", v)
	}
}

func (l *Log) Fatal(v any) {
	log.Fatalf("[ERROR] %v\n", v)
}
