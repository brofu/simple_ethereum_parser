package logging

import (
	"log"
	"os"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LeveLSilent
)

type Logger interface {
	SetLogLevel(LogLevel)
	Errorf(string, ...any)
	Warnf(string, ...any)
	Infof(string, ...any)
	Debugf(string, ...any)
}

// defaultLogger. A wrapper of the built-in log
// An implementation of Interface `Logger`
type defaultLogger struct {
	logger *log.Logger
	level  LogLevel
}

func NewDefaultLogger(level LogLevel) *defaultLogger {
	log := &defaultLogger{
		logger: log.Default(),
		level:  level,
	}
	return log
}

func (this *defaultLogger) SetLogLevel(level LogLevel) {
	if level < LevelDebug {
		level = LevelDebug
	}
	if level > LevelError {
		level = LevelError
	}
	this.level = level
}
func (this *defaultLogger) Errorf(format string, v ...any) {
	if this.level <= LevelError {
		this.logger.Printf("ERROR|"+format, v...)
	}
}

func (this *defaultLogger) Warnf(format string, v ...any) {
	if this.level <= LevelWarn {
		this.logger.Printf("WARNING|"+format, v...)
	}
}

func (this *defaultLogger) Infof(format string, v ...any) {
	if this.level <= LevelInfo {
		this.logger.Printf("INFO|"+format, v...)
	}
}

func (this *defaultLogger) Debugf(format string, v ...any) {
	if this.level <= LevelDebug {
		this.logger.Printf("DEBUG|"+format, v...)
	}
}

type fileLogger struct {
	logger *log.Logger
	level  LogLevel
}

func NewFileLogger(level LogLevel, file string) *fileLogger {
	log := &fileLogger{
		logger: log.Default(),
		level:  level,
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	log.logger.SetOutput(f)
	return log
}

func (this *fileLogger) SetLogLevel(level LogLevel) {
	if level < LevelDebug {
		level = LevelDebug
	}
	if level > LevelError {
		level = LevelError
	}
	this.level = level
}
func (this *fileLogger) Errorf(format string, v ...any) {
	if this.level <= LevelError {
		this.logger.Printf("ERROR|"+format, v...)
	}
}

func (this *fileLogger) Warnf(format string, v ...any) {
	if this.level <= LevelWarn {
		this.logger.Printf("WARNING|"+format, v...)
	}
}

func (this *fileLogger) Infof(format string, v ...any) {
	if this.level <= LevelInfo {
		this.logger.Printf("INFO|"+format, v...)
	}
}

func (this *fileLogger) Debugf(format string, v ...any) {
	if this.level <= LevelDebug {
		this.logger.Printf("DEBUG|"+format, v...)
	}
}
