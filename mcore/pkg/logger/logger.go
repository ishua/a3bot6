package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	INFO = iota
	DEBUG
)

var logLevel = INFO

var (
	infoLogger  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
)

func SetLogLevel(level int) {
	logLevel = level
}

func Info(msg string) {
	if logLevel >= INFO {
		infoLogger.Println(msg)
	}
}

func Debug(msg string) {
	if logLevel >= DEBUG {
		debugLogger.Println(msg)
	}
}

func Fatal(msg string) {
	infoLogger.Println(msg)
	os.Exit(1)
}

func Infof(format string, args ...interface{}) {
	if logLevel >= INFO {
		infoLogger.Println(fmt.Sprintf(format, args...))
	}
}

func Debugf(format string, args ...interface{}) {
	if logLevel >= DEBUG {
		debugLogger.Println(fmt.Sprintf(format, args...))
	}
}

func Fatalf(format string, args ...interface{}) {
	infoLogger.Println(fmt.Sprintf(format, args...))
	os.Exit(1)
}
